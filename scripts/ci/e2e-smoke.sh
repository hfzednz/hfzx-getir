#!/usr/bin/env bash
# E2E smoke: in-memory identity/catalog/cart/location + BFFs, customer journeys, Playwright, ZAP.
# Set RC_FULL=1 to also boot checkout/payment/order/inventory/finance/settlement and run k6.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
NET="nexora-e2e"
TENANT="11111111-1111-1111-1111-111111111111"
VARIANT="22222222-2222-2222-2222-222222222222"
PRINCIPAL="22222222-2222-2222-2222-222222222222"
DEMO_CART="33333333-3333-3333-3333-333333333333"

SERVICES=(
  identity-service
  catalog-service
  cart-service
  location-service
  bff-customer
  bff-admin
)

if [[ "${RC_FULL:-}" == "1" ]]; then
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
    bff-admin
  )
fi

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
  # RATE_LIMIT_PER_MINUTE=0 disables in-memory IP quotas (location default 240/min).
  # k6 home traffic is one BFF IP; the 240/min cap is a product control, not a latency SLO.
  docker run -d --name "nexora-e2e-$s" --network "$NET" --network-alias "$s" \
    -e HTTP_ADDR=:8080 \
    -e DATABASE_URL= \
    -e REDIS_URL= \
    -e KAFKA_BROKERS= \
    -e RATE_LIMIT_PER_MINUTE=0 \
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
if [[ "${RC_FULL:-}" == "1" ]]; then
  run_svc checkout-service
  run_svc payment-service
  run_svc order-service
  run_svc inventory-service
  run_svc finance-ledger-service
  run_svc settlement-service
fi
run_svc bff-admin -p 8114:8080
if [[ "${RC_FULL:-}" == "1" ]]; then
  run_svc bff-customer -p 8111:8080 \
    -e IDENTITY_URL=http://identity-service:8080 \
    -e CATALOG_URL=http://catalog-service:8080 \
    -e CART_URL=http://cart-service:8080 \
    -e LOCATION_URL=http://location-service:8080 \
    -e CHECKOUT_URL=http://checkout-service:8080 \
    -e PAYMENT_URL=http://payment-service:8080 \
    -e ORDER_URL=http://order-service:8080
else
  run_svc bff-customer -p 8111:8080 \
    -e IDENTITY_URL=http://identity-service:8080 \
    -e CATALOG_URL=http://catalog-service:8080 \
    -e CART_URL=http://cart-service:8080 \
    -e LOCATION_URL=http://location-service:8080
fi

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

http_expect() {
  local expect="$1" out="$2" url="$3"
  shift 3
  local code
  code="$(curl -sS --max-time 10 -o "$out" -w "%{http_code}" "$@" "$url" || true)"
  echo "HTTP $code (want $expect) $url"
  if [[ -s "$out" ]]; then
    head -c 300 "$out"; echo
  fi
  if [[ "$code" != "$expect" ]]; then
    dump_fail "expected HTTP $expect got $code $url"
  fi
}

echo "==> observability request id"
RID="e2e-obs-57"
OBS_HDR="$(mktemp)"
curl -sS --max-time 10 -D "$OBS_HDR" -o /tmp/e2e-obs.json \
  -H "X-Request-Id: ${RID}" -H "X-Tenant-Id: ${TENANT}" \
  "${CUST}/health" >/dev/null
if ! grep -qi "X-Request-Id: ${RID}" "$OBS_HDR"; then
  dump_fail "BFF did not echo X-Request-Id"
fi
curl -sS --max-time 10 -D "$OBS_HDR" -o /tmp/e2e-obs-cat.json \
  -H "X-Request-Id: ${RID}" -H "X-Tenant-Id: ${TENANT}" \
  "http://${CAT_IP}:8080/health" >/dev/null
if ! grep -qi "X-Request-Id: ${RID}" "$OBS_HDR"; then
  dump_fail "catalog did not echo X-Request-Id"
fi
echo "OK observability"

