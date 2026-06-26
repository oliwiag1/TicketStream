import { AxiosError } from "axios";
import type { SeatDTO } from "../../types/domain";

export function groupSeatsByRow(seats: SeatDTO[]): Array<[string, SeatDTO[]]> {
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

export function createIdempotencyKey(): string {
  if ("randomUUID" in crypto) {
    return crypto.randomUUID();
  }
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export function isStatus(error: unknown, status: number): boolean {
  return error instanceof AxiosError && error.response?.status === status;
}
