export function buildEventSocketURL(eventId: string): string {
  const base = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080";
  return `${base}/ws/events/${eventId}`;
}