echo "==> negative journeys"
http_expect 400 /tmp/e2e-otp-nophone.json "${CUST}/v1/customer/auth/otp/start" \
  "${HDR[@]}" -d '{"phone":""}'
http_expect 400 /tmp/e2e-otp-notenant.json "${CUST}/v1/customer/auth/otp/start" \
  -H "Content-Type: application/json" -d '{"phone":"+905551112233"}'
http_json /tmp/e2e-otp-neg.json "${CUST}/v1/customer/auth/otp/start" -d '{"phone":"+905551119999"}'
CHALLENGE="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-otp-neg.json"))
print(d.get("challengeId") or d.get("ChallengeID") or "")
PY
)"
BAD_OTP="$(curl -sS --max-time 10 -o /tmp/e2e-otp-bad.json -w "%{http_code}" "${HDR[@]}" \
  "${CUST}/v1/customer/auth/otp/verify" -d "{\"challengeId\":\"${CHALLENGE}\",\"code\":\"000000\"}" || true)"
echo "HTTP $BAD_OTP (want 400|401) otp verify"
if [[ "$BAD_OTP" != "400" && "$BAD_OTP" != "401" ]]; then
  dump_fail "expected 400/401 otp verify got $BAD_OTP"
fi
EMPTY_CART="$(curl -sS --max-time 10 -o /tmp/e2e-empty-cart.json -w "%{http_code}" "${HDR[@]}" \
  "${CUST}/v1/customer/cart/items" -d '{"cartId":"","sku":"x","qty":1,"unitMinor":100}' || true)"
echo "HTTP $EMPTY_CART (want 4xx/502) empty cart"
if [[ "$EMPTY_CART" != 4* && "$EMPTY_CART" != "502" ]]; then
  dump_fail "expected client/upstream error empty cart got $EMPTY_CART"
fi
http_expect 404 /tmp/e2e-bad-product.json "http://${CAT_IP}:8080/v1/catalog/products/00000000-0000-0000-0000-000000000000" \
  "${HDR[@]}"
# Business state after negatives: health still 200.
http_json /tmp/e2e-health-after-neg.json "${CUST}/health"
echo "OK negative"

echo "==> journey browse_catalog (BFF home without search index + catalog list)"
http_json /tmp/e2e-home.json "${CUST}/v1/customer/home?lat=41.0&lng=29.0"
python3 - <<'PY'
import json
p=json.load(open("/tmp/e2e-home.json"))
assert isinstance(p, dict), p
print("home_keys", sorted(p.keys()))
PY
http_json /tmp/e2e-products.json "http://${CAT_IP}:8080/v1/catalog/products?limit=5"
echo "OK browse_catalog"

echo "==> journey catalog product create + detail"
http_json /tmp/e2e-create-prod.json "http://${CAT_IP}:8080/v1/catalog/products" \
  -d '{"kind":"standard","slug":"e2e-milk","skuCode":"E2E-MILK"}'
PROD_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-create-prod.json"))
p=d.get("product") or d
print(p.get("id") or p.get("ID") or "")
PY
)"
if [[ -z "$PROD_ID" ]]; then
  dump_fail "create product missing id"
fi
http_json /tmp/e2e-get-prod.json "http://${CAT_IP}:8080/v1/catalog/products/${PROD_ID}"
http_json /tmp/e2e-var.json "http://${CAT_IP}:8080/v1/catalog/products/${PROD_ID}/variants" \
  -d '{"skuCode":"E2E-MILK-1","name":"1L"}'
echo "OK product_detail productId=$PROD_ID"

echo "==> journey add_to_cart + duplicate add"
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
http_json /tmp/e2e-addcart2.json "${CUST}/v1/customer/cart/items" \
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

if [[ "${RC_FULL:-}" == "1" ]]; then
  CHK_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-checkout-service)"
  PAY_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-payment-service)"
  ORD_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-order-service)"
  INV_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-inventory-service)"
  FIN_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-finance-ledger-service)"
  SET_IP="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nexora-e2e-settlement-service)"

  echo "==> RC checkout preview + address + validate + complete (idempotent)"
  http_json /tmp/e2e-preview.json "${CUST}/v1/customer/checkout/preview" \
    -H "X-Nexora-User: ${PRINCIPAL}" \
    -d "{\"cartId\":\"${DEMO_CART}\",\"principalId\":\"${PRINCIPAL}\"}"
  SESS_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-preview.json"))
