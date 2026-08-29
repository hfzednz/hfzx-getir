#!/usr/bin/env bash
# Reseed the phone-staging catalog after a Codespace reboot.
# Catalog/cart/order services run in memory, so every restart starts with an empty catalog.
set -euo pipefail
ROOT="${ROOT:-/workspaces/hfzx-getir}"
NET="${NET:-nexora-phone-staging}"
TENANT="${TENANT:-11111111-1111-1111-1111-111111111111}"

seed() {
  docker run --rm --network "$NET" curlimages/curl:8.5.0 \
    -sS -H "Content-Type: application/json" -H "X-Tenant-Id: ${TENANT}" "$@"
}

seed -X POST http://catalog-service:8080/v1/catalog/products \
  -d '{"kind":"standard","slug":"fresh-milk","skuCode":"MILK-1"}' >/tmp/p80-prod.json
PROD_ID="$(python3 - <<'PY'
import json
d = json.load(open("/tmp/p80-prod.json"))
p = d.get("product") or d
print(p.get("id") or p.get("ID") or "")
PY
)"
if [[ -z "$PROD_ID" ]]; then
  echo "FAIL seed product id: $(cat /tmp/p80-prod.json)"
  exit 1
fi

seed -X PUT "http://catalog-service:8080/v1/catalog/products/${PROD_ID}/locales/en" \
  -d '{"title":"Fresh Milk","description":"1L"}' >/tmp/p80-locale.json
seed -X POST "http://catalog-service:8080/v1/catalog/products/${PROD_ID}/variants" \
  -d '{"skuCode":"MILK-1L","name":"1L"}' >/tmp/p80-variant.json
seed -X POST http://catalog-service:8080/v1/catalog/search/reindex -d '{}' >/tmp/p80-reindex.json || true

echo "OK catalog seed product=${PROD_ID}"
