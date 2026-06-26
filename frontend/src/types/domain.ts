export type SeatStatus = "available" | "locked" | "sold";

export type EventDTO = {
  id: string;
  title: string;
  startsAt: string;
};

export type EventAPIResponse = {
  id: string;
  title: string;
  starts_at: string;
};

export type EventsResponse = {
  items: EventAPIResponse[];
  meta?: {
    cached?: boolean;
  };
};

export type SeatDTO = {
  id: string;
  row: string;
  number: number;
  status: SeatStatus;
};

export type SeatsResponse = {
  event_id: string;
  seats: SeatDTO[];
};

export type ReserveResponse = {
  reservation_id: string;
  event_id: string;
  seat_id: string;
  status: "locked";
  expires_at: string;
};

export type PayResponse = {
  status: "accepted";
  reservation_id: string;
  already_paid?: boolean;
};

export type CancelReservationResponse = {
  message: "released";
  reservation_id: string;
};

export type SeatUpdateEvent = {
  type: "seat_status_changed";
  event_id: string;
  seat_id: string;
  status: SeatStatus;
  sequence_number: number;
  changed_at: string;
};

export type SeatSnapshotItem = {
  seat_id: string;
  row: string;
  number: number;
  status: SeatStatus;
  changed_at: string;
};

export type SeatSnapshotEvent = {
  type: "snapshot";
  event_id: string;
  sequence_number: number;
  seats: SeatSnapshotItem[];
};
