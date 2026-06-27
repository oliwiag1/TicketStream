import { describe, expect, it } from "vitest";
import { createIdempotencyKey, groupSeatsByRow } from "./seatUtils";

describe("groupSeatsByRow", () => {
  it("groups seats by row and sorts by seat number", () => {
    const grouped = groupSeatsByRow([
      { id: "s3", row: "A", number: 3, status: "available" },
      { id: "s1", row: "A", number: 1, status: "locked" },
      { id: "s2", row: "B", number: 2, status: "sold" },
      { id: "s4", row: "A", number: 2, status: "available" }
    ]);

    expect(grouped).toEqual([
      [
        "A",
        [
          { id: "s1", row: "A", number: 1, status: "locked" },
          { id: "s4", row: "A", number: 2, status: "available" },
          { id: "s3", row: "A", number: 3, status: "available" }
        ]
      ],
      [["B"][0], [{ id: "s2", row: "B", number: 2, status: "sold" }]]
    ]);
  });
});

describe("createIdempotencyKey", () => {
  it("returns a non-empty key", () => {
    const key = createIdempotencyKey();
    expect(typeof key).toBe("string");
    expect(key.length).toBeGreaterThan(0);
  });
});
