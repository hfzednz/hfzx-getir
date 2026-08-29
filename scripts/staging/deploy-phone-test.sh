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
  tracking-service
  realtime-gateway
  supplier-service
  bff-customer
  bff-courier
  bff-warehouse
  bff-admin
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
  local alias="${name#nexora-staging-}"
  for _ in $(seq 1 60); do
    # Prefer network-alias curl so this works when the script runs via docker.sock helper.
    if docker run --rm --network "$NET" curlimages/curl:8.5.0 \
      -fsS --max-time 2 "http://${alias}:8080/health" >/dev/null 2>&1; then
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
# LAN bind: publish on 0.0.0.0 so phones on the same Wi-Fi can reach the BFF/identity.
# Override with PHONE_BIND_HOST=127.0.0.1 for loopback-only.
PHONE_BIND_HOST="${PHONE_BIND_HOST:-0.0.0.0}"

run_svc identity-service -p "${PHONE_BIND_HOST}:8081:8080" \
  -e CORS_ALLOWED_ORIGINS="${CORS_ALLOWED_ORIGINS:-*}"
run_svc catalog-service
run_svc cart-service
run_svc location-service
run_svc checkout-service \
  -e CART_URL=http://cart-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e PAYMENT_URL=http://payment-service:8080 \
  -e INVENTORY_URL=http://inventory-service:8080
run_svc payment-service
run_svc order-service
run_svc inventory-service
run_svc finance-ledger-service -p "${PHONE_BIND_HOST}:8091:8080" \
  -e IDENTITY_URL=http://identity-service:8080
run_svc settlement-service
run_svc tracking-service
run_svc realtime-gateway -p "${PHONE_BIND_HOST}:8115:8080" \
  -e SSE_TICKET_SECRET="${SSE_TICKET_SECRET:-nexora-phone-staging-sse}" \
  -e REALTIME_PUBLISH_TOKEN="${REALTIME_PUBLISH_TOKEN:-nexora-phone-staging-publish}"
run_svc supplier-service -p "${PHONE_BIND_HOST}:8117:8080" \
  -e IDENTITY_URL=http://identity-service:8080
run_svc bff-customer -p "${PHONE_BIND_HOST}:8111:8080" \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e CATALOG_URL=http://catalog-service:8080 \
  -e CART_URL=http://cart-service:8080 \
  -e LOCATION_URL=http://location-service:8080 \
  -e CHECKOUT_URL=http://checkout-service:8080 \
  -e PAYMENT_URL=http://payment-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e SSE_TICKET_SECRET="${SSE_TICKET_SECRET:-nexora-phone-staging-sse}"
run_svc bff-courier -p "${PHONE_BIND_HOST}:8112:8080" \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e REALTIME_URL=http://realtime-gateway:8080 \
  -e REALTIME_PUBLISH_TOKEN="${REALTIME_PUBLISH_TOKEN:-nexora-phone-staging-publish}"
run_svc bff-warehouse -p "${PHONE_BIND_HOST}:8113:8080" \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e REALTIME_URL=http://realtime-gateway:8080 \
  -e REALTIME_PUBLISH_TOKEN="${REALTIME_PUBLISH_TOKEN:-nexora-phone-staging-publish}"
run_svc bff-admin -p "${PHONE_BIND_HOST}:8114:8080" \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e ORDER_URL=http://order-service:8080

for s in "${SERVICES[@]}"; do
  wait_health "nexora-staging-$s"
done

echo "==> smoke"
curl -fsS "http://127.0.0.1:8111/health" >/dev/null
# The storefront feed requires a customer session, so an anonymous call must be
# rejected. A 200 here would mean the BFF stopped enforcing authentication.
home_status=$(curl -sS -o /dev/null -w '%{http_code}' -H "X-Tenant-Id: ${TENANT}" \
  "http://127.0.0.1:8111/v1/customer/home?lat=41.0&lng=29.0")
if [[ "$home_status" != "401" ]]; then
  echo "FAIL smoke: anonymous /v1/customer/home returned $home_status, expected 401" >&2
  exit 1
fi
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
  echo "LAN BFF:        http://${PHONE_BIND_HOST}:8111 (use PC LAN IP from the phone)"
  echo "LAN identity:   http://${PHONE_BIND_HOST}:8081"
fi

cat <<EOF

PHONE TEST STAGING READY (disposable in-memory stack)

Local BFF:     http://127.0.0.1:8111
LAN BFF:       http://<PC-LAN-IP>:8111  (published on ${PHONE_BIND_HOST}:8111)
LAN identity:  http://<PC-LAN-IP>:8081
Public URL:    ${STAGING_PUBLIC_URL:-<optional HTTPS via STAGING_DOMAIN>}
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
