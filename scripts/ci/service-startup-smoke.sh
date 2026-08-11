#!/usr/bin/env bash
# Startup smoke: build+run a core commerce subset against compose infra; hit /health|/ready; SIGTERM.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE="$ROOT/infra/docker/docker-compose.yml"
PROJECT="nexora-startup-smoke"
NET="${PROJECT}_default"

SERVICES=(
  identity-service
  catalog-service
  cart-service
  checkout-service
  order-service
  payment-service
  inventory-service
  warehouse-service
  settlement-service
  finance-ledger-service
  realtime-gateway
  bff-customer
)

cleanup() {
  for s in "${SERVICES[@]}"; do
    docker rm -f "nexora-smoke-$s" >/dev/null 2>&1 || true
  done
  docker compose -p "$PROJECT" -f "$COMPOSE" down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> infra up"
docker compose -p "$PROJECT" -f "$COMPOSE" up -d postgres redis kafka
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T postgres pg_isready -U nexora -d nexora

# Ensure network exists (compose creates ${PROJECT}_default)
NET="$(docker compose -p "$PROJECT" -f "$COMPOSE" ps -q postgres | xargs -I{} docker inspect -f '{{range $k,$v := .NetworkSettings.Networks}}{{$k}}{{end}}' {})"
echo "NET=$NET"

fail=0
pass=0
port_base=18080
i=0
for s in "${SERVICES[@]}"; do
  i=$((i + 1))
  host_port=$((port_base + i))
  echo "==> build $s"
  tmp="$(mktemp)"
  sed -E -e 's|golang:1\.22-alpine|golang:1.26-alpine|g' -e 's|golang:1\.23-alpine|golang:1.26-alpine|g' \
    "$ROOT/services/$s/Dockerfile" >"$tmp"
  docker build -f "$tmp" -t "nexora/${s}:ci" "$ROOT/services/$s"
  rm -f "$tmp"

  echo "==> run $s on :$host_port"
  docker rm -f "nexora-smoke-$s" >/dev/null 2>&1 || true
  if ! docker run -d --name "nexora-smoke-$s" --network "$NET" \
    -p "${host_port}:8080" \
    -e HTTP_ADDR=:8080 \
    -e DATABASE_URL= \
    -e REDIS_URL= \
    -e KAFKA_BROKERS= \
    "nexora/${s}:ci" >/dev/null; then
    # Retry without host port publish (still force HTTP_ADDR for health probes via container IP).
    docker rm -f "nexora-smoke-$s" >/dev/null 2>&1 || true
    docker run -d --name "nexora-smoke-$s" --network "$NET" \
      -e HTTP_ADDR=:8080 \
      -e DATABASE_URL= \
      -e REDIS_URL= \
      -e KAFKA_BROKERS= \
      "nexora/${s}:ci" >/dev/null
  fi

  sleep 3
  ip="$(docker inspect -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' "nexora-smoke-$s")"
  ok=0
  for path in /health /ready /v1/health /livez /readyz; do
    if curl -fsS --max-time 3 "http://${ip}:8080${path}" >/dev/null 2>&1 \
      || curl -fsS --max-time 3 "http://127.0.0.1:${host_port}${path}" >/dev/null 2>&1; then
      ok=1
      echo "OK $s health via $path"
      break
    fi
  done
  # Probe common alternate listen ports from image ENV (8080-8126)
  if [[ "$ok" -eq 0 ]]; then
    for p in 8080 8081 8082 8083 8085 8086 8087 8088 8089 8090 8091 8092 8111 8115; do
      if curl -fsS --max-time 2 "http://${ip}:${p}/health" >/dev/null 2>&1 \
        || curl -fsS --max-time 2 "http://${ip}:${p}/ready" >/dev/null 2>&1; then
        ok=1
        echo "OK $s health on container port $p"
        break
      fi
    done
  fi
  if [[ "$ok" -ne 1 ]]; then
    echo "FAIL $s health"
    docker logs "nexora-smoke-$s" 2>&1 | tail -n 40 || true
    fail=$((fail + 1))
    continue
  fi

  echo "==> SIGTERM $s"
  docker stop -t 15 "nexora-smoke-$s" >/dev/null
  pass=$((pass + 1))
done

echo "STARTUP_PASS=$pass STARTUP_FAIL=$fail"
[[ "$fail" -eq 0 ]]
