# Super App QA Suite

| Suite | Scope |
|-------|--------|
| Unit | `go test ./...` in `services/superapp-service` |
| Compatibility | Manifest `minShellVersion` vs shell |
| Sandbox | Permission allow-list validation |
| Lifecycle | Install → update → rollback → remove |
| Store | Browse / rate / monetization rules |
| Load | Shell resolve under concurrent subjects (k6 stub in qa/) |

## Cert gates

- Seed mini-apps returns ≥17 modules
- Resolve only returns LiveOps-enabled installs
- Signed manifests required (signature + checksum)
