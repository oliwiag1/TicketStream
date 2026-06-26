import { LogIn, Ticket, UserPlus } from "lucide-react";
import { login, register } from "../auth/keycloak";

export function EntryPage() {
  return (
    <main className="min-h-screen">
      <section className="mx-auto grid min-h-screen max-w-7xl items-center gap-8 px-4 py-10 sm:px-6 lg:grid-cols-[minmax(0,1fr)_420px] lg:px-8">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-mint">
            TicketStream
          </p>
          <h1 className="mt-4 max-w-3xl text-4xl font-semibold tracking-tight text-ink sm:text-5xl">
            Rezerwuj bilety bez walki o to samo miejsce.
          </h1>
          <p className="mt-5 max-w-2xl text-base leading-7 text-rail">
            TicketStream pokazuje dostępność miejsc, blokuje wybrane miejsce na czas
            płatności i chroni przed podwójną sprzedażą przy dużym ruchu.
          </p>

          <div className="mt-8 flex flex-col gap-3 sm:flex-row">
            <button
              type="button"
              onClick={() => void login()}
              className="inline-flex items-center justify-center gap-2 rounded-md bg-ink px-5 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
            >
              <LogIn className="h-4 w-4" aria-hidden="true" />
              Zaloguj
            </button>
            <button
              type="button"
              onClick={() => void register()}
              className="inline-flex items-center justify-center gap-2 rounded-md border border-line bg-white px-5 py-3 text-sm font-semibold text-ink transition hover:border-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
            >
              <UserPlus className="h-4 w-4" aria-hidden="true" />
              Załóż konto
            </button>
          </div>
        </div>

        <div className="rounded-md border border-line bg-white p-5 shadow-panel">
          <div className="flex items-center gap-3 border-b border-line pb-4">
            <div className="flex h-11 w-11 items-center justify-center rounded-md bg-mint/10 text-mint">
              <Ticket className="h-5 w-5" aria-hidden="true" />
            </div>
            <div>
              <h2 className="font-semibold text-ink">Demo flow</h2>
              <p className="text-sm text-rail">Lista wydarzeń, miejsca, rezerwacja, płatność.</p>
            </div>
          </div>

          <dl className="mt-5 space-y-4 text-sm">
            <EntryStat label="Status miejsc" value="available / locked / sold" />
            <EntryStat label="Płatność" value="Idempotency-Key" />
            <EntryStat label="Sesja" value="Keycloak OIDC + PKCE" />
          </dl>
        </div>
      </section>
    </main>
  );
}

function EntryStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-line bg-paper px-3 py-3">
      <dt className="text-xs font-semibold uppercase tracking-[0.14em] text-rail">
        {label}
      </dt>
      <dd className="mt-1 font-medium text-ink">{value}</dd>
    </div>
  );
}