print(d.get("sessionId") or d.get("SessionID") or d.get("id") or "")
PY
)"
  if [[ -z "$SESS_ID" ]]; then
    dump_fail "checkout preview missing sessionId"
  fi
  http_json /tmp/e2e-addr.json "http://${CHK_IP}:8080/v1/checkout/sessions/${SESS_ID}" \
    -X PATCH \
    -d '{"address":{"line1":"Test St 1","city":"Istanbul","lat":41.0,"lng":29.0}}'
  http_json /tmp/e2e-validate.json "http://${CHK_IP}:8080/v1/checkout/sessions/${SESS_ID}/validate" -d '{}'
  python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-validate.json"))
status=(d.get("status") or "").lower()
assert status in ("ready","completed"), d
print("checkout_status", status, "total", (d.get("quote") or {}).get("totalMinor"))
PY
  http_json /tmp/e2e-place.json "${CUST}/v1/customer/checkout/place" \
    -H "X-Nexora-User: ${PRINCIPAL}" \
    -d "{\"cartId\":\"${DEMO_CART}\",\"paymentMethod\":\"card\",\"sessionId\":\"${SESS_ID}\",\"principalId\":\"${PRINCIPAL}\"}"
  ORDER_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-place.json"))
print(d.get("orderId") or d.get("OrderID") or "")
PY
)"
  if [[ -z "$ORDER_ID" ]]; then
    dump_fail "place missing orderId"
  fi
  http_json /tmp/e2e-place2.json "${CUST}/v1/customer/checkout/place" \
    -H "X-Nexora-User: ${PRINCIPAL}" \
    -d "{\"cartId\":\"${DEMO_CART}\",\"paymentMethod\":\"card\",\"sessionId\":\"${SESS_ID}\",\"principalId\":\"${PRINCIPAL}\"}"
  ORDER_ID2="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-place2.json"))
print(d.get("orderId") or d.get("OrderID") or "")
PY
)"
  if [[ "$ORDER_ID" != "$ORDER_ID2" ]]; then
    dump_fail "duplicate place created two orders $ORDER_ID vs $ORDER_ID2"
  fi
  echo "OK checkout_place orderId=$ORDER_ID"

  echo "==> RC order create + history + detail + duplicate create"
  http_json /tmp/e2e-ord.json "http://${ORD_IP}:8080/v1/orders" \
    -d "{\"customerPrincipalId\":\"${PRINCIPAL}\",\"type\":\"instant\",\"currency\":\"TRY\",\"idempotencyKey\":\"rc-order-1\",\"lines\":[{\"SKUCode\":\"E2E-MILK-1\",\"TitleSnapshot\":\"Milk\",\"Qty\":1,\"UnitPriceMinor\":1500,\"VariantID\":\"44444444-4444-4444-4444-444444444444\"}]}"
  OMS_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-ord.json"))
o=d.get("order") or d
print(o.get("id") or o.get("ID") or "")
PY
)"
  http_json /tmp/e2e-ord2.json "http://${ORD_IP}:8080/v1/orders" \
    -d "{\"customerPrincipalId\":\"${PRINCIPAL}\",\"type\":\"instant\",\"currency\":\"TRY\",\"idempotencyKey\":\"rc-order-1\",\"lines\":[{\"SKUCode\":\"E2E-MILK-1\",\"TitleSnapshot\":\"Milk\",\"Qty\":1,\"UnitPriceMinor\":1500,\"VariantID\":\"44444444-4444-4444-4444-444444444444\"}]}"
  OMS_ID2="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-ord2.json"))
