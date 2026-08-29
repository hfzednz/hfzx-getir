#!/usr/bin/env bash
# Prompt 77: add tracking + realtime + role BFFs + role Next apps on an existing phone-test Codespace.
# Does NOT recreate identity/catalog/cart/order (preserves in-memory OTP users and orders).
set -euo pipefail
ROOT="${ROOT:-/workspaces/hfzx-getir}"
cd "$ROOT"
NET="nexora-phone-staging"
TENANT="11111111-1111-1111-1111-111111111111"
PUBLIC_BASE="https://ominous-sniffle-v66jrrr567vcwpvj"
CUSTOMER_URL="${PUBLIC_BASE}-3000.app.github.dev"

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
    "$@" \
    "nexora/${s}:phone-staging"
}

echo "==> source markers"
grep -q 'TRACKING_URL' services/bff-customer/internal/adapters/httpclients/config.go
grep -q 'DepsFromEnv' services/bff-warehouse/cmd/bff-warehouse/main.go
grep -q 'offers/{id}/complete' services/bff-courier/internal/adapters/http/handlers.go
echo "OK source markers"

echo "==> keep existing core containers"
for s in identity-service catalog-service cart-service location-service payment-service \
         inventory-service settlement-service checkout-service order-service; do
  docker start "nexora-staging-$s" >/dev/null 2>&1 || true
done

echo "==> build new images"
for s in tracking-service realtime-gateway supplier-service platform-ops-service \
         bff-customer bff-courier bff-warehouse bff-admin; do
  echo "    $s"
  build_one "$s"
done

docker network inspect "$NET" >/dev/null 2>&1 || docker network create "$NET" >/dev/null

echo "==> finance-ledger publish 8091 (same image, dest-mode wipe of empty journals is OK)"
FIN_IMG="$(docker inspect -f '{{.Config.Image}}' nexora-staging-finance-ledger-service 2>/dev/null || echo nexora/finance-ledger-service:phone-staging)"
docker rm -f nexora-staging-finance-ledger-service >/dev/null 2>&1 || true
docker run -d --name nexora-staging-finance-ledger-service --network "$NET" \
  --network-alias finance-ledger-service \
  -p 0.0.0.0:8091:8080 \
  -e HTTP_ADDR=:8080 -e DATABASE_URL= -e REDIS_URL= -e KAFKA_BROKERS= \
  -e OTP_DEV_MODE=true -e RATE_LIMIT_PER_MINUTE=0 \
  "$FIN_IMG" >/dev/null

echo "==> run tracking / realtime / supplier / platform-ops / role BFFs"
run_new tracking-service >/dev/null
run_new realtime-gateway -p 0.0.0.0:8115:8080 >/dev/null
run_new supplier-service -p 0.0.0.0:8117:8080 >/dev/null
run_new platform-ops-service -p 0.0.0.0:8110:8080 >/dev/null
run_new bff-courier -p 0.0.0.0:8112:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e REALTIME_URL=http://realtime-gateway:8080 >/dev/null
run_new bff-warehouse -p 0.0.0.0:8113:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  -e REALTIME_URL=http://realtime-gateway:8080 >/dev/null
run_new bff-admin -p 0.0.0.0:8114:8080 \
  -e ORDER_URL=http://order-service:8080 >/dev/null

echo "==> recreate bff-customer with TRACKING_URL (orders persist in order-service)"
docker rm -f nexora-staging-bff-customer >/dev/null 2>&1 || true
docker run -d --name nexora-staging-bff-customer --network "$NET" --network-alias bff-customer \
  -p 0.0.0.0:8111:8080 \
  -e HTTP_ADDR=:8080 -e DATABASE_URL= -e REDIS_URL= -e KAFKA_BROKERS= \
  -e OTP_DEV_MODE=true -e RATE_LIMIT_PER_MINUTE=0 \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e CATALOG_URL=http://catalog-service:8080 \
  -e CART_URL=http://cart-service:8080 \
  -e LOCATION_URL=http://location-service:8080 \
  -e CHECKOUT_URL=http://checkout-service:8080 \
  -e PAYMENT_URL=http://payment-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e TRACKING_URL=http://tracking-service:8080 \
  nexora/bff-customer:phone-staging >/dev/null

for s in tracking-service realtime-gateway supplier-service platform-ops-service \
         bff-customer bff-courier bff-warehouse bff-admin finance-ledger-service; do
  wait_health "$s"
done

