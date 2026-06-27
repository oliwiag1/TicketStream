package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

func CSRFProtect(allowOrigins []string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, raw := range allowOrigins {
		value := strings.TrimSpace(raw)
		if value == "" {
			continue
		}
		allowed[value] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if !isMutatingMethod(c.Request().Method) {
				return next(c)
			}

			origin := strings.TrimSpace(c.Request().Header.Get("Origin"))
			if origin != "" {
				if _, ok := allowed[origin]; !ok {
					return c.JSON(http.StatusForbidden, map[string]string{"error": "csrf_origin_not_allowed"})
				}
				return next(c)
			}

			referer := strings.TrimSpace(c.Request().Header.Get("Referer"))
			if referer != "" {
				u, err := url.Parse(referer)
				if err != nil {
					return c.JSON(http.StatusForbidden, map[string]string{"error": "csrf_invalid_referer"})
				}
				refererOrigin := u.Scheme + "://" + u.Host
				if _, ok := allowed[refererOrigin]; !ok {
					return c.JSON(http.StatusForbidden, map[string]string{"error": "csrf_referer_not_allowed"})
				}
			}

			return next(c)
		}
	}
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}
