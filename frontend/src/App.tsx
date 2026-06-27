import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { ListChecks, LogOut, ShieldCheck, Ticket } from "lucide-react";
import { logout } from "./auth/keycloak";
import { EntryPage } from "./pages/EntryPage";
import { EventListPage } from "./pages/EventListPage";
import { MyTicketsPage } from "./pages/MyTicketsPage";
import { api } from "./services/api";
import { StateMessage } from "./components/common/StateMessage";

type CurrentUser = {
  sub: string;
  preferred_username: string;
  email: string;
  roles: string[];
  scope: string;
};

type AppProps = {
  initialAuthenticated: boolean;
};

export function App({ initialAuthenticated }: AppProps) {
  const [isAuthenticated, setIsAuthenticated] = useState(initialAuthenticated);
  const [activeView, setActiveView] = useState<"events" | "tickets">("events");
  const userQuery = useQuery({
    queryKey: ["current-user"],
    queryFn: async () => {
      const response = await api.get<CurrentUser>("/auth/me");
      return response.data;
    },
    enabled: isAuthenticated,
    retry: false
  });

  const user = userQuery.data;

  useEffect(() => {
    if (isAuthenticated && userQuery.isError) {
      setIsAuthenticated(false);
    }
  }, [isAuthenticated, userQuery.isError]);

  if (!isAuthenticated) {
    return <EntryPage />;
  }

  if (userQuery.isLoading) {
    return (
      <main className="flex min-h-screen items-center justify-center px-4">
        <StateMessage icon="loading" label="Ładowanie sesji użytkownika" />
      </main>
    );
  }

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
                onClick={() => {
                  setIsAuthenticated(false);
                  void logout();
                }}
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
        <div className="mb-5 inline-flex w-full flex-wrap items-center gap-2 rounded-md border border-line bg-white p-2 sm:w-auto">
          <button
            type="button"
            onClick={() => setActiveView("events")}
            className={[
              "inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2",
              activeView === "events"
                ? "bg-ink text-white"
                : "border border-line bg-white text-ink hover:border-ink"
            ].join(" ")}
          >
            <Ticket className="h-4 w-4" aria-hidden="true" />
            Zakup biletu
          </button>
          <button
            type="button"
            onClick={() => setActiveView("tickets")}
            className={[
              "inline-flex items-center gap-2 rounded-md px-3 py-2 text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2",
              activeView === "tickets"
                ? "bg-ink text-white"
                : "border border-line bg-white text-ink hover:border-ink"
            ].join(" ")}
          >
            <ListChecks className="h-4 w-4" aria-hidden="true" />
            Moje bilety
          </button>
        </div>

        {activeView === "events" ? <EventListPage /> : <MyTicketsPage />}
      </main>
    </div>
  );
}
