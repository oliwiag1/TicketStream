type InfoRowProps = {
  label: string;
  value: string;
};

export function InfoRow({ label, value }: InfoRowProps) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.14em] text-rail">
        {label}
      </dt>
      <dd className="mt-1 break-words font-medium text-ink">{value}</dd>
    </div>
  );
}
