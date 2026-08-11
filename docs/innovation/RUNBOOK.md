# Innovation Runbook

1. `POST /v1/innovation/bootstrap/catalog`
2. Enable TRL-ready modules: `POST /v1/innovation/modules/enable` `{"key":"green.carbon"}`
3. Simulations: start → complete by id
4. Register edge/IoT/robots/drones as needed
5. Drain outbox: `POST /v1/innovation/outbox/publish`

LiveOps port may deny enablement; escalate to liveops-service for flags.
