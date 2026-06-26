import { SeatButton } from "./SeatButton";
import type { PayResponse, ReserveResponse, SeatDTO } from "../../types/domain";

type SeatGridProps = {
  seatsByRow: Array<[string, SeatDTO[]]>;
  selectedSeat: SeatDTO | null;
  reservation: ReserveResponse | null;
  payment: PayResponse | null;
  isBusy: boolean;
  onSeatSelect: (seat: SeatDTO) => void;
};

export function SeatGrid({
  seatsByRow,
  selectedSeat,
  reservation,
  payment,
  isBusy,
  onSeatSelect
}: SeatGridProps) {
  if (seatsByRow.length === 0) {
    return null;
  }

  return (
    <div className="mt-4 space-y-3">
      {seatsByRow.map(([row, rowSeats]) => (
        <div key={row} className="grid grid-cols-[32px_minmax(0,1fr)] gap-2">
          <div className="flex h-10 items-center justify-center rounded-md bg-paper text-sm font-semibold text-rail">
            {row}
          </div>
          <div className="grid grid-cols-5 gap-2 sm:grid-cols-8 lg:grid-cols-10">
            {rowSeats.map((seat) => (
              <SeatButton
                key={seat.id}
                seat={seat}
                isSelected={selectedSeat?.id === seat.id}
                isReserved={reservation?.seat_id === seat.id}
                isBusy={isBusy}
                isCheckoutLocked={Boolean(reservation || payment)}
                isPaymentComplete={Boolean(payment)}
                onSelect={() => onSeatSelect(seat)}
              />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
