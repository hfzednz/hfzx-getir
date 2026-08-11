import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  scenarios: {
    load: { executor: 'ramping-vus', startVUs: 0, stages: [
      { duration: '1m', target: 50 },
      { duration: '3m', target: 50 },
      { duration: '1m', target: 0 },
    ]},
    spike: { executor: 'ramping-vus', startTime: '6m', startVUs: 0, stages: [
      { duration: '30s', target: 200 },
      { duration: '1m', target: 200 },
      { duration: '30s', target: 0 },
    ]},
  },
  thresholds: {
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<500'],
  },
};

const BASE = __ENV.CHECKOUT_BASE || 'http://localhost:8087';

export default function () {
  const res = http.get(`${BASE}/health`);
  check(res, { 'health 200': (r) => r.status === 200 });
  sleep(0.5);
}
