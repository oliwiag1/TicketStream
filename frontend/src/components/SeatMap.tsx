import { useEffect, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { RefreshCw } from "lucide-react";
import { StateMessage } from "./common/StateMessage";
import { ReservationPanel } from "./seats/ReservationPanel";
import { SeatGrid } from "./seats/SeatGrid";
import { SeatLegend } from "./seats/SeatLegend";
import {
  createIdempotencyKey,
  groupSeatsByRow,
  isStatus
} from "./seats/seatUtils";
import {
  cancelReservation,
  getEventSeats,
  payReservation,
  reserveSeat
} from "../services/api";
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
      setMessage("Miejsce zarezerwowane. Kliknij je ponownie, aby anulować.");
      void invalidateSeats(queryClient, eventId);
    },
    onError: (error) => {
      if (isStatus(error, 409)) {
        setMessage("To miejsce jest już zajęte. Odświeżyliśmy mapę miejsc.");
        void invalidateSeats(queryClient, eventId);
        return;
      }
      setMessage("Nie udało się zarezerwować miejsca.");
    }
  });

  const cancelMutation = useMutation({
    mutationFn: (reservationId: string) => cancelReservation(reservationId),
    onMutate: () => {
      setMessage(null);
    },
    onSuccess: () => {
      setSelectedSeat(null);
      setReservation(null);
      setPayment(null);
      setMessage("Rezerwacja została anulowana.");
      void invalidateSeats(queryClient, eventId);
    },
    onError: () => {
      setMessage("Nie udało się anulować rezerwacji.");
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
      void invalidateSeats(queryClient, eventId);
    },
    onError: (error) => {
      if (isStatus(error, 409)) {
        setMessage("Rezerwacja wygasła albo nie jest już dostępna.");
        void invalidateSeats(queryClient, eventId);
        return;
      }
      setMessage("Nie udało się przyjąć płatności.");
    }
  });

  const handleSeatSelect = (seat: SeatDTO) => {
    if (reservation?.seat_id === seat.id && !payment) {
      cancelMutation.mutate(reservation.reservation_id);
      return;
    }
    setSelectedSeat(seat);
    reserveMutation.mutate(seat.id);
  };

  const handlePay = () => {
    if (reservation) {
      payMutation.mutate(reservation.reservation_id);
    }
  };

  const handleCancel = () => {
    if (reservation) {
      cancelMutation.mutate(reservation.reservation_id);
    }
  };

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
            <StateMessage
              icon="loading"
              label="Ładowanie miejsc"
              className="mt-4"
            />
          ) : null}
          {seatsQuery.isError ? (
            <StateMessage
              icon="error"
              label="Nie udało się pobrać miejsc."
              className="mt-4"
            />
          ) : null}
          {!seatsQuery.isLoading && !seatsQuery.isError && seats.length === 0 ? (
            <StateMessage
              icon="empty"
              label="Brak miejsc dla tego wydarzenia."
              className="mt-4"
            />
          ) : null}

          <SeatGrid
            seatsByRow={seatsByRow}
            selectedSeat={selectedSeat}
            reservation={reservation}
            payment={payment}
            isBusy={reserveMutation.isPending || cancelMutation.isPending}
            onSeatSelect={handleSeatSelect}
          />
        </div>

        <ReservationPanel
          message={message}
          reservedSeat={reservedSeat}
          reservation={reservation}
          payment={payment}
          isPaying={payMutation.isPending}
          isReserving={reserveMutation.isPending}
          isCancelling={cancelMutation.isPending}
          onPay={handlePay}
          onCancel={handleCancel}
        />
      </div>
    </section>
  );
}

function invalidateSeats(queryClient: ReturnType<typeof useQueryClient>, eventId: string) {
  return queryClient.invalidateQueries({ queryKey: ["event-seats", eventId] });
}
