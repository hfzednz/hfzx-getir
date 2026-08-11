#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
missing=0
req=(
  apps services packages infra ops docs qa tools store .github ADR
  go.work Makefile README.md
  docs/constitution/MASTER_BLUEPRINT.md
  docs/monorepo/STRUCTURE.md
  docs/launch/service-registry.yaml
  infra/docker/docker-compose.yml
  infra/argocd/applicationset.yaml
  infra/helm/nexora/Chart.yaml
  scripts/bootstrap.sh
)
for rel in "${req[@]}"; do
  if [[ ! -e "$ROOT/$rel" ]]; then
    echo "MISSING $rel"
    missing=1
  fi
done
svc=$(find "$ROOT/services" -mindepth 1 -maxdepth 1 -type d | wc -l | tr -d ' ')
if [[ "$svc" -lt 40 ]]; then
  echo "MISSING services count < 40 (got $svc)"
  missing=1
fi
if [[ "$missing" -ne 0 ]]; then
  exit 1
fi
echo "OK structure ($svc services)"
