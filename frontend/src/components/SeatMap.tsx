import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import {
  AlertCircle,
  CheckCircle2,
  CreditCard,
  Loader2,
  RefreshCw,
  Ticket
} from "lucide-react";
import { getEventSeats, payReservation, reserveSeat } from "../services/api";
import type { EventDTO, PayResponse, ReserveResponse, SeatDTO } from "../types/domain";

type SeatMapProps = {
  eventItem: EventDTO | null;
};

export function SeatMap({ eventItem }: SeatMapProps) {
  const queryClient = useQueryClient();
  const [selectedSeat, setSelectedSeat] = useState<SeatDTO | null>(null);
  const [reservation, setReservation] = useState<ReserveResponse | null>(null);
  const [payment, setPayment] = useState<PayResponse | null>(null);
  const [message, setMessage] = useState<string | null>(null);

  const eventId = eventItem?.id ?? "";
  const seatsQuery = useQuery({
    queryKey: ["event-seats", eventId],
    queryFn: () => getEventSeats(eventId),
    enabled: Boolean(eventId)
  });

  useEffect(() => {
    setSelectedSeat(null);
    setReservation(null);
    setPayment(null);
    setMessage(null);
  }, [eventId]);

  const seats = seatsQuery.data ?? [];
  const seatsByRow = useMemo(() => groupSeatsByRow(seats), [seats]);
  const reservedSeat =
    seats.find((seat) => seat.id === reservation?.seat_id) ?? selectedSeat;

  const reserveMutation = useMutation({
    mutationFn: (seatId: string) => reserveSeat(eventId, seatId),
    onMutate: () => {
      setMessage(null);
      setReservation(null);
      setPayment(null);
    },
    onSuccess: (nextReservation) => {
      setReservation(nextReservation);
      setMessage("Miejsce zarezerwowane. Możesz przejść do płatności.");
      void queryClient.invalidateQueries({ queryKey: ["event-seats", eventId] });
    },
    onError: (error) => {
      if (isStatus(error, 409)) {
        setMessage("To miejsce jest już zajęte. Odświeżyliśmy mapę miejsc.");
        void queryClient.invalidateQueries({ queryKey: ["event-seats", eventId] });
        return;
      }
      setMessage("Nie udało się zarezerwować miejsca.");
    }
  });

  const payMutation = useMutation({
    mutationFn: (reservationId: string) =>
      payReservation(reservationId, createIdempotencyKey()),
    onMutate: () => {
      setMessage(null);
    },
    onSuccess: (nextPayment) => {
      setPayment(nextPayment);
      setMessage("Bilet kupiony.");
      void queryClient.invalidateQueries({ queryKey: ["event-seats", eventId] });
    },
    onError: (error) => {
      if (isStatus(error, 409)) {
        setMessage("Rezerwacja wygasła albo nie jest już dostępna.");
        void queryClient.invalidateQueries({ queryKey: ["event-seats", eventId] });
        return;
      }
      setMessage("Nie udało się przyjąć płatności.");
    }
  });

  if (!eventItem) {
    return (
      <section className="rounded-md border border-dashed border-line bg-white p-8 text-center text-rail">
        Wybierz wydarzenie, aby zobaczyć miejsca.
      </section>
    );
  }

  return (
    <section className="rounded-md border border-line bg-white p-4 shadow-panel">
      <div className="flex flex-col gap-4 xl:flex-row xl:items-start xl:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-mint">
            Mapa miejsc
          </p>
          <h2 className="mt-1 text-xl font-semibold text-ink">{eventItem.title}</h2>
          <p className="mt-1 text-sm text-rail">
            {new Date(eventItem.startsAt).toLocaleString("pl-PL")}
          </p>
        </div>
        <button
          type="button"
          onClick={() => void seatsQuery.refetch()}
          className="inline-flex items-center justify-center gap-2 rounded-md border border-line bg-white px-3 py-2 text-sm font-medium text-ink transition hover:border-ink focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2"
        >
          <RefreshCw className="h-4 w-4" aria-hidden="true" />
          Odśwież miejsca
        </button>
      </div>

      <div className="mt-5 grid gap-5 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div>
          <SeatLegend />

          {seatsQuery.isLoading ? (
            <StateMessage icon="loading" label="Ładowanie miejsc" />
          ) : null}
          {seatsQuery.isError ? (
            <StateMessage icon="error" label="Nie udało się pobrać miejsc." />
          ) : null}
          {!seatsQuery.isLoading && !seatsQuery.isError && seats.length === 0 ? (
            <StateMessage icon="empty" label="Brak miejsc dla tego wydarzenia." />
          ) : null}

          {seatsByRow.length > 0 ? (
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
                        isBusy={reserveMutation.isPending}
                        isCheckoutLocked={Boolean(reservation || payment)}
                        onSelect={() => {
                          setSelectedSeat(seat);
                          reserveMutation.mutate(seat.id);
                        }}
                      />
                    ))}
                  </div>
                </div>
              ))}
            </div>
          ) : null}
        </div>

        <aside className="rounded-md border border-line bg-paper p-4">
          <h3 className="text-base font-semibold text-ink">Rezerwacja</h3>
          {message ? (
            <div
              className={[
                "mt-3 flex gap-2 rounded-md border px-3 py-2 text-sm",
                payment
                  ? "border-mint/30 bg-mint/10 text-mint"
                  : "border-amber/30 bg-amber/10 text-amber"
              ].join(" ")}
            >
              {payment ? (
                <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
              ) : (
                <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" aria-hidden="true" />
              )}
              <span>{message}</span>
            </div>
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
            <div className="mt-5 rounded-md border border-mint/30 bg-white p-4">
              <div className="flex items-center gap-2 text-mint">
                <Ticket className="h-5 w-5" aria-hidden="true" />
                <span className="font-semibold">Bilet kupiony</span>
              </div>
              <p className="mt-2 text-sm text-rail">
                Potwierdzenie dla rezerwacji {payment.reservation_id}.
              </p>
            </div>
          ) : (
            <button
              type="button"
              disabled={!reservation || payMutation.isPending || reserveMutation.isPending}
              onClick={() => {
                if (reservation) {
                  payMutation.mutate(reservation.reservation_id);
                }
              }}
              className="mt-5 inline-flex w-full items-center justify-center gap-2 rounded-md bg-ink px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-mint focus:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-300"
            >
              {payMutation.isPending ? (
                <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
              ) : (
                <CreditCard className="h-4 w-4" aria-hidden="true" />
              )}
              Zapłać
            </button>
          )}
        </aside>
      </div>
    </section>
  );
}

