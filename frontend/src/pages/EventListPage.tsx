import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { AlertCircle, CalendarDays, Loader2, RefreshCw } from "lucide-react";
import { SeatMap } from "../components/SeatMap";
import { getEvents } from "../services/api";
import type { EventDTO } from "../types/domain";

export function EventListPage() {
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);
  const eventsQuery = useQuery({
    queryKey: ["events"],
    queryFn: getEvents
  });

  const events = eventsQuery.data ?? [];

  useEffect(() => {
    if (events.length === 0) {
      setSelectedEventId(null);
      return;
    }

    if (!selectedEventId || !events.some((eventItem) => eventItem.id === selectedEventId)) {
      setSelectedEventId(events[0].id);
    }
  }, [events, selectedEventId]);

  const selectedEvent =
    events.find((eventItem) => eventItem.id === selectedEventId) ?? null;

  return (
    <div className="grid gap-5 lg:grid-cols-[360px_minmax(0,1fr)]">
      <section className="rounded-md border border-line bg-white p-4 shadow-panel">
        <div className="flex items-start justify-between gap-3">
          <div>
            <h2 className="text-lg font-semibold text-ink">Wydarzenia</h2>
            <p className="mt-1 text-sm text-rail">
              Wybierz termin, a potem miejsce na sali.
            </p>
          </div>
          <button
            type="button"
            onClick={() => void eventsQuery.refetch()}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-line bg-white text-rail transition hover:border-ink hover:text-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
            aria-label="Odśwież wydarzenia"
          >
            <RefreshCw className="h-4 w-4" aria-hidden="true" />
          </button>
        </div>

        <div className="mt-4">
          {eventsQuery.isLoading ? <LoadingMessage label="Ładowanie wydarzeń" /> : null}
          {eventsQuery.isError ? (
            <ErrorMessage label="Nie udało się pobrać wydarzeń." />
          ) : null}
          {!eventsQuery.isLoading && !eventsQuery.isError && events.length === 0 ? (
            <EmptyMessage label="Brak dostępnych wydarzeń." />
          ) : null}
          {events.length > 0 ? (
            <ul className="space-y-2">
              {events.map((eventItem) => (
                <li key={eventItem.id}>
                  <EventButton
                    eventItem={eventItem}
                    isSelected={eventItem.id === selectedEventId}
                    onSelect={() => setSelectedEventId(eventItem.id)}
                  />
                </li>
              ))}
            </ul>
          ) : null}
        </div>
      </section>

      <SeatMap eventItem={selectedEvent} />
    </div>
  );
}

function EventButton({
  eventItem,
  isSelected,
  onSelect
}: {
  eventItem: EventDTO;
  isSelected: boolean;
  onSelect: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onSelect}
      className={[
        "w-full rounded-md border p-3 text-left transition focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2",
        isSelected
          ? "border-mint bg-mint/10 text-ink"
          : "border-line bg-white text-ink hover:border-rail"
      ].join(" ")}
    >
      <span className="block font-semibold">{eventItem.title}</span>
      <span className="mt-2 flex items-center gap-2 text-sm text-rail">
        <CalendarDays className="h-4 w-4" aria-hidden="true" />
        {new Date(eventItem.startsAt).toLocaleString("pl-PL")}
      </span>
    </button>
  );
}

function LoadingMessage({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-2 rounded-md border border-line bg-paper px-3 py-2 text-sm text-rail">
      <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
      {label}
    </div>
  );
}

function ErrorMessage({ label }: { label: string }) {
  return (
    <div className="flex items-center gap-2 rounded-md border border-danger/30 bg-danger/10 px-3 py-2 text-sm text-danger">
      <AlertCircle className="h-4 w-4" aria-hidden="true" />
      {label}
    </div>
  );
}

function EmptyMessage({ label }: { label: string }) {
  return (
    <div className="rounded-md border border-dashed border-line px-3 py-6 text-center text-sm text-rail">
      {label}
    </div>
  );
}
