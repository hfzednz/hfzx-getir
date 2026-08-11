$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "== Prompt-45 build validate =="
& "$PSScriptRoot\verify-structure.ps1"
& "$PSScriptRoot\test-go-focus.ps1"

Push-Location "$Root\services\hyperscale-cert-service"
go test ./...
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

Write-Host "Go focus + hyperscale OK"
Write-Host "Optional: cd apps/mobile_customer; flutter analyze --no-fatal-infos"
Write-Host "Optional: cd apps/admin_web; npm run lint"
Write-Host "Reports: docs/build/"
