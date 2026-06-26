import { SeatLegendItem } from "./SeatLegendItem";

export function SeatLegend() {
  return (
    <div className="flex flex-wrap gap-2 text-xs text-rail">
      <SeatLegendItem className="bg-mint/10 ring-mint/30" label="Dostępne" />
      <SeatLegendItem className="bg-amber/10 ring-amber/30" label="Zarezerwowane" />
      <SeatLegendItem className="bg-danger/10 ring-danger/30" label="Sprzedane" />
    </div>
  );
}
