export type SeatStatus = "available" | "locked" | "sold";

export type EventDTO = {
  id: string;
  title: string;
  startsAt: string;
};

export type SeatUpdateEvent = {
  event_id: string;
  seat_id: string;
  status: SeatStatus;
  sequence_number: number;
  changed_at: string;
};
