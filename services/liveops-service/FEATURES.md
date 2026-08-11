# LiveOps Features

- Flags: boolean, %, geo/city/warehouse/segment/role/OS/version rules, dependencies, emergency off
- Sticky subject bucketing (FNV)
- Remote config namespaces with LiveOps event patches
- Experiments: AB/AA/MVT/canary, sticky assign, complete + optional auto-rollout, AI winner hint port
- Approvals for change requests; instant rollback (flag/config/experiment/emergency)
- Outbox on `liveops.events`; metrics port to data-platform
