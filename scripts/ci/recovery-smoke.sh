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

wait_redis() {
  echo "==> wait redis"
  for _ in $(seq 1 30); do
    if docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli ping 2>/dev/null | grep -q PONG; then
      echo "OK redis"
      return 0
    fi
    sleep 2
  done
  echo "FAIL redis not ready"
  docker compose -p "$PROJECT" -f "$COMPOSE" ps
  return 1
}

wait_kafka() {
  echo "==> wait kafka broker API"
  for _ in $(seq 1 60); do
    if docker compose -p "$PROJECT" -f "$COMPOSE" exec -T kafka \
        bash -lc 'command -v kafka-broker-api-versions.sh >/dev/null && kafka-broker-api-versions.sh --bootstrap-server localhost:9092 >/dev/null 2>&1' \
      || docker compose -p "$PROJECT" -f "$COMPOSE" exec -T kafka \
        bash -lc 'test -x /opt/bitnami/kafka/bin/kafka-broker-api-versions.sh && /opt/bitnami/kafka/bin/kafka-broker-api-versions.sh --bootstrap-server localhost:9092 >/dev/null 2>&1'; then
      echo "OK kafka broker API"
      return 0
    fi
    sleep 3
  done
  echo "FAIL kafka broker not ready"
  docker compose -p "$PROJECT" -f "$COMPOSE" ps
  docker compose -p "$PROJECT" -f "$COMPOSE" logs kafka | tail -n 80
  return 1
}

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

wait_redis

echo "==> restart redis"
docker compose -p "$PROJECT" -f "$COMPOSE" restart redis
wait_redis

echo "==> postgres still ready after redis restart"
"${PG[@]}" pg_isready -U nexora -d nexora

echo "==> restart kafka"
docker compose -p "$PROJECT" -f "$COMPOSE" restart kafka
wait_kafka
"${PG[@]}" pg_isready -U nexora -d nexora
echo "OK postgres data plane intact after kafka restart"

echo "RECOVERY_SMOKE_PASS"
