#!/usr/bin/env bash
# Rebuild BFF + order images, restart phone-test stack, seed catalog, start Next, run journey.
set -euo pipefail
ROOT="${ROOT:-/workspaces/hfzx-getir}"
cd "$ROOT"
NET="nexora-phone-staging"
TENANT="11111111-1111-1111-1111-111111111111"
PUBLIC="https://ominous-sniffle-v66jrrr567vcwpvj-3000.app.github.dev"

echo "==> verify Place source contains validate"
grep -q '/validate' services/bff-customer/internal/adapters/httpclients/clients.go
grep -q 'CheckoutAddress' services/bff-customer/internal/adapters/httpclients/clients.go
grep -q 'json:"variantId"' services/order-service/internal/app/create.go
echo "OK source markers"

build_one() {
  local s="$1"
  local tmp
  tmp="$(mktemp)"
  sed -E -e 's|golang:1\.22-alpine|golang:1.26-alpine|g' -e 's|golang:1\.23-alpine|golang:1.26-alpine|g' \
    "$ROOT/services/$s/Dockerfile" >"$tmp"
  docker build -f "$tmp" -t "nexora/${s}:phone-staging" "$ROOT/services/$s"
  rm -f "$tmp"
}

echo "==> rebuild bff-customer + order-service"
build_one bff-customer
build_one order-service

echo "==> verify image strings"
BFF_CID="$(docker create nexora/bff-customer:phone-staging)"
docker cp "$BFF_CID":/usr/local/bin/bff-customer /tmp/bff-customer.bin
docker rm "$BFF_CID" >/dev/null
if ! strings /tmp/bff-customer.bin | grep -q '/validate'; then
  echo "FAIL BFF image missing /validate"
  exit 1
fi
echo "OK BFF image has /validate"

docker network inspect "$NET" >/dev/null 2>&1 || docker network create "$NET" >/dev/null

run_or_start() {
  local name="$1"
  shift
  if docker inspect "$name" >/dev/null 2>&1; then
    docker start "$name" >/dev/null || true
  fi
}

echo "==> start unchanged containers"
for s in identity-service catalog-service cart-service location-service payment-service \
         inventory-service finance-ledger-service settlement-service; do
  docker start "nexora-staging-$s" >/dev/null || true
done

echo "==> recreate checkout with CART_URL/ORDER_URL"
docker rm -f nexora-staging-checkout-service >/dev/null 2>&1 || true
docker run -d --name nexora-staging-checkout-service --network "$NET" --network-alias checkout-service \
  -e HTTP_ADDR=:8080 -e DATABASE_URL= -e REDIS_URL= -e KAFKA_BROKERS= \
  -e OTP_DEV_MODE=true -e RATE_LIMIT_PER_MINUTE=0 \
  -e CART_URL=http://cart-service:8080 \
  -e ORDER_URL=http://order-service:8080 \
  -e PAYMENT_URL=http://payment-service:8080 \
  -e INVENTORY_URL=http://inventory-service:8080 \
  nexora/checkout-service:phone-staging >/dev/null

echo "==> recreate order-service (new image)"
docker rm -f nexora-staging-order-service >/dev/null 2>&1 || true
docker run -d --name nexora-staging-order-service --network "$NET" --network-alias order-service \
  -e HTTP_ADDR=:8080 -e DATABASE_URL= -e REDIS_URL= -e KAFKA_BROKERS= \
  -e OTP_DEV_MODE=true -e RATE_LIMIT_PER_MINUTE=0 \
  nexora/order-service:phone-staging >/dev/null

echo "==> recreate bff-customer (new image)"
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
  nexora/bff-customer:phone-staging >/dev/null

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
  docker logs "nexora-staging-$alias" 2>&1 | tail -n 30 || true
  return 1
}

for s in identity-service catalog-service cart-service location-service checkout-service \
         payment-service order-service inventory-service finance-ledger-service \
         settlement-service bff-customer; do
  wait_health "$s"
done

echo "==> seed catalog"
H='Content-Type: application/json'
TH="X-Tenant-Id: ${TENANT}"
seed() {
  docker run --rm --network "$NET" curlimages/curl:8.5.0 -sS -H "$H" -H "$TH" "$@"
}

seed -X POST http://catalog-service:8080/v1/catalog/products \
  -d '{"kind":"standard","slug":"fresh-milk","skuCode":"MILK-1"}' > /tmp/prod.json
echo "create product: $(cat /tmp/prod.json)"
PROD_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/prod.json"))
p=d.get("product") or d
print(p.get("id") or p.get("ID") or "")
PY
)"
if [[ -z "$PROD_ID" ]]; then
  echo "FAIL seed product id"
  exit 1
fi
seed -X PUT "http://catalog-service:8080/v1/catalog/products/${PROD_ID}/locales/en" \
  -d '{"title":"Fresh Milk","description":"1L"}' >/tmp/locale.json
seed -X POST "http://catalog-service:8080/v1/catalog/products/${PROD_ID}/variants" \
  -d '{"skuCode":"MILK-1L","name":"1L"}' >/tmp/var.json
echo "variant: $(cat /tmp/var.json)"
seed -X POST http://catalog-service:8080/v1/catalog/search/reindex -d '{}' >/tmp/reindex.json || true
echo "OK catalog seed product=$PROD_ID"

echo "==> start customer-web"
pkill -f "next dev" >/dev/null 2>&1 || true
cd "$ROOT/apps/customer-web"
if [[ ! -d node_modules ]]; then
  npm install
fi
export BFF_CUSTOMER_INTERNAL=http://127.0.0.1:8111
export NEXT_PUBLIC_BFF_CUSTOMER_URL="$PUBLIC"
export NEXT_PUBLIC_TENANT_ID="$TENANT"
nohup npm run dev > /tmp/customer-web.log 2>&1 &
for _ in $(seq 1 60); do
  code="$(curl -sS -o /dev/null -w '%{http_code}' --max-time 2 http://127.0.0.1:3000 || true)"
  if [[ "$code" =~ ^[2345] ]]; then
    echo "OK next :3000 http=$code"
    break
  fi
  sleep 2
done

echo "==> patch journey SKU"
python3 - <<PY
from pathlib import Path
p = Path("$ROOT/scripts/staging/codespace-journey.py")
t = p.read_text()
t = t.replace('SKU = "70fe49fa-7a60-4b7d-ae16-090116e8acbb"', 'SKU = "$PROD_ID"')
p.write_text(t)
print("journey sku", "$PROD_ID")
PY

echo "==> run journey"
python3 "$ROOT/scripts/staging/codespace-journey.py"
echo "APPLY_DONE product=$PROD_ID"
