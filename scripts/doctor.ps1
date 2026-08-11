$ErrorActionPreference = "Continue"
Write-Host "== doctor =="
foreach ($cmd in @("go","flutter","node","pnpm","docker","terraform","helm")) {
  $c = Get-Command $cmd -ErrorAction SilentlyContinue
  if ($c) {
    Write-Host "$cmd : $($c.Source)"
  } else {
    Write-Host "$cmd : MISSING"
  }
}
if (Get-Command go -ErrorAction SilentlyContinue) { go version }
