import { Ticket } from "lucide-react";
import type { PayResponse } from "../../types/domain";

type TicketConfirmationProps = {
  payment: PayResponse;
};

export function TicketConfirmation({ payment }: TicketConfirmationProps) {
  return (
    <div className="mt-5 rounded-md border border-mint/30 bg-white p-4">
      <div className="flex items-center gap-2 text-mint">
        <Ticket className="h-5 w-5" aria-hidden="true" />
        <span className="font-semibold">Bilet kupiony</span>
      </div>
      <p className="mt-2 text-sm text-rail">
        Potwierdzenie dla rezerwacji {payment.reservation_id}.
      </p>
    </div>
  );
}
