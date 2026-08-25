#!/usr/bin/env bash
# E2E smoke: in-memory identity/catalog/cart/location + BFFs, customer journeys, Playwright, ZAP.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
NET="nexora-e2e"
TENANT="11111111-1111-1111-1111-111111111111"
VARIANT="22222222-2222-2222-2222-222222222222"

SERVICES=(
  identity-service
  catalog-service
  cart-service
  location-service
  bff-customer
  bff-admin
)

cleanup() {
  for s in "${SERVICES[@]}"; do
    docker rm -f "nexora-e2e-$s" >/dev/null 2>&1 || true
  done
  docker network rm "$NET" >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> network"
docker network inspect "$NET" >/dev/null 2>&1 || docker network create "$NET" >/dev/null

build_one() {
  local s="$1"
  local tmp
  tmp="$(mktemp)"
  sed -E -e 's|golang:1\.22-alpine|golang:1.26-alpine|g' -e 's|golang:1\.23-alpine|golang:1.26-alpine|g' \
    "$ROOT/services/$s/Dockerfile" >"$tmp"
  docker build -f "$tmp" -t "nexora/${s}:e2e" "$ROOT/services/$s"
  rm -f "$tmp"
}

run_svc() {
  local s="$1"
  shift
  docker rm -f "nexora-e2e-$s" >/dev/null 2>&1 || true
  docker run -d --name "nexora-e2e-$s" --network "$NET" --network-alias "$s" \
    -e HTTP_ADDR=:8080 \
    -e DATABASE_URL= \
    -e REDIS_URL= \
    -e KAFKA_BROKERS= \
    "$@" \
    "nexora/${s}:e2e" >/dev/null
}

wait_health() {
  local name="$1"
  local ok=0
  for _ in $(seq 1 45); do
    local ip
    ip="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "$name" 2>/dev/null || true)"
    if [[ -n "$ip" ]] && curl -fsS --max-time 2 "http://${ip}:8080/health" >/dev/null 2>&1; then
      ok=1
      echo "OK health $name"
      break
    fi
    sleep 2
  done
  if [[ "$ok" -ne 1 ]]; then
    echo "FAIL health $name"
    docker logs "$name" 2>&1 | tail -n 50 || true
    return 1
  fi
}

for s in "${SERVICES[@]}"; do
  echo "==> build $s"
  build_one "$s"
done

echo "==> run backends"
run_svc identity-service
run_svc catalog-service
run_svc cart-service
run_svc location-service
run_svc bff-admin -p 8114:8080
run_svc bff-customer -p 8111:8080 \
  -e IDENTITY_URL=http://identity-service:8080 \
  -e CATALOG_URL=http://catalog-service:8080 \
  -e CART_URL=http://cart-service:8080 \
  -e LOCATION_URL=http://location-service:8080

for s in "${SERVICES[@]}"; do
  wait_health "nexora-e2e-$s"
done

HDR=(-H "Content-Type: application/json" -H "X-Tenant-Id: ${TENANT}")
CUST="http://127.0.0.1:8111"
CART_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-cart-service)"
CAT_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-catalog-service)"

dump_fail() {
  echo "FAIL $1"
  shift
  for s in "${SERVICES[@]}"; do
    echo "----- logs nexora-e2e-$s -----"
    docker logs "nexora-e2e-$s" 2>&1 | tail -n 30 || true
  done
  exit 1
}

http_json() {
  local out="$1" url="$2"
  shift 2
  local code
  code="$(curl -sS --max-time 10 -o "$out" -w "%{http_code}" "${HDR[@]}" "$url" "$@" || true)"
  echo "HTTP $code $url"
  if [[ -s "$out" ]]; then
    head -c 500 "$out"; echo
  fi
  if [[ "$code" != "200" && "$code" != "201" ]]; then
    dump_fail "HTTP $code $url"
  fi
}

echo "==> journey browse_catalog (BFF home without search index + catalog list)"
# Home with q= hits catalog search; empty OpenSearch-backed search is not required for this gate.
http_json /tmp/e2e-home.json "${CUST}/v1/customer/home?lat=41.0&lng=29.0"
python3 - <<'PY'
import json
p=json.load(open("/tmp/e2e-home.json"))
assert isinstance(p, dict), p
print("home_keys", sorted(p.keys()))
PY
http_json /tmp/e2e-products.json "http://${CAT_IP}:8080/v1/catalog/products?limit=5"
echo "OK browse_catalog"

echo "==> journey add_to_cart"
http_json /tmp/e2e-cart.json "http://${CART_IP}:8080/v1/cart" -d '{"guestToken":"e2e-guest","currency":"TRY"}'
CART_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-cart.json"))
print(d.get("ID") or d.get("id") or d.get("cartId") or "")
PY
)"
if [[ -z "$CART_ID" ]]; then
  dump_fail "create cart missing id"
fi
http_json /tmp/e2e-addcart.json "${CUST}/v1/customer/cart/items" \
  -d "{\"cartId\":\"${CART_ID}\",\"sku\":\"${VARIANT}\",\"qty\":1,\"unitMinor\":1500}"
echo "OK add_to_cart cartId=$CART_ID"

echo "==> journey auth otp start"
http_json /tmp/e2e-otp.json "${CUST}/v1/customer/auth/otp/start" -d "{\"phone\":\"+905551112233\"}"
python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-otp.json"))
cid=d.get("challengeId") or d.get("ChallengeID")
assert cid, d
print("challenge", cid)
PY
echo "OK auth otp start"

echo "==> Playwright (API request tests)"
cd "$ROOT/qa/playwright"
npm install --no-fund --no-audit
ADMIN_BASE="http://127.0.0.1:8114" CUSTOMER_BASE="http://127.0.0.1:8111" npx playwright test --reporter=list

echo "==> ZAP baseline against customer BFF"
set +e
docker run --rm --network "$NET" -t ghcr.io/zaproxy/zaproxy:stable \
  zap-baseline.py -t http://bff-customer:8080 -I
zap_rc=$?
set -e
# 0 = pass, 2 = warnings only; 1 = errors, 3 = high/critical
if [[ "$zap_rc" -eq 0 || "$zap_rc" -eq 2 ]]; then
  echo "OK zap-baseline rc=$zap_rc"
else
  echo "FAIL zap-baseline rc=$zap_rc"
  exit 1
fi

echo "E2E_SMOKE_PASS"