function SeatButton({
  seat,
  isSelected,
  isReserved,
  isBusy,
  isCheckoutLocked,
  onSelect
}: {
  seat: SeatDTO;
  isSelected: boolean;
  isReserved: boolean;
  isBusy: boolean;
  isCheckoutLocked: boolean;
  onSelect: () => void;
}) {
  const isAvailable = seat.status === "available";
  const disabled = !isAvailable || isBusy || isCheckoutLocked;
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

function SeatLegend() {
  return (
    <div className="flex flex-wrap gap-2 text-xs text-rail">
      <LegendItem className="bg-mint/10 ring-mint/30" label="Dostępne" />
      <LegendItem className="bg-amber/10 ring-amber/30" label="Zarezerwowane" />
      <LegendItem className="bg-danger/10 ring-danger/30" label="Sprzedane" />
    </div>
  );
}

function LegendItem({ className, label }: { className: string; label: string }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-md border border-line bg-white px-2 py-1">
      <span className={`h-3 w-3 rounded-sm ring-1 ${className}`} />
      {label}
    </span>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.14em] text-rail">
        {label}
      </dt>
      <dd className="mt-1 break-words font-medium text-ink">{value}</dd>
    </div>
  );
}

function StateMessage({
  icon,
  label
}: {
  icon: "loading" | "error" | "empty";
  label: string;
}) {
  return (
    <div className="mt-4 flex items-center gap-2 rounded-md border border-line bg-paper px-3 py-3 text-sm text-rail">
      {icon === "loading" ? (
        <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />
      ) : icon === "error" ? (
        <AlertCircle className="h-4 w-4 text-danger" aria-hidden="true" />
      ) : null}
      {label}
    </div>
  );
}

function groupSeatsByRow(seats: SeatDTO[]): Array<[string, SeatDTO[]]> {
  const grouped = seats.reduce<Map<string, SeatDTO[]>>((acc, seat) => {
    const rowSeats = acc.get(seat.row) ?? [];
    rowSeats.push(seat);
    acc.set(seat.row, rowSeats);
    return acc;
  }, new Map());

  return Array.from(grouped.entries()).map(([row, rowSeats]) => [
    row,
    rowSeats.slice().sort((a, b) => a.number - b.number)
  ]);
}

function formatSeat(seat: SeatDTO | null): string {
  if (!seat) {
    return "nie wybrano";
  }
  return `${seat.row}${seat.number}`;
}

function createIdempotencyKey(): string {
  if ("randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function isStatus(error: unknown, status: number): boolean {
  return error instanceof AxiosError && error.response?.status === status;
}
