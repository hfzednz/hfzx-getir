#!/usr/bin/env bash
# Build every service Dockerfile under services/*/Dockerfile (context = service dir).
# Bumps golang:1.22 → 1.26 in a temp Dockerfile so go.mod requiring 1.23+/1.26 can build.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export DOCKER_BUILDKIT=1

fail=0
pass=0
mapfile -d '' dfs < <(find "$ROOT/services" -mindepth 2 -maxdepth 2 -name Dockerfile -print0 | sort -z)
if [[ "${#dfs[@]}" -eq 0 ]]; then
  echo "ERROR: no Dockerfiles found"
  exit 1
fi

for df in "${dfs[@]}"; do
  [[ -z "$df" ]] && continue
  dir="$(dirname "$df")"
  name="$(basename "$dir")"
  echo "==> docker build $name"
  tmp="$(mktemp)"
  # Satisfy modern go.mod versions without rewriting committed Dockerfiles permanently.
  sed -E \
    -e 's|golang:1\.22-alpine|golang:1.26-alpine|g' \
    -e 's|golang:1\.23-alpine|golang:1.26-alpine|g' \
    -e 's|(RUN[[:space:]]+)(CGO_ENABLED=0)|\1GOTOOLCHAIN=local \2|g' \
    "$df" >"$tmp"
  if docker build -f "$tmp" -t "nexora/${name}:ci" "$dir"; then
    pass=$((pass + 1))
  else
    echo "FAIL $name"
    fail=$((fail + 1))
  fi
  rm -f "$tmp"
done

echo "DOCKER_PASS=$pass DOCKER_FAIL=$fail EXPECTED=${#dfs[@]}"
[[ "$fail" -eq 0 ]]
[[ "$pass" -eq "${#dfs[@]}" ]]
