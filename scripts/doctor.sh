#!/usr/bin/env bash
set -euo pipefail
echo "== doctor =="
for cmd in go flutter node pnpm docker terraform helm; do
  if command -v "$cmd" >/dev/null 2>&1; then
    echo "$cmd: $(command -v "$cmd")"
  else
    echo "$cmd: MISSING"
  fi
done
command -v go >/dev/null 2>&1 && go version || true