echo "==> docker network probe"
docker run --rm --network "$NET" curlimages/curl:8.5.0 -fsS http://tracking-service:8080/health >/dev/null
docker run --rm --network "$NET" curlimages/curl:8.5.0 -fsS http://realtime-gateway:8080/health >/dev/null
echo "OK service-to-service"

echo "==> start web apps"
export NEXT_TELEMETRY_DISABLED=1
export NODE_OPTIONS="--max-old-space-size=256"
export NEXT_PUBLIC_TENANT_ID="$TENANT"
export NEXT_PUBLIC_IDENTITY_URL=""
export NEXT_PUBLIC_BFF_CUSTOMER_URL=""
export NEXT_PUBLIC_BFF_COURIER_URL=""
export NEXT_PUBLIC_BFF_WAREHOUSE_URL=""
export NEXT_PUBLIC_BFF_ADMIN_URL=""
export NEXT_PUBLIC_FINANCE_URL=""
export NEXT_PUBLIC_SUPPLIER_URL=""
export NEXT_PUBLIC_REALTIME_URL=""
export NEXT_PUBLIC_PLATFORM_OPS_URL=""
export NEXT_PUBLIC_ALLOW_MOCK_FALLBACK="false"
export BFF_CUSTOMER_INTERNAL=http://127.0.0.1:8111
export BFF_COURIER_INTERNAL=http://127.0.0.1:8112
export BFF_WAREHOUSE_INTERNAL=http://127.0.0.1:8113
export BFF_ADMIN_INTERNAL=http://127.0.0.1:8114
export REALTIME_INTERNAL=http://127.0.0.1:8115
export IDENTITY_INTERNAL=http://127.0.0.1:8081
export FINANCE_INTERNAL=http://127.0.0.1:8091
export SUPPLIER_INTERNAL=http://127.0.0.1:8117
export PLATFORM_OPS_INTERNAL=http://127.0.0.1:8110

# Keep customer :3000; restart it so realtime rewrite/route is picked up.
pkill -f "next dev --turbopack -p 3001" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3002" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3003" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3004" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3005" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3006" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3100" >/dev/null 2>&1 || true
pkill -f "next dev --turbopack -p 3200" >/dev/null 2>&1 || true
pkill -f "apps/customer-web" >/dev/null 2>&1 || true
pkill -f "next-server" >/dev/null 2>&1 || true
sleep 2

start_app() {
  local dir="$1"
  local name="$2"
  local port="$3"
  cd "$ROOT/apps/$dir"
  if [[ ! -d node_modules ]]; then
    echo "    npm install $name"
    npm install --no-audit --no-fund --silent
  fi
  echo "    start $name :$port"
  nohup npm run dev > "/tmp/prompt77-$name.log" 2>&1 &
  echo $! > "/tmp/prompt77-$name.pid"
}

start_app customer-web customer-web 3000
start_app warehouse-web warehouse-web 3002
start_app courier-web courier-web 3001
start_app supplier-web supplier-web 3003
start_app finance-web finance-web 3004
start_app support-web support-web 3005
start_app operations-web operations-web 3006
start_app admin_web admin_web 3100
start_app super_admin_web super_admin_web 3200

wait_http() {
  local port="$1"
  local name="$2"
  for _ in $(seq 1 90); do
    code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 2 "http://127.0.0.1:${port}/login" || true)"
    if [[ "$code" =~ ^(200|307|302|404) ]]; then
      echo "OK web $name :$port http=$code"
      return 0
    fi
    sleep 2
  done
  echo "FAIL web $name :$port last_http=${code:-none}"
  tail -n 40 "/tmp/prompt77-$name.log" 2>/dev/null || true
  return 1
}

wait_http 3000 customer-web || true
wait_http 3002 warehouse-web || true
wait_http 3001 courier-web || true
wait_http 3003 supplier-web || true
wait_http 3004 finance-web || true
wait_http 3005 support-web || true
wait_http 3006 operations-web || true
wait_http 3100 admin_web || true
wait_http 3200 super_admin_web || true

echo "==> free memory / containers"
free -h || true
docker ps --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'
df -h / | tail -n 1 || true

echo "==> go tests (role BFFs + customer track)"
(cd "$ROOT/services/bff-warehouse" && go test ./...)
(cd "$ROOT/services/bff-courier" && go test ./...)
(cd "$ROOT/services/bff-customer" && go test ./internal/app ./internal/adapters/httpclients)

echo "==> journey"
python3 "$ROOT/scripts/staging/prompt77-journey.py"

echo "PROMPT77_APPLY_DONE customer=$CUSTOMER_URL"
