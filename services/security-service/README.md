# NEXORA Security Service

Enterprise Security, Compliance, Governance & Risk (GRC) platform.

- HTTP `:8109` · `/v1/security`
- gRPC stub `:9109`
- Memory mode when `DATABASE_URL` empty
- Vault / OPA / SIEM / SOAR / IAM-trust / fraud facade / AI guardrail **ports**
- Does **not** own IAM credentials or payment PSP/PAN security

```bash
make test && make run
```

See `ARCHITECTURE.md` and `FEATURES.md`.
