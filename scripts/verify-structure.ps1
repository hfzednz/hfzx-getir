$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$required = @(
  "apps",
  "services",
  "packages",
  "infra",
  "ops",
  "docs",
  "qa",
  "tools",
  "store",
  ".github",
  "ADR",
  "go.work",
  "Makefile",
  "README.md",
  "docs/constitution/MASTER_BLUEPRINT.md",
  "docs/monorepo/STRUCTURE.md",
  "docs/launch/service-registry.yaml",
  "infra/docker/docker-compose.yml",
  "infra/argocd/applicationset.yaml",
  "infra/helm/nexora/Chart.yaml",
  "scripts/bootstrap.ps1"
)
$missing = @()
foreach ($rel in $required) {
  $p = Join-Path $Root $rel
  if (-not (Test-Path $p)) { $missing += $rel }
}
$svcCount = (Get-ChildItem (Join-Path $Root "services") -Directory).Count
if ($svcCount -lt 40) { $missing += "services count < 40 (got $svcCount)" }

if ($missing.Count -gt 0) {
  Write-Host "FAIL missing:"
  $missing | ForEach-Object { Write-Host " - $_" }
  exit 1
}
Write-Host "OK structure ($svcCount services)"
