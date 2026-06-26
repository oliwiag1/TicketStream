import { RefreshCw } from "lucide-react";
import { StateMessage } from "../common/StateMessage";
import { EventButton } from "./EventButton";
import type { EventDTO } from "../../types/domain";

type EventListPanelProps = {
  events: EventDTO[];
  selectedEventId: string | null;
  isLoading: boolean;
  isError: boolean;
  onRefresh: () => void;
  onSelectEvent: (eventId: string) => void;
};

export function EventListPanel({
  events,
  selectedEventId,
  isLoading,
  isError,
  onRefresh,
  onSelectEvent
}: EventListPanelProps) {
  return (
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
          onClick={onRefresh}
          className="inline-flex h-9 w-9 items-center justify-center rounded-md border border-line bg-white text-rail transition hover:border-ink hover:text-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
          aria-label="Odśwież wydarzenia"
        >
          <RefreshCw className="h-4 w-4" aria-hidden="true" />
        </button>
      </div>

      <div className="mt-4">
        {isLoading ? (
          <StateMessage icon="loading" label="Ładowanie wydarzeń" className="py-2" />
        ) : null}
        {isError ? (
          <StateMessage
            icon="error"
            label="Nie udało się pobrać wydarzeń."
            className="border-danger/30 bg-danger/10 py-2 text-danger"
          />
        ) : null}
        {!isLoading && !isError && events.length === 0 ? (
          <div className="rounded-md border border-dashed border-line px-3 py-6 text-center text-sm text-rail">
            Brak dostępnych wydarzeń.
          </div>
        ) : null}
        {events.length > 0 ? (
          <ul className="space-y-2">
            {events.map((eventItem) => (
              <li key={eventItem.id}>
                <EventButton
                  eventItem={eventItem}
                  isSelected={eventItem.id === selectedEventId}
                  onSelect={() => onSelectEvent(eventItem.id)}
                />
              </li>
            ))}
          </ul>
        ) : null}
      </div>
    </section>
  );
}
