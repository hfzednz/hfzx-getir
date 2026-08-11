#!/usr/bin/env bash
# Migration smoke: clean Postgres from infra compose → apply ALL service SQL migrations.
# Also verifies Redis PING and Kafka broker API (hard fail, not warn-only).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE="$ROOT/infra/docker/docker-compose.yml"
PROJECT="nexora-mig-smoke"

cleanup() {
  docker compose -p "$PROJECT" -f "$COMPOSE" down -v --remove-orphans >/dev/null 2>&1 || true
}
trap cleanup EXIT

echo "==> compose config"
docker compose -f "$COMPOSE" config -q

echo "==> start postgres + redis + kafka"
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

echo "==> wait redis"
for _ in $(seq 1 30); do
  if docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli PING 2>/dev/null | grep -q PONG; then
    break
  fi
  sleep 2
done
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli PING | grep -q PONG

echo "==> wait kafka"
kafka_ready=0
for _ in $(seq 1 60); do
  if docker compose -p "$PROJECT" -f "$COMPOSE" exec -T kafka \
      bash -lc 'command -v kafka-broker-api-versions.sh >/dev/null && kafka-broker-api-versions.sh --bootstrap-server localhost:9092 >/dev/null 2>&1' \
    || docker compose -p "$PROJECT" -f "$COMPOSE" exec -T kafka \
      bash -lc 'test -x /opt/bitnami/kafka/bin/kafka-broker-api-versions.sh && /opt/bitnami/kafka/bin/kafka-broker-api-versions.sh --bootstrap-server localhost:9092 >/dev/null 2>&1'; then
    kafka_ready=1
    break
  fi
  sleep 3
done
if [[ "$kafka_ready" -ne 1 ]]; then
  echo "ERROR: kafka broker not ready"
  docker compose -p "$PROJECT" -f "$COMPOSE" logs kafka | tail -n 80
  exit 1
fi

ensure_db() {
  local db="$1"
  local exists
  exists="$("${PG[@]}" psql -U nexora -d nexora -Atc "SELECT 1 FROM pg_database WHERE datname='${db}'" | tr -d '[:space:]')"
  if [[ "$exists" != "1" ]]; then
    "${PG[@]}" psql -U nexora -d nexora -v ON_ERROR_STOP=1 -c "CREATE DATABASE \"${db}\";"
  fi
  # Extensions are created by service migrations (pgcrypto); no silent || true here.
}

apply_svc() {
  local svc="$1"
  local dir="$ROOT/services/$svc/migrations"
  local db="nexora_${svc//-/_}"
  echo "==> migrate $svc → $db"
  ensure_db "$db"
  local f
  # shellcheck disable=SC2012
  for f in $(ls "$dir"/*.sql 2>/dev/null | sort); do
    echo ">> $svc $(basename "$f")"
    "${PG[@]}" psql -U nexora -d "$db" -v ON_ERROR_STOP=1 -f - <"$f"
  done
  # Minimal evidence tables were created
  local tables
  tables="$("${PG[@]}" psql -U nexora -d "$db" -Atc "SELECT count(*) FROM information_schema.tables WHERE table_schema='public'" | tr -d '[:space:]')"
  if [[ -z "$tables" || "$tables" -lt 1 ]]; then
    echo "ERROR: $svc produced no public tables"
    exit 1
  fi
  echo "OK $svc tables=$tables"
}

echo "==> discover migration-bearing services"
mapfile -t mig_dirs < <(find "$ROOT/services" -mindepth 2 -maxdepth 2 -type d -name migrations | sort)
if [[ "${#mig_dirs[@]}" -eq 0 ]]; then
  echo "ERROR: no migrations directories found"
  exit 1
fi
echo "MIGRATION_SERVICE_COUNT=${#mig_dirs[@]}"

fail=0
for dir in "${mig_dirs[@]}"; do
  svc="$(basename "$(dirname "$dir")")"
  if ! apply_svc "$svc"; then
    echo "FAIL migrate $svc"
    fail=$((fail + 1))
  fi
done
if [[ "$fail" -ne 0 ]]; then
  echo "MIGRATION_FAILS=$fail"
  exit 1
fi

echo "==> redis ping (application-level smoke)"
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli SET nexora:ci:smoke 1 EX 30 >/dev/null
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T redis redis-cli GET nexora:ci:smoke | grep -q 1

echo "==> kafka topic smoke"
docker compose -p "$PROJECT" -f "$COMPOSE" exec -T kafka bash -lc '
  BIN=""
  if command -v kafka-topics.sh >/dev/null 2>&1; then BIN=kafka-topics.sh
  elif test -x /opt/bitnami/kafka/bin/kafka-topics.sh; then BIN=/opt/bitnami/kafka/bin/kafka-topics.sh
  else echo "kafka-topics.sh missing"; exit 1; fi
  "$BIN" --bootstrap-server localhost:9092 --create --if-not-exists --topic nexora.ci.smoke --partitions 1 --replication-factor 1
  "$BIN" --bootstrap-server localhost:9092 --list | grep -q nexora.ci.smoke
'

echo "MIGRATION_SMOKE_OK services=${#mig_dirs[@]}"
