#!/usr/bin/env bash
# Deploy disposable phone-test staging stack (in-memory services, MockPSP, OTP_DEV_MODE).
# Run on a Docker host with a public IP. Pair with HTTPS reverse proxy (Caddy) when STAGING_DOMAIN is set.
#
# Usage:
#   export STAGING_DOMAIN=api-staging.example.com   # optional; enables Caddy TLS
#   export STAGING_PUBLIC_URL=https://api-staging.example.com/v1  # for APK build / docs
#   bash scripts/staging/deploy-phone-test.sh
#
# Do NOT use production credentials. Data is ephemeral (empty DATABASE_URL dev mode).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
NET="nexora-phone-staging"
TENANT="11111111-1111-1111-1111-111111111111"

SERVICES=(
  identity-service
  catalog-service
  cart-service
  location-service
  checkout-service
  payment-service
  order-service
  inventory-service
  finance-ledger-service
  settlement-service
  bff-customer
)

build_one() {
  local s="$1"
  local tmp
  tmp="$(mktemp)"
  sed -E -e 's|golang:1\.22-alpine|golang:1.26-alpine|g' -e 's|golang:1\.23-alpine|golang:1.26-alpine|g' \
    "$ROOT/services/$s/Dockerfile" >"$tmp"
  docker build -f "$tmp" -t "nexora/${s}:phone-staging" "$ROOT/services/$s"
  rm -f "$tmp"
}

run_svc() {
  local s="$1"
  shift
  docker rm -f "nexora-staging-$s" >/dev/null 2>&1 || true
  docker run -d --name "nexora-staging-$s" --network "$NET" --network-alias "$s" \
    -e HTTP_ADDR=:8080 \
    -e DATABASE_URL= \
    -e REDIS_URL= \
    -e KAFKA_BROKERS= \
    -e OTP_DEV_MODE=true \
    -e RATE_LIMIT_PER_MINUTE=0 \
    "$@" \
    "nexora/${s}:phone-staging" >/dev/null
}

wait_health() {
  local name="$1"
  for _ in $(seq 1 60); do
    local ip
    ip="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$name" 2>/dev/null || true)"
    if [[ -n "$ip" ]] && curl -fsS --max-time 2 "http://${ip}:8080/health" >/dev/null 2>&1; then
      echo "OK health $name"
      return 0
    fi
    sleep 2
  done
  echo "FAIL health $name"
  docker logs "$name" 2>&1 | tail -n 40 || true
  return 1
}

echo "==> network"
docker network inspect "$NET" >/dev/null 2>&1 || docker network create "$NET" >/dev/null

echo "==> build images"
for s in "${SERVICES[@]}"; do
  echo "    $s"
  build_one "$s"
done

echo "==> run services (in-memory dev mode, MockPSP, OTP_DEV_MODE=true)"
run_svc identity-service
run_svc catalog-service
run_svc cart-service
run_svc location-service
run_svc checkout-service
run_svc payment-service
run_svc order-service
run_svc inventory-service
run_svc finance-ledger-service
run_svc settlement-service
run_svc bff-customer -p 127.0.0.1:8111:8080 \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e CATALOG_URL=http://catalog-service:8080 \
  -e CART_URL=http://cart-service:8080 \
  -e LOCATION_URL=http://location-service:8080 \
  -e CHECKOUT_URL=http://checkout-service:8080 \
  -e PAYMENT_URL=http://payment-service:8080 \
  -e ORDER_URL=http://order-service:8080

for s in "${SERVICES[@]}"; do
  wait_health "nexora-staging-$s"
done

echo "==> smoke"
curl -fsS "http://127.0.0.1:8111/health" >/dev/null
curl -fsS -H "X-Tenant-Id: ${TENANT}" \
  "http://127.0.0.1:8111/v1/customer/home?lat=41.0&lng=29.0" >/dev/null
echo "OK smoke"

if [[ -n "${STAGING_DOMAIN:-}" ]]; then
  echo "==> Caddy reverse proxy (HTTPS)"
  docker rm -f nexora-staging-caddy >/dev/null 2>&1 || true
  cat > /tmp/nexora-staging-Caddyfile <<EOF
${STAGING_DOMAIN} {
  reverse_proxy host.docker.internal:8111
}
EOF
  docker run -d --name nexora-staging-caddy --network host \
    -v /tmp/nexora-staging-Caddyfile:/etc/caddy/Caddyfile:ro \
    caddy:2-alpine >/dev/null
  echo "OK Caddy listening on https://${STAGING_DOMAIN}"
  echo "    Point DNS A/AAAA record to this host before opening the app."
else
  echo "WARN: STAGING_DOMAIN unset — BFF bound to http://127.0.0.1:8111 only."
  echo "      Set STAGING_DOMAIN and re-run, or use a tunnel (cloudflared/ngrok) to port 8111."
fi

cat <<EOF

PHONE TEST STAGING READY (disposable in-memory stack)

Local BFF:     http://127.0.0.1:8111
Public URL:    ${STAGING_PUBLIC_URL:-<set STAGING_PUBLIC_URL after HTTPS is live>}
Tenant header: X-Tenant-Id: ${TENANT}
Test phone:    +905551112233
OTP:           docker logs nexora-staging-identity-service | grep otp.dev_mode
Payment:       MockPSP sandbox (token tok_ok in API tests)

Build Android APK:
  NEXORA_STAGING_BASE_URL="${STAGING_PUBLIC_URL:-https://YOUR-DOMAIN/v1}" bash scripts/ci/android-staging-customer.sh

Stop stack:
  for s in ${SERVICES[*]}; do docker rm -f nexora-staging-\$s; done
  docker network rm ${NET} 2>/dev/null || true

EOF
