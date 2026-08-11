#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
echo "== NEXORA bootstrap =="
bash scripts/doctor.sh
bash scripts/verify-structure.sh
if command -v docker >/dev/null 2>&1; then
  docker compose -f infra/docker/docker-compose.yml up -d
else
  echo "docker not found — skip compose"
fi
echo "Next: make deps && make test-go-focus"
echo "See docs/monorepo/STRUCTURE.md"
