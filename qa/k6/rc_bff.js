import http from 'k6/http';
import { check, sleep, group } from 'k6';

/**
 * Prompt 57 — Release Candidate k6 against GitHub Actions ubuntu-latest.
 *
 * Infrastructure assumptions (do not treat as production SLA):
 * - 2 vCPU / 7 GB GHA runner
 * - in-memory BFF + catalog/cart/location (no Postgres pool, no OpenSearch)
 * - traffic is browse-heavy Quick Commerce: health, home, catalog list, cart create
 * - checkout/payment are exercised in rc-journeys.sh, not at peak VU here
 *
 * Staged profile:
 *   warm-up  10s @ 2 VU
 *   baseline 20s @ 4 VU
 *   peak     20s @ 8 VU
 *   stress   15s @ 10 VU
 *
 * 12 in-memory services share a 2 vCPU GHA runner; home fans out to location.
 * Higher VU counts produced ~36% HTTP failures in run 32905683949 and are not a fair SLO.
 */
export const options = {
  scenarios: {
    staged: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '10s', target: 2 },
        { duration: '20s', target: 4 },
        { duration: '20s', target: 8 },
        { duration: '15s', target: 10 },
        { duration: '5s', target: 0 },
      ],
      gracefulRampDown: '5s',
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
    http_req_duration: ['p(50)<800', 'p(95)<2000', 'p(99)<4000'],
    checks: ['rate>0.95'],
  },
};

const BASE = __ENV.BFF_BASE || 'http://127.0.0.1:8111';
const TENANT = __ENV.TENANT_ID || '11111111-1111-1111-1111-111111111111';

export default function () {
  const params = {
    headers: {
      'X-Tenant-Id': TENANT,
      'Content-Type': 'application/json',
      'X-Request-Id': `k6-${__VU}-${__ITER}`,
    },
    timeout: '10s',
  };
  group('auth_health', () => {
    const h = http.get(`${BASE}/health`, params);
    check(h, { 'health 200': (r) => r.status === 200 });
  });
  group('home', () => {
    const home = http.get(`${BASE}/v1/customer/home?lat=41.0&lng=29.0`, params);
    check(home, { 'home 200': (r) => r.status === 200 });
  });
  sleep(0.3);
}
