import { CreditCard, Loader2, RotateCcw } from "lucide-react";

type ReservationActionsProps = {
  canAct: boolean;
  isPaying: boolean;
  isReserving: boolean;
  isCancelling: boolean;
  onPay: () => void;
  onCancel: () => void;
};

export function ReservationActions({
  canAct,
  isPaying,
  isReserving,
  isCancelling,
  onPay,
  onCancel
}: ReservationActionsProps) {
  const disabled = !canAct || isPaying || isReserving || isCancelling;

  return (
    <div className="mt-5 space-y-2">
      <button
        type="button"
        disabled={disabled}
        onClick={onPay}
        className="inline-flex w-full items-center justify-center gap-2 rounded-md bg-ink px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-300"
      >
        {isPaying ? (
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
        ) : (
          <CreditCard className="h-4 w-4" aria-hidden="true" />
        )}
        Zapłać
      </button>

      <button
        type="button"
        disabled={disabled}
        onClick={onCancel}
        className="inline-flex w-full items-center justify-center gap-2 rounded-md border border-line bg-white px-4 py-3 text-sm font-semibold text-ink transition hover:border-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-rail"
      >
        {isCancelling ? (
          <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
        ) : (
          <RotateCcw className="h-4 w-4" aria-hidden="true" />
        )}
        Anuluj rezerwację
      </button>
    </div>
  );
}
