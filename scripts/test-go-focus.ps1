$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$services = @(
  "identity-service",
  "order-service",
  "payment-service",
  "catalog-service",
  "bff-customer",
  "bff-admin"
)
foreach ($s in $services) {
  Write-Host "=== $s ==="
  Push-Location (Join-Path $Root "services\$s")
  go test ./...
  if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
  Pop-Location
}
Write-Host "OK focus Go tests"
