import type { SeatDTO } from "../../types/domain";

type SeatButtonProps = {
  seat: SeatDTO;
  isSelected: boolean;
  isReserved: boolean;
  isBusy: boolean;
  isCheckoutLocked: boolean;
  isPaymentComplete: boolean;
  onSelect: () => void;
};

export function SeatButton({
  seat,
  isSelected,
  isReserved,
  isBusy,
  isCheckoutLocked,
  isPaymentComplete,
  onSelect
}: SeatButtonProps) {
  const isAvailable = seat.status === "available";
  const disabled =
    isPaymentComplete ||
    (!isAvailable && !isReserved) ||
    isBusy ||
    (isCheckoutLocked && !isReserved);
  const statusClass = isReserved
    ? "border-mint bg-mint text-white"
    : seat.status === "available"
      ? "border-mint/30 bg-mint/10 text-ink hover:border-mint"
      : seat.status === "locked"
        ? "border-amber/30 bg-amber/10 text-amber"
        : "border-danger/30 bg-danger/10 text-danger";

  return (
    <button
      type="button"
      disabled={disabled}
      onClick={onSelect}
      className={[
        "flex h-10 min-w-0 items-center justify-center rounded-md border text-sm font-semibold transition focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2 disabled:cursor-not-allowed",
        statusClass,
        isSelected && !isReserved ? "ring-2 ring-mint ring-offset-2" : ""
      ].join(" ")}
      aria-label={`Miejsce ${seat.row}${seat.number}, ${seat.status}`}
    >
      {seat.number}
    </button>
  );
}
