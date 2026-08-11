import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 20,
  duration: '10m',
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(99)<1000'],
  },
};

const BASE = __ENV.BFF_BASE || 'http://localhost:8111';

export default function () {
  const res = http.get(`${BASE}/health`);
  check(res, { ok: (r) => r.status === 200 });
  sleep(1);
}
