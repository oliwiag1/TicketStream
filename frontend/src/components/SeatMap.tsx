import { useEffect, useMemo, useRef, useState } from "react";
import type { Dispatch, SetStateAction } from "react";
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
import { connectEventSocket } from "../services/ws";
import type {
  EventDTO,
  PayResponse,
  ReserveResponse,
  SeatDTO,
  SeatSnapshotEvent,
  SeatUpdateEvent
} from "../types/domain";

type SeatMapProps = {
  eventItem: EventDTO | null;
};

export function SeatMap({ eventItem }: SeatMapProps) {
  const queryClient = useQueryClient();
  const [selectedSeat, setSelectedSeat] = useState<SeatDTO | null>(null);
  const [reservation, setReservation] = useState<ReserveResponse | null>(null);
  const [payment, setPayment] = useState<PayResponse | null>(null);
  const [message, setMessage] = useState<string | null>(null);
  const [isSocketConnected, setIsSocketConnected] = useState(false);
  const [socketWarning, setSocketWarning] = useState<string | null>(null);
  const lastSequenceNumberRef = useRef(0);
  const reconnectTimerRef = useRef<number | null>(null);

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
    setIsSocketConnected(false);
    setSocketWarning(null);
    lastSequenceNumberRef.current = 0;
  }, [eventId]);

  useEffect(() => {
    if (!eventId) {
      return;
    }

    let closedByApp = false;
    let socket: WebSocket | null = null;

    const openSocket = () => {
      if (closedByApp) {
        return;
      }

      socket = connectEventSocket(eventId, {
        onOpen: () => {
          setIsSocketConnected(true);
          setSocketWarning(null);
        },
        onClose: () => {
          setIsSocketConnected(false);
          if (closedByApp) {
            return;
          }
          setSocketWarning("Połączenie realtime zostało zerwane. Trwa ponowne łączenie.");
          reconnectTimerRef.current = window.setTimeout(openSocket, 1000);
        },
        onError: () => {
          if (!closedByApp) {
            setSocketWarning("Realtime chwilowo niedostępny. Widok może wymagać odświeżenia.");
          }
        },
        onMessage: (payload) => {
          if (payload.type === "snapshot") {
            applySnapshot(queryClient, eventId, payload);
            reconcileLocalStateFromSnapshot(
              payload,
              setSelectedSeat,
              setReservation,
              setPayment,
              setMessage
            );
            lastSequenceNumberRef.current = payload.sequence_number;
            return;
          }

          const update = payload as SeatUpdateEvent;
          if (update.sequence_number <= lastSequenceNumberRef.current) {
            return;
          }

          if (update.sequence_number > lastSequenceNumberRef.current + 1) {
            void invalidateSeats(queryClient, eventId);
            setSocketWarning("Wykryto lukę w aktualizacjach. Odświeżamy stan miejsc.");
          }

          lastSequenceNumberRef.current = update.sequence_number;
          applySeatUpdate(queryClient, eventId, update);
          reconcileLocalStateFromUpdate(
            update,
            setSelectedSeat,
            setReservation,
            setPayment,
            setMessage
          );
        }
      });
    };

    openSocket();

    return () => {
      closedByApp = true;
      setIsSocketConnected(false);
      setSocketWarning(null);
      if (reconnectTimerRef.current !== null) {
        window.clearTimeout(reconnectTimerRef.current);
        reconnectTimerRef.current = null;
      }
      socket?.close();
    };
  }, [eventId, queryClient]);

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
          <div className="mb-3 flex items-center justify-between gap-3 rounded-md border border-line bg-paper px-3 py-2 text-sm text-rail">
            <span>
              Realtime: {isSocketConnected ? "połączono" : "łączenie / offline"}
            </span>
            {socketWarning ? <span className="text-amber">{socketWarning}</span> : null}
          </div>
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

function applySnapshot(
  queryClient: ReturnType<typeof useQueryClient>,
  eventId: string,
  snapshot: SeatSnapshotEvent
) {
  queryClient.setQueryData<SeatDTO[]>(["event-seats", eventId], (currentSeats = []) => {
    const currentById = new Map(currentSeats.map((seat) => [seat.id, seat]));
    return snapshot.seats.map((seat) => {
      const current = currentById.get(seat.seat_id);
      return {
        id: seat.seat_id,
        row: current?.row ?? seat.row,
        number: current?.number ?? seat.number,
        status: seat.status
      };
    });
  });
}

function applySeatUpdate(
  queryClient: ReturnType<typeof useQueryClient>,
  eventId: string,
  update: SeatUpdateEvent
) {
  queryClient.setQueryData<SeatDTO[]>(["event-seats", eventId], (currentSeats = []) =>
    currentSeats.map((seat) =>
      seat.id === update.seat_id ? { ...seat, status: update.status } : seat
    )
  );
}

function reconcileLocalStateFromSnapshot(
  snapshot: SeatSnapshotEvent,
  setSelectedSeat: Dispatch<SetStateAction<SeatDTO | null>>,
  setReservation: Dispatch<SetStateAction<ReserveResponse | null>>,
  setPayment: Dispatch<SetStateAction<PayResponse | null>>,
  setMessage: Dispatch<SetStateAction<string | null>>
) {
  const seatStatuses = new Map(snapshot.seats.map((seat) => [seat.seat_id, seat.status]));

  setSelectedSeat((current) => {
    if (!current) {
      return current;
    }
    const nextStatus = seatStatuses.get(current.id);
    return nextStatus ? { ...current, status: nextStatus } : current;
  });

  setReservation((current) => {
    if (!current) {
      return current;
    }
    const nextStatus = seatStatuses.get(current.seat_id);
    if (nextStatus === "available") {
      setPayment(null);
      setMessage("Rezerwacja nie jest już aktywna.");
      return null;
    }
    return current;
  });
}

function reconcileLocalStateFromUpdate(
  update: SeatUpdateEvent,
  setSelectedSeat: Dispatch<SetStateAction<SeatDTO | null>>,
  setReservation: Dispatch<SetStateAction<ReserveResponse | null>>,
  setPayment: Dispatch<SetStateAction<PayResponse | null>>,
  setMessage: Dispatch<SetStateAction<string | null>>
) {
  setSelectedSeat((current) =>
    current && current.id === update.seat_id ? { ...current, status: update.status } : current
  );

  setReservation((current) => {
    if (!current || current.seat_id !== update.seat_id) {
      return current;
    }

    if (update.status === "available") {
      setPayment(null);
      setMessage("Rezerwacja została zwolniona.");
      return null;
    }

    return current;
  });
}
