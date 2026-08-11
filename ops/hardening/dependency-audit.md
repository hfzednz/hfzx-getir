# Dependency Audit Checklist

- [ ] `go list -m all` per service; no critical CVEs
- [ ] Flutter `pub outdated` reviewed for customer app
- [ ] Container base images weekly refresh
- [ ] Kafka/Redis/Postgres client libs within supported majors
- [ ] No secrets in git; Vault rotation cadence verified

Owned by security-service for vuln SoT; this checklist feeds hyperscale `security` gate.
