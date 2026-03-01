import http from "k6/http";
import { check, sleep } from "k6";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const REDIRECT_CODE = __ENV.REDIRECT_CODE;
const VUS = Number(__ENV.VUS || 20);
const DURATION = __ENV.DURATION || "30s";
const SLEEP_MS = Number(__ENV.SLEEP_MS || 0);

if (!REDIRECT_CODE) {
  throw new Error("REDIRECT_CODE is required. Example: REDIRECT_CODE=abc123");
}

export const options = {
  vus: VUS,
  duration: DURATION,
  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<500"],
    checks: ["rate>0.99"],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/${REDIRECT_CODE}`, {
    redirects: 0,
  });

  check(res, {
    "status is 302": (r) => r.status === 302,
  });

  if (SLEEP_MS > 0) {
    sleep(SLEEP_MS / 1000);
  }
}
