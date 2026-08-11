import http from 'k6/http';
import { check } from 'k6';

export const options = { vus: 20, duration: '30s' };

export default function () {
  const tenant = __ENV.TENANT_ID || '00000000-0000-4000-8000-000000000001';
  const res = http.get(`${__ENV.BASE_URL || 'http://localhost:8124'}/v1/hyperscale/gates`, {
    headers: { 'X-Tenant-Id': tenant },
  });
  check(res, { 'gates 200': (r) => r.status === 200 });
}