o=d.get("order") or d
print(o.get("id") or o.get("ID") or "")
PY
)"
  if [[ -n "$OMS_ID" && "$OMS_ID" != "$OMS_ID2" ]]; then
    dump_fail "duplicate order create $OMS_ID vs $OMS_ID2"
  fi
  http_json /tmp/e2e-ord-list.json "http://${ORD_IP}:8080/v1/orders?limit=5"
  if [[ -n "$OMS_ID" ]]; then
    http_json /tmp/e2e-ord-get.json "http://${ORD_IP}:8080/v1/orders/${OMS_ID}"
  fi
  echo "OK order_history omsId=$OMS_ID"

  echo "==> RC payment intent + capture + duplicate refund"
  http_json /tmp/e2e-pay.json "http://${PAY_IP}:8080/v1/payments/intents" \
    -H "X-Nexora-User: ${PRINCIPAL}" \
    -d "{\"orderId\":\"rc-pay-1\",\"amountMinor\":1500,\"currency\":\"TRY\",\"methodType\":\"card\",\"idempotencyKey\":\"rc-pay-int-1\",\"principalId\":\"${PRINCIPAL}\"}"
  INTENT="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-pay.json"))
print(d.get("id") or (d.get("intent") or {}).get("id") or "")
PY
)"
  http_json /tmp/e2e-auth.json "http://${PAY_IP}:8080/v1/payments/intents/${INTENT}/authorize" \
    -d '{"token":"tok_ok","idempotencyKey":"rc-pay-auth-1"}'
  http_json /tmp/e2e-cap.json "http://${PAY_IP}:8080/v1/payments/intents/${INTENT}/capture" \
    -d '{"idempotencyKey":"rc-pay-cap-1"}'
  http_json /tmp/e2e-ref.json "http://${PAY_IP}:8080/v1/payments/intents/${INTENT}/refund" \
    -d '{"amountMinor":500,"reason":"partial","idempotencyKey":"rc-pay-ref-1"}'
  http_json /tmp/e2e-ref2.json "http://${PAY_IP}:8080/v1/payments/intents/${INTENT}/refund" \
    -d '{"amountMinor":500,"reason":"partial","idempotencyKey":"rc-pay-ref-1"}'
  python3 - <<'PY'
import json
a=json.load(open("/tmp/e2e-ref.json"))
b=json.load(open("/tmp/e2e-ref2.json"))
aid=(a.get("refund") or a).get("id")
bid=(b.get("refund") or b).get("id")
assert aid and aid==bid, (a,b)
print("refund_id", aid)
PY
  echo "OK payment intent=$INTENT"

  echo "==> RC inventory receive + duplicate reserve"
  WH="aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaa1"
  VAR="bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbb1"
  http_json /tmp/e2e-wh.json "http://${INV_IP}:8080/v1/inventory/warehouses" \
    -d '{"code":"RC1","name":"RC WH","timezone":"UTC","status":"active"}'
  WH_ID="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-wh.json"))
print(d.get("id") or d.get("ID") or (d.get("warehouse") or {}).get("id") or (d.get("warehouse") or {}).get("ID") or "")
PY
)"
  if [[ -z "$WH_ID" ]]; then WH_ID="$WH"; fi
  http_json /tmp/e2e-recv.json "http://${INV_IP}:8080/v1/inventory/stock/receive" \
    -d "{\"warehouseId\":\"${WH_ID}\",\"variantId\":\"${VAR}\",\"skuCode\":\"E2E-MILK-1\",\"qty\":10,\"idempotencyKey\":\"rc-recv-1\"}"
  http_json /tmp/e2e-res.json "http://${INV_IP}:8080/v1/inventory/reservations/soft" \
    -d "{\"idempotencyKey\":\"rc-soft-1\",\"lines\":[{\"warehouseId\":\"${WH_ID}\",\"variantId\":\"${VAR}\",\"skuCode\":\"E2E-MILK-1\",\"qty\":2}]}"
  http_json /tmp/e2e-res2.json "http://${INV_IP}:8080/v1/inventory/reservations/soft" \
    -d "{\"idempotencyKey\":\"rc-soft-1\",\"lines\":[{\"warehouseId\":\"${WH_ID}\",\"variantId\":\"${VAR}\",\"skuCode\":\"E2E-MILK-1\",\"qty\":2}]}"
  python3 - <<'PY'
