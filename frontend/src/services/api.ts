import axios from "axios";
import { getAccessToken } from "../auth/keycloak";
import type {
  EventDTO,
  EventsResponse,
  PayResponse,
  ReserveResponse,
  SeatDTO,
  SeatsResponse
} from "../types/domain";

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL ?? "http://localhost:8080",
  withCredentials: true
});

api.interceptors.request.use(async (requestConfig) => {
  const token = await getAccessToken();
  if (token) {
    requestConfig.headers = requestConfig.headers ?? {};
    requestConfig.headers.Authorization = `Bearer ${token}`;
  }
  return requestConfig;
});

export async function getEvents(): Promise<EventDTO[]> {
  const response = await api.get<EventsResponse>("/events");
  return response.data.items.map((eventItem) => ({
    id: eventItem.id,
    title: eventItem.title,
    startsAt: eventItem.starts_at
  }));
}

export async function getEventSeats(eventId: string): Promise<SeatDTO[]> {
  const response = await api.get<SeatsResponse>(`/events/${eventId}/seats`);
  return response.data.seats;
}

export async function reserveSeat(
  eventId: string,
  seatId: string
): Promise<ReserveResponse> {
  const response = await api.post<ReserveResponse>(`/events/${eventId}/reserve`, {
    seat_id: seatId
  });
  return response.data;
}

export async function payReservation(
  reservationId: string,
  idempotencyKey: string
): Promise<PayResponse> {
  const response = await api.post<PayResponse>(
    "/pay",
    {
      reservation_id: reservationId,
      payment_method: "card"
    },
    {
      headers: {
        "Idempotency-Key": idempotencyKey
      }
    }
  );
  return response.data;
}
