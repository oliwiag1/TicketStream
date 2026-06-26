import type { SeatSnapshotEvent, SeatUpdateEvent } from "../types/domain";

export type EventSocketMessage = SeatSnapshotEvent | SeatUpdateEvent;

type EventSocketHandlers = {
  onMessage: (message: EventSocketMessage) => void;
  onOpen?: () => void;
  onClose?: () => void;
  onError?: () => void;
};

export function buildEventSocketURL(eventId: string): string {
  const configuredBase = import.meta.env.VITE_WS_URL;
  if (configuredBase) {
    return `${configuredBase}/ws/events/${eventId}`;
  }

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  return `${protocol}//${window.location.hostname}:8080/ws/events/${eventId}`;
}

export function connectEventSocket(
  eventId: string,
  handlers: EventSocketHandlers
): WebSocket {
  const socket = new WebSocket(buildEventSocketURL(eventId));

  socket.addEventListener("open", () => {
    handlers.onOpen?.();
  });

  socket.addEventListener("message", (event) => {
    try {
      const payload = JSON.parse(String(event.data)) as EventSocketMessage;
      handlers.onMessage(payload);
    } catch {
      handlers.onError?.();
    }
  });

  socket.addEventListener("close", () => {
    handlers.onClose?.();
  });

  socket.addEventListener("error", () => {
    handlers.onError?.();
  });

  return socket;
}