import json
def oid(d):
    if not isinstance(d, dict):
        return ""
    for k in ("id", "ID"):
        if d.get(k):
            return str(d[k])
    for wrap in ("reservation", "warehouse"):
        w = d.get(wrap) or {}
        for k in ("id", "ID"):
            if isinstance(w, dict) and w.get(k):
                return str(w[k])
    return ""
a=json.load(open("/tmp/e2e-res.json"))
b=json.load(open("/tmp/e2e-res2.json"))
aid, bid = oid(a), oid(b)
assert aid and aid==bid, (a,b)
print("reservation", aid)
PY
  echo "OK inventory"

  echo "==> RC finance ledger + settlement idempotency"
  http_json /tmp/e2e-acc1.json "http://${FIN_IP}:8080/v1/ledger/accounts" \
    -d '{"code":"1000","name":"Cash","type":"asset","currency":"TRY"}'
  http_json /tmp/e2e-acc2.json "http://${FIN_IP}:8080/v1/ledger/accounts" \
    -d '{"code":"4000","name":"Revenue","type":"revenue","currency":"TRY"}'
  CASH="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-acc1.json"))
print(d.get("id") or "")
PY
)"
  REV="$(python3 - <<'PY'
import json
d=json.load(open("/tmp/e2e-acc2.json"))
print(d.get("id") or "")
PY
)"
  JBODY="{\"currency\":\"TRY\",\"reference\":\"rc-pay\",\"idempotencyKey\":\"rc-j-1\",\"lines\":[{\"accountId\":\"${CASH}\",\"debitMinor\":1500},{\"accountId\":\"${REV}\",\"creditMinor\":1500}]}"
  http_json /tmp/e2e-j1.json "http://${FIN_IP}:8080/v1/ledger/journals" -d "$JBODY"
  http_json /tmp/e2e-j2.json "http://${FIN_IP}:8080/v1/ledger/journals" -d "$JBODY"
  python3 - <<'PY'
import json
a=json.load(open("/tmp/e2e-j1.json"))
b=json.load(open("/tmp/e2e-j2.json"))
assert a.get("id") and a.get("id")==b.get("id"), (a,b)
print("journal", a.get("id"), "debit", a.get("debitTotal"))
PY
  http_json /tmp/e2e-set1.json "http://${SET_IP}:8080/v1/settlements/batches" \
    -d '{"currency":"TRY","periodStart":"2026-08-01T00:00:00Z","periodEnd":"2026-08-07T00:00:00Z","idempotencyKey":"rc-set-1","actorId":"'"${PRINCIPAL}"'"}'
  http_json /tmp/e2e-set2.json "http://${SET_IP}:8080/v1/settlements/batches" \
    -d '{"currency":"TRY","periodStart":"2026-08-01T00:00:00Z","periodEnd":"2026-08-07T00:00:00Z","idempotencyKey":"rc-set-1","actorId":"'"${PRINCIPAL}"'"}'
  python3 - <<'PY'
import json
a=json.load(open("/tmp/e2e-set1.json"))
b=json.load(open("/tmp/e2e-set2.json"))
assert a.get("id") and a.get("id")==b.get("id"), (a,b)
print("batch", a.get("id"))
PY
  echo "OK finance_settlement"

  echo "==> k6 staged load (GHA-scale)"
  docker run --rm --network "$NET" \
    -e BFF_BASE=http://bff-customer:8080 \
    -e TENANT_ID="$TENANT" \
    -v "$ROOT/qa/k6:/scripts" grafana/k6:latest run \
    --summary-trend-stats="avg,min,med,p(50),p(95),p(99),max" \
    /scripts/rc_bff.js
  echo "OK k6"
fi

echo "==> Playwright (API request tests)"
cd "$ROOT/qa/playwright"
npm install --no-fund --no-audit
ADMIN_BASE="http://127.0.0.1:8114" CUSTOMER_BASE="http://127.0.0.1:8111" npx playwright test --project=api --reporter=list

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
