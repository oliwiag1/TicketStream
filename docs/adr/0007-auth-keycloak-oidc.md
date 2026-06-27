# ADR-0007: Autentykacja przez Keycloak (OIDC)

## Status

Zaakceptowano

| Pole | Treść |
| :---- | :---- |
| **Decyzja** | Logowanie realizujemy przez Keycloak, a warstwa kliencka używa przepływu Authorization Code + PKCE. Warstwa serwerowa weryfikuje tokeny JWT przez JWKS. |
| **Kontekst** | Wymagane jest rozróżnienie użytkownika zalogowanego i niezalogowanego oraz ochrona punktów końcowych mutujących. |
| **Alternatywy** | 1) Własny moduł auth i sesje cookie; 2) Stateless JWT bez zewnętrznego IdP; 3) SaaS IdP (Auth0/Okta). |
| **Uzasadnienie** | Keycloak daje gotowy i standardowy mechanizm OIDC, przyspiesza wdrożenie i ogranicza ryzyko błędów w obszarze bezpieczeństwa. |
| **Kompromisy** | 1) Dodatkowy komponent do utrzymania; 2) Zależność od dostępności IdP; 3) Potrzeba pilnowania konfiguracji realm/client w środowiskach. |




