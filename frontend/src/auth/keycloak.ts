import Keycloak from "keycloak-js";

const keycloak = new Keycloak({
  url: import.meta.env.VITE_KEYCLOAK_URL ?? "http://localhost:18081",
  realm: import.meta.env.VITE_KEYCLOAK_REALM ?? "ticketstream",
  clientId: import.meta.env.VITE_KEYCLOAK_CLIENT_ID ?? "ticketstream-frontend"
});

export async function initAuth(): Promise<boolean> {
  const authenticated = await keycloak.init({
    onLoad: "check-sso",
    pkceMethod: "S256",
    checkLoginIframe: false
  });

  return authenticated;
}

export async function login(): Promise<void> {
  await keycloak.login({
    redirectUri: window.location.origin
  });
}

export async function register(): Promise<void> {
  await keycloak.register({
    redirectUri: window.location.origin
  });
}

export async function getAccessToken(): Promise<string | null> {
  if (!keycloak.authenticated) {
    return null;
  }

  await keycloak.updateToken(30);
  return keycloak.token ?? null;
}

export async function logout(): Promise<void> {
  await keycloak.logout({
    redirectUri: window.location.origin
  });
}

export { keycloak };
