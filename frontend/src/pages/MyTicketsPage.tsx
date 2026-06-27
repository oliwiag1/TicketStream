import { useMemo } from "react";
import type { ReactNode } from "react";
import { useQuery } from "@tanstack/react-query";
import { CalendarClock, CreditCard, RefreshCw, Ticket } from "lucide-react";
import { StateMessage } from "../components/common/StateMessage";
import { getMyTickets } from "../services/api";

export function MyTicketsPage() {
  const ticketsQuery = useQuery({
    queryKey: ["my-tickets"],
    queryFn: getMyTickets
  });

  const tickets = ticketsQuery.data ?? [];
  const totals = useMemo(() => {
    const now = Date.now();
    const upcoming = tickets.filter((ticket) => Date.parse(ticket.eventStartsAt) >= now).length;
    return {
      all: tickets.length,
      upcoming
    };
  }, [tickets]);

  return (
    <section className="rounded-md border border-line bg-white p-4 shadow-panel">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-mint">
            Centrum użytkownika
          </p>
          <h2 className="mt-1 text-xl font-semibold text-ink">Moje bilety</h2>
          <p className="mt-1 text-sm text-rail">
            Wszystkie opłacone bilety przypisane do Twojego konta.
          </p>
        </div>

        <button
          type="button"
          onClick={() => void ticketsQuery.refetch()}
          className="inline-flex items-center justify-center gap-2 rounded-md border border-line bg-white px-3 py-2 text-sm font-medium text-ink transition hover:border-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
        >
          <RefreshCw className="h-4 w-4" aria-hidden="true" />
          Odśwież listę
        </button>
      </div>

      <div className="mt-4 grid gap-3 sm:grid-cols-2">
        <StatCard label="Łącznie biletów" value={String(totals.all)} />
        <StatCard label="Nadchodzące wydarzenia" value={String(totals.upcoming)} />
      </div>

      {ticketsQuery.isLoading ? (
        <StateMessage icon="loading" label="Ładowanie biletów" className="mt-5" />
      ) : null}
      {ticketsQuery.isError ? (
        <StateMessage
          icon="error"
          label="Nie udało się pobrać biletów. Spróbuj ponownie."
          className="mt-5"
        />
      ) : null}
      {!ticketsQuery.isLoading && !ticketsQuery.isError && tickets.length === 0 ? (
        <StateMessage
          icon="empty"
          label="Nie masz jeszcze kupionych biletów."
          className="mt-5"
        />
      ) : null}

      {!ticketsQuery.isLoading && !ticketsQuery.isError && tickets.length > 0 ? (
        <div className="mt-5 grid gap-4 lg:grid-cols-2">
          {tickets.map((ticketItem) => (
            <article
              key={ticketItem.reservationId}
              className="rounded-md border border-line bg-paper p-4"
            >
              <div className="flex items-start justify-between gap-3">
                <div>
                  <p className="text-xs font-semibold uppercase tracking-[0.16em] text-mint">
                    Bilet opłacony
                  </p>
                  <h3 className="mt-1 text-lg font-semibold text-ink">{ticketItem.eventTitle}</h3>
                </div>
                <span className="inline-flex items-center gap-1 rounded-md border border-mint/30 bg-mint/10 px-2 py-1 text-xs font-semibold text-mint">
                  <CreditCard className="h-3.5 w-3.5" aria-hidden="true" />
                  {ticketItem.status}
                </span>
              </div>

              <dl className="mt-4 grid gap-2 text-sm text-rail">
                <TicketInfoRow
                  icon={<Ticket className="h-4 w-4" aria-hidden="true" />}
                  label="Miejsce"
                  value={`${ticketItem.seatRow}${ticketItem.seatNumber}`}
                />
                <TicketInfoRow
                  icon={<CalendarClock className="h-4 w-4" aria-hidden="true" />}
                  label="Start wydarzenia"
                  value={new Date(ticketItem.eventStartsAt).toLocaleString("pl-PL")}
                />
                <TicketInfoRow
                  icon={<CalendarClock className="h-4 w-4" aria-hidden="true" />}
                  label="Zakupiono"
                  value={new Date(ticketItem.purchasedAt).toLocaleString("pl-PL")}
                />
              </dl>

              <div className="mt-4 rounded-md border border-line bg-white px-3 py-2 text-xs text-rail">
                <p className="font-semibold uppercase tracking-[0.14em] text-ink">Dane biletu</p>
                <p className="mt-1 break-all">Nr rezerwacji: {ticketItem.reservationId}</p>
                <p className="mt-1 break-all">ID wydarzenia: {ticketItem.eventId}</p>
              </div>
            </article>
          ))}
        </div>
      ) : null}
    </section>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-md border border-line bg-paper px-3 py-3">
      <p className="text-xs font-semibold uppercase tracking-[0.14em] text-rail">{label}</p>
      <p className="mt-1 text-xl font-semibold text-ink">{value}</p>
    </div>
  );
}

function TicketInfoRow({
  icon,
  label,
  value
}: {
  icon: ReactNode;
  label: string;
  value: string;
}) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-md border border-line bg-white px-3 py-2">
      <dt className="inline-flex items-center gap-2 text-rail">
        {icon}
        {label}
      </dt>
      <dd className="font-semibold text-ink">{value}</dd>
    </div>
  );
}