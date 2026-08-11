#!/usr/bin/env bash
# Race-test every Go service module (requires CGO + gcc; use ubuntu-latest).
# Discovers only services/*/go.mod (not nested packages). GOWORK=off avoids workspace coupling.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
export CGO_ENABLED=1
export GOWORK=off
export GOTOOLCHAIN=auto

fail=0
pass=0
mapfile -d '' mods < <(find "$ROOT/services" -mindepth 2 -maxdepth 2 -name go.mod -print0 | sort -z)
if [[ "${#mods[@]}" -eq 0 ]]; then
  echo "ERROR: no service go.mod files found under $ROOT/services"
  exit 1
fi

for mod in "${mods[@]}"; do
  [[ -z "$mod" ]] && continue
  dir="$(dirname "$mod")"
  name="$(basename "$dir")"
  echo "==> race $name"
  if (cd "$dir" && go test -race -count=1 ./...); then
    pass=$((pass + 1))
  else
    echo "FAIL $name"
    fail=$((fail + 1))
  fi
done

echo "RACE_PASS=$pass RACE_FAIL=$fail"
if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
echo "RACE_PASS_ALL"
