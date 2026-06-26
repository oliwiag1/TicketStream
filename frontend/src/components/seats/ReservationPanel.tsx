import { InfoRow } from "./InfoRow";
import { ReservationActions } from "./ReservationActions";
import { ReservationNotice } from "./ReservationNotice";
import { TicketConfirmation } from "./TicketConfirmation";
import type { PayResponse, ReserveResponse, SeatDTO } from "../../types/domain";

type ReservationPanelProps = {
  message: string | null;
  reservedSeat: SeatDTO | null;
  reservation: ReserveResponse | null;
  payment: PayResponse | null;
  isPaying: boolean;
  isReserving: boolean;
  isCancelling: boolean;
  onPay: () => void;
  onCancel: () => void;
};

export function ReservationPanel({
  message,
  reservedSeat,
  reservation,
  payment,
  isPaying,
  isReserving,
  isCancelling,
  onPay,
  onCancel
}: ReservationPanelProps) {
  return (
    <aside className="rounded-md border border-line bg-paper p-4">
      <h3 className="text-base font-semibold text-ink">Rezerwacja</h3>
      {message ? (
        <ReservationNotice message={message} isSuccess={Boolean(payment)} />
      ) : null}

      <dl className="mt-4 space-y-3 text-sm">
        <InfoRow label="Miejsce" value={formatSeat(reservedSeat)} />
        <InfoRow
          label="Status"
          value={payment ? "opłacone" : reservation ? "zarezerwowane" : "brak"}
        />
        <InfoRow
          label="Rezerwacja"
          value={reservation?.reservation_id ?? "nie utworzono"}
        />
      </dl>

      {payment ? (
        <TicketConfirmation payment={payment} />
      ) : (
        <ReservationActions
          canAct={Boolean(reservation)}
          isPaying={isPaying}
          isReserving={isReserving}
          isCancelling={isCancelling}
          onPay={onPay}
          onCancel={onCancel}
        />
      )}
    </aside>
  );
}

function formatSeat(seat: SeatDTO | null): string {
  if (!seat) {
    return "nie wybrano";
  }
  return `${seat.row}${seat.number}`;
}
