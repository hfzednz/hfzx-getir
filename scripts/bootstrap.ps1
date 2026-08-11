$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "== NEXORA bootstrap =="
& "$PSScriptRoot\doctor.ps1"
& "$PSScriptRoot\verify-structure.ps1"

if (Get-Command docker -ErrorAction SilentlyContinue) {
  Write-Host "Starting local infra compose..."
  docker compose -f infra/docker/docker-compose.yml up -d
} else {
  Write-Host "docker not found — skip compose"
}

Write-Host "Next:"
Write-Host "  1) make deps-go / go work sync"
Write-Host "  2) pnpm install (web)"
Write-Host "  3) flutter pub get in apps/mobile_*"
Write-Host "  4) make test-go-focus"
Write-Host "See docs/monorepo/STRUCTURE.md and docs/guides/ONBOARDING.md"
