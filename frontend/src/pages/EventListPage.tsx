import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { EventListPanel } from "../components/events/EventListPanel";
import { SeatMap } from "../components/SeatMap";
import { getEvents } from "../services/api";

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
      <EventListPanel
        events={events}
        selectedEventId={selectedEventId}
        isLoading={eventsQuery.isLoading}
        isError={eventsQuery.isError}
        onRefresh={() => void eventsQuery.refetch()}
        onSelectEvent={setSelectedEventId}
      />
      <SeatMap eventItem={selectedEvent} />
    </div>
  );
}
