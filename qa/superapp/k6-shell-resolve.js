import http from 'k6/http';
import { check } from 'k6';

export const options = { vus: 5, duration: '15s' };

export default function () {
  const tenant = __ENV.TENANT_ID || '00000000-0000-4000-8000-000000000001';
  const res = http.get(
    `${__ENV.BASE_URL || 'http://localhost:8121'}/v1/superapp/shell/resolve?subjectId=k6-user&shellVersion=1.0.0`,
    { headers: { 'X-Tenant-Id': tenant } },
  );
  check(res, { 'status 200 or 400': (r) => r.status === 200 || r.status === 400 });
}
