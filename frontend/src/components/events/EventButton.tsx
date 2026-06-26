import { CalendarDays } from "lucide-react";
import type { EventDTO } from "../../types/domain";

type EventButtonProps = {
  eventItem: EventDTO;
  isSelected: boolean;
  onSelect: () => void;
};

export function EventButton({ eventItem, isSelected, onSelect }: EventButtonProps) {
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
