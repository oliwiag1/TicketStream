type SeatMapProps = {
  eventId: string;
};

export function SeatMap({ eventId }: SeatMapProps) {
  return (
    <section>
      <h3>Mapa miejsc (stub)</h3>
      <p>Event: {eventId}</p>
      <div className="seat-grid" aria-label="seat-map">
        {Array.from({ length: 20 }).map((_, index) => (
          <button key={index} type="button" className="seat available">
            {index + 1}
          </button>
        ))}
      </div>
    </section>
  );
}
