import http from "k6/http";
import { check, sleep } from "k6";
import { SharedArray } from "k6/data";

const baseURL = __ENV.LOAD_BASE_URL || "http://localhost:8080";
const token = __ENV.LOAD_BEARER_TOKEN || "";
const eventID = __ENV.LOAD_EVENT_ID || "10000000-0000-0000-0000-000000000001";

http.setResponseCallback(http.expectedStatuses(200, 202, 401, 409));

const seats = new SharedArray("seat ids", () => {
  const fromEnv = (__ENV.LOAD_SEAT_IDS || "").trim();
  if (fromEnv.length === 0) {
    return ["20000000-0000-0000-0000-000000000001"];
  }
  return fromEnv.split(",").map((value) => value.trim()).filter(Boolean);
});

export const options = {
  scenarios: {
    flash_sale: {
      executor: "ramping-arrival-rate",
      startRate: 5,
      timeUnit: "1s",
      preAllocatedVUs: 30,
      maxVUs: 200,
      stages: [
        { target: 20, duration: "20s" },
        { target: 60, duration: "40s" },
        { target: 0, duration: "20s" }
      ]
    }
  },
  thresholds: {
    http_req_failed: ["rate<0.15"],
    http_req_duration: ["p(95)<1200", "p(99)<2500"]
  }
};

function headersWithAuth(extra = {}) {
  const headers = {
    "Content-Type": "application/json",
    ...extra
  };
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return headers;
}

export default function () {
  const seatID = seats[Math.floor(Math.random() * seats.length)];
  const reservePayload = JSON.stringify({ seat_id: seatID });

  const reserveRes = http.post(
    `${baseURL}/events/${eventID}/reserve`,
    reservePayload,
    { headers: headersWithAuth({ Origin: "http://localhost:3000" }) }
  );

  const reserveAccepted = check(reserveRes, {
    "reserve accepted or conflict": (r) => r.status === 202 || r.status === 409 || r.status === 401,
  });

  if (!reserveAccepted || reserveRes.status !== 202) {
    sleep(0.1);
    return;
  }

  const reserveBody = reserveRes.json();
  const reservationID = reserveBody && reserveBody.reservation_id;
  if (!reservationID) {
    sleep(0.1);
    return;
  }

  const paymentPayload = JSON.stringify({
    reservation_id: reservationID,
    payment_method: "card"
  });

  const payRes = http.post(`${baseURL}/pay`, paymentPayload, {
    headers: headersWithAuth({
      Origin: "http://localhost:3000",
      "Idempotency-Key": `${__VU}-${__ITER}-${Date.now()}`
    })
  });

  check(payRes, {
    "pay accepted or conflict": (r) => r.status === 200 || r.status === 409 || r.status === 401,
  });

  sleep(0.2);
}
