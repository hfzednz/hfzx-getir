#!/usr/bin/env bash
# Security sanity: tracked-secret scan + govulncheck per Go module.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

echo "==> tracked secret scan"
bad="$(git ls-files | grep -E '(^|/)\.env$|(^|/)\.env\.[^/]+$|\.pem$|\.p12$|\.pfx$|(^|/)id_rsa|(^|/)id_ed25519|(^|/)service-account.*\.json$|(^|/)credentials\.json$' || true)"
# allow committed examples only
if [[ -n "$bad" ]]; then
  filtered="$(printf '%s\n' "$bad" | grep -vE '\.example(\.|$)|env\.example' || true)"
  if [[ -n "$filtered" ]]; then
    echo "FAIL tracked secret-like files:"
    printf '%s\n' "$filtered"
    exit 1
  fi
fi
echo "OK no tracked secrets"

echo "==> govulncheck"
export GOWORK=off
export GOTOOLCHAIN=auto
go install golang.org/x/vuln/cmd/govulncheck@latest
fail=0
for mod in services/*/go.mod; do
  dir="$(dirname "$mod")"
  name="$(basename "$dir")"
  echo "==> govulncheck $name"
  if ! (cd "$dir" && govulncheck ./...); then
    echo "FAIL govulncheck $name"
    fail=$((fail + 1))
  fi
done
echo "GOVULN_FAILS=$fail"
test "$fail" -eq 0
echo "SECURITY_SANITY_PASS"
