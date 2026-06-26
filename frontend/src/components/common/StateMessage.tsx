import { AlertCircle, Loader2 } from "lucide-react";

type StateMessageProps = {
  icon: "loading" | "error" | "empty";
  label: string;
  className?: string;
};

export function StateMessage({ icon, label, className = "" }: StateMessageProps) {
  return (
    <div
      className={[
        "flex items-center gap-2 rounded-md border border-line bg-paper px-3 py-3 text-sm text-rail",
        className
      ].join(" ")}
    >
      {icon === "loading" ? (
        <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
      ) : icon === "error" ? (
        <AlertCircle className="h-4 w-4 text-danger" aria-hidden="true" />
      ) : null}
      {label}
    </div>
  );
}
