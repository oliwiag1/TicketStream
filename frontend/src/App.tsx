import { useEffect, useState } from "react";
import { logout } from "./auth/keycloak";
import { EventListPage } from "./pages/EventListPage";
import { api } from "./services/api";

type CurrentUser = {
  sub: string;
  preferred_username: string;
  email: string;
  roles: string[];
  scope: string;
};

export function App() {
  const [user, setUser] = useState<CurrentUser | null>(null);

  useEffect(() => {
    const loadUser = async () => {
      try {
        const response = await api.get<CurrentUser>("/auth/me");
        setUser(response.data);
      } catch {
        setUser(null);
      }
    };

    void loadUser();
  }, []);

  return (
    <div className="app-shell">
      <header className="app-header">
        <h1>TicketStream</h1>
        <p>System rezerwacji biletow o wysokiej wspolbieznosci</p>
        {user ? (
          <div>
            <span>
              Zalogowany: {user.preferred_username || user.email || user.sub}
            </span>
            <button type="button" onClick={() => void logout()}>
              Wyloguj
            </button>
          </div>
        ) : null}
      </header>
      <main>
        <EventListPage />
      </main>
    </div>
  );
}
