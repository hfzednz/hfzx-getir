#!/usr/bin/env bash
# Prompt 79: rebuild identity + RBAC edges on an existing phone-test Codespace.
# Does NOT recreate catalog/cart/order/checkout (preserves catalog + in-memory orders).
# Identity IS recreated so OTP principals get real role bindings (dest-mode memory).
set -euo pipefail
ROOT="${ROOT:-/workspaces/hfzx-getir}"
cd "$ROOT"
NET="nexora-phone-staging"
SSE_SECRET="${SSE_TICKET_SECRET:-nexora-phone-staging-sse}"
PUB_TOKEN="${REALTIME_PUBLISH_TOKEN:-nexora-phone-staging-publish}"

build_one() {
  local s="$1"
  local tmp
  tmp="$(mktemp)"
  sed -E -e 's|golang:1\.22-alpine|golang:1.26-alpine|g' -e 's|golang:1\.23-alpine|golang:1.26-alpine|g' \
    "$ROOT/services/$s/Dockerfile" >"$tmp"
  docker build -f "$tmp" -t "nexora/${s}:phone-staging" "$ROOT/services/$s"
  rm -f "$tmp"
}

wait_health() {
  local alias="$1"
  for _ in $(seq 1 40); do
    if docker run --rm --network "$NET" curlimages/curl:8.5.0 \
      -fsS --max-time 2 "http://${alias}:8080/health" >/dev/null 2>&1; then
      echo "OK health $alias"
      return 0
    fi
    sleep 2
  done
  echo "FAIL health $alias"
  docker logs "nexora-staging-$alias" 2>&1 | tail -n 40 || true
  return 1
}

run_new() {
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
    -e IDENTITY_URL=http://identity-service:8080 \
    "$@" \
    "nexora/${s}:phone-staging" >/dev/null
}

echo "==> keep catalog/order/checkout (do not recreate)"
for s in catalog-service cart-service location-service payment-service \
         inventory-service settlement-service checkout-service order-service tracking-service; do
  docker start "nexora-staging-$s" >/dev/null 2>&1 || true
done

echo "==> build RBAC images"
for s in identity-service realtime-gateway finance-ledger-service supplier-service \
         platform-ops-service bff-customer bff-courier bff-warehouse bff-admin; do
  echo "    $s"
  build_one "$s"
done

echo "==> recreate identity + edges"
run_new identity-service -p 0.0.0.0:8081:8080 -e CORS_ALLOWED_ORIGINS="*"
run_new realtime-gateway -p 0.0.0.0:8115:8080 \
  -e SSE_TICKET_SECRET="$SSE_SECRET" \
  -e REALTIME_PUBLISH_TOKEN="$PUB_TOKEN"
run_new finance-ledger-service -p 0.0.0.0:8091:8080
run_new supplier-service -p 0.0.0.0:8117:8080
run_new platform-ops-service -p 0.0.0.0:8110:8080
run_new bff-customer -p 0.0.0.0:8111:8080 \
  -e CATALOG_URL=http://catalog-service:8080 \
  -e CART_URL=http://cart-service:8080 \
  -e LOCATION_URL=http://location-service:8080 \
  -e CHECKOUT_URL=http://checkout-service:8080 \
  -e PAYMENT_URL=http://payment-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e SSE_TICKET_SECRET="$SSE_SECRET"
run_new bff-courier -p 0.0.0.0:8112:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e REALTIME_URL=http://realtime-gateway:8080 \
  -e REALTIME_PUBLISH_TOKEN="$PUB_TOKEN"
run_new bff-warehouse -p 0.0.0.0:8113:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e REALTIME_URL=http://realtime-gateway:8080 \
  -e REALTIME_PUBLISH_TOKEN="$PUB_TOKEN"
run_new bff-admin -p 0.0.0.0:8114:8080 \
  -e ORDER_URL=http://order-service:8080

for s in identity-service realtime-gateway finance-ledger-service supplier-service \
         platform-ops-service bff-customer bff-courier bff-warehouse bff-admin tracking-service; do
  wait_health "$s"
done

echo "==> customer-web still up?"
curl -sS -o /dev/null -w "customer_login:%{http_code}\n" --max-time 5 http://127.0.0.1:3000/login || true
free -h | head -2
echo "PROMPT79_APPLY_DONE"
