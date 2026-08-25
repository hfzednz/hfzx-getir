#!/usr/bin/env bash
# Disposable CI recovery: restart Redis/Kafka, verify Postgres still ready.
# Never run against production infrastructure.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE="$ROOT/infra/docker/docker-compose.yml"
PROJECT="nexora-rc-recovery"

cleanup() {
  docker compose -p "$PROJECT" -f "$COMPOSE" down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> infra up"
docker compose -p "$PROJECT" -f "$COMPOSE" up -d postgres redis kafka

PG=(docker compose -p "$PROJECT" -f "$COMPOSE" exec -T postgres)
echo "==> wait postgres"
for _ in $(seq 1 60); do
  if "${PG[@]}" pg_isready -U nexora -d nexora >/dev/null 2>&1; then
    break
  fi
  sleep 2
done
"${PG[@]}" pg_isready -U nexora -d nexora

echo "==> redis ping"
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli ping | grep -q PONG

echo "==> restart redis"
docker compose -p "$PROJECT" -f "$COMPOSE" restart redis
for _ in $(seq 1 30); do
  if docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli ping 2>/dev/null | grep -q PONG; then
    echo "OK redis after restart"
    break
  fi
  sleep 2
done
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli ping | grep -q PONG

echo "==> postgres still ready after redis restart"
"${PG[@]}" pg_isready -U nexora -d nexora

echo "==> restart kafka"
docker compose -p "$PROJECT" -f "$COMPOSE" restart kafka
sleep 8
docker compose -p "$PROJECT" -f "$COMPOSE" ps kafka | grep -qi "running\|up" || {
  echo "FAIL kafka not running after restart"
  docker compose -p "$PROJECT" -f "$COMPOSE" ps
  exit 1
}
"${PG[@]}" pg_isready -U nexora -d nexora
echo "OK postgres data plane intact after kafka restart"

echo "RECOVERY_SMOKE_PASS"
