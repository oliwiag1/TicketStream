package auth

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

type TokenClaims struct {
	jwt.RegisteredClaims
	Subject           string                           `json:"sub"`
	PreferredUsername string                           `json:"preferred_username"`
	Email             string                           `json:"email"`
	Scope             string                           `json:"scope"`
	AuthorizedParty   string                           `json:"azp"`
	RealmAccess       RealmAccessClaims                `json:"realm_access"`
	ResourceAccess    map[string]ResourceAccessClaims  `json:"resource_access"`
}

type RealmAccessClaims struct {
	Roles []string `json:"roles"`
}

type ResourceAccessClaims struct {
	Roles []string `json:"roles"`
}

type Validator struct {
	issuerURL string
	audience  string
	jwksURL   string

	mu      sync.RWMutex
	jwks    *keyfunc.JWKS
	initErr error
}

func NewValidator(issuerURL, audience, jwksURL string) *Validator {
	return &Validator{issuerURL: issuerURL, audience: audience, jwksURL: jwksURL}
}

func (v *Validator) Verify(ctx context.Context, rawToken string) (*TokenClaims, error) {
	if rawToken == "" {
		return nil, errors.New("missing token")
	}

	if err := v.ensureJWKS(ctx); err != nil {
		return nil, err
	}

	v.mu.RLock()
	jwks := v.jwks
	v.mu.RUnlock()

	if jwks == nil {
		return nil, errors.New("jwks verifier unavailable")
	}

	claims := &TokenClaims{}
	parsedToken, err := jwt.ParseWithClaims(
		rawToken,
		claims,
		jwks.Keyfunc,
		jwt.WithIssuer(v.issuerURL),
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
	)
	if err != nil {
		return nil, err
	}
	if !parsedToken.Valid {
		return nil, errors.New("token is not valid")
	}
	if !hasAudience(claims.Audience, v.audience) && claims.AuthorizedParty != v.audience {
		return nil, errors.New("token audience is not valid")
	}

	return claims, nil
}

func hasAudience(audiences jwt.ClaimStrings, expected string) bool {
	for _, audience := range audiences {
		if audience == expected {
			return true
		}
	}
	return false
}

func (v *Validator) ensureJWKS(ctx context.Context) error {
	v.mu.RLock()
	if v.jwks != nil {
		v.mu.RUnlock()
		return nil
	}
	if v.initErr != nil {
		err := v.initErr
		v.mu.RUnlock()
		return err
	}
	v.mu.RUnlock()

	v.mu.Lock()
	defer v.mu.Unlock()

	if v.jwks != nil {
		return nil
	}
	if v.initErr != nil {
		return v.initErr
	}

	options := keyfunc.Options{
		Ctx:               ctx,
		RefreshErrorHandler: func(err error) {},
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  time.Minute,
		RefreshTimeout:    10 * time.Second,
		RefreshUnknownKID: true,
	}

	jwks, err := keyfunc.Get(v.jwksURL, options)
	if err != nil {
		v.initErr = err
		return err
	}

	v.jwks = jwks
	return nil
}
