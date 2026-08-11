import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    orders: { executor: 'constant-arrival-rate', rate: 100, timeUnit: '1s', duration: '30s', preAllocatedVUs: 50 },
  },
};

export default function () {
  // Placeholder against BFF/order path — configure BASE_URL in CI.
  const res = http.get(`${__ENV.BASE_URL || 'http://localhost:8111'}/health`);
  check(res, { 'health ok': (r) => r.status === 200 });
}
