import { useQuery } from "@tanstack/react-query";
import { LogOut, ShieldCheck } from "lucide-react";
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
  const userQuery = useQuery({
    queryKey: ["current-user"],
    queryFn: async () => {
      const response = await api.get<CurrentUser>("/auth/me");
      return response.data;
    },
    retry: false
  });

  const user = userQuery.data;

  return (
    <div className="min-h-screen">
      <header className="border-b border-line bg-white/85 backdrop-blur">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-5 sm:px-6 lg:flex-row lg:items-center lg:justify-between lg:px-8">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-mint">
              TicketStream
            </p>
            <h1 className="mt-1 text-3xl font-semibold tracking-tight text-ink">
              Zakup biletu
            </h1>
          </div>
          {user ? (
            <div className="flex flex-wrap items-center gap-3 rounded-md border border-line bg-paper px-3 py-2 text-sm text-rail">
              <ShieldCheck className="h-4 w-4 text-mint" aria-hidden="true" />
              <span className="max-w-[260px] truncate">
                {user.preferred_username || user.email || user.sub}
              </span>
              <button
                type="button"
                onClick={() => void logout()}
                className="inline-flex items-center gap-2 rounded-md border border-line bg-white px-3 py-1.5 font-medium text-ink transition hover:border-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
              >
                <LogOut className="h-4 w-4" aria-hidden="true" />
                Wyloguj
              </button>
            </div>
          ) : null}
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-6 sm:px-6 lg:px-8">
        <EventListPage />
      </main>
    </div>
  );
}
