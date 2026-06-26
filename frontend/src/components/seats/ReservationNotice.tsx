import { AlertCircle, CheckCircle2 } from "lucide-react";

type ReservationNoticeProps = {
  message: string;
  isSuccess: boolean;
};

export function ReservationNotice({ message, isSuccess }: ReservationNoticeProps) {
  return (
    <div
      className={[
        "mt-3 flex gap-2 rounded-md border px-3 py-2 text-sm",
        isSuccess
          ? "border-mint/30 bg-mint/10 text-mint"
          : "border-amber/30 bg-amber/10 text-amber"
      ].join(" ")}
    >
      {isSuccess ? (
        <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
      ) : (
        <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
      )}
      <span>{message}</span>
    </div>
  );
}
