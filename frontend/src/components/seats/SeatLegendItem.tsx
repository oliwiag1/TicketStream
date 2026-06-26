type SeatLegendItemProps = {
  className: string;
  label: string;
};

export function SeatLegendItem({ className, label }: SeatLegendItemProps) {
  return (
    <span className="inline-flex items-center gap-2 rounded-md border border-line bg-white px-2 py-1">
      <span className={`h-3 w-3 rounded-sm ring-1 ${className}`} />
      {label}
    </span>
  );
}
