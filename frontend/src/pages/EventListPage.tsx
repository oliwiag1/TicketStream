import { useMemo } from "react";
import { SeatMap } from "../components/SeatMap";

export function EventListPage() {
  const events = useMemo(
    () => [
      { id: "evt-1", title: "Koncert testowy", startsAt: "2026-07-01T19:00:00Z" },
      { id: "evt-2", title: "Festiwal testowy", startsAt: "2026-07-15T18:00:00Z" }
    ],
    []
  );

  return (
    <section className="card">
      <h2>Nadchodzace wydarzenia</h2>
      <ul className="event-list">
        {events.map((eventItem) => (
          <li key={eventItem.id}>
            <strong>{eventItem.title}</strong>
            <div>{new Date(eventItem.startsAt).toLocaleString("pl-PL")}</div>
          </li>
        ))}
      </ul>
      <SeatMap eventId="evt-1" />
    </section>
  );
}
