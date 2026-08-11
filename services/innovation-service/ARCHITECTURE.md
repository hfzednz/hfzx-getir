# NEXORA Innovation & Future Expansion Platform

> Binding under Master Blueprint §7 (`innovation-service`).  
> Stack: Go control plane · optional Python/Rust research adapters · Flutter XR/voice stubs · Kafka · WASM/edge hooks · OTel.  
> **Hard rules:** Does **not** replace OMS, inventory, dispatch, payments, AI model SoT, LiveOps flags, Super App module registry, or open-platform keys.  
> Owns experimental/innovation modules, simulation runs, digital-twin *models*, IoT/robot/drone *registries & missions metadata*, edge node registry, TRL scoring, green-tech metrics, quantum-safe *hooks*, research sandboxes.

## Mission

Future-proof the ecosystem with optional, isolated, production-ready innovation modules — no redesign of existing architecture required to adopt emerging tech.

## Architecture

```mermaid
flowchart LR
  Lab[Innovation Lab] --> INV[innovation-service :8122]
  INV --> Sims[Simulation Engine]
  INV --> Twins[Digital Twins]
  INV --> Edge[Edge Registry]
  INV --> IoT[IoT Registry]
  INV --> Robots[Robot/Drone APIs]
  INV --> Outbox --> Kafka
  Sims -.read models.-> DomainPorts[Existing services via ports]
  INV --> LiveOps[enable gate port]
  INV --> AI[research assist port]
```

## Boundaries

| Owns | Does not own |
|------|----------------|
| Innovation feature registry + TRL | LiveOps experiment SoT |
| Simulation job lifecycle | OMS / inventory stock SoT |
| Digital twin model metadata | Real warehouse WMS |
| IoT device registry + telemetry *ingest meta* | Hardware firmware |
| Robot/drone fleet *registry* + mission plans | Physical fleet control SoT |
| Optional blockchain/XR/quantum hooks | Payment / ledger SoT |
| Carbon / energy *metrics* | ERP sustainability filings |

## Folder structure

```text
services/innovation-service/
docs/innovation/
qa/innovation/
packages/flutter/innovation_xr_stub/
packages/python/innovation_research/
packages/rust/innovation_edge_stub/
```

## Dependency graph

```mermaid
flowchart TB
  Admin[Admin Console] --> INV
  INV --> Sims
  INV --> Twins
  INV --> Edge
  INV --> IoT
  INV --> Robots
  INV --> Green
  INV --> Lab
  Sims --> Ports[inventory/dispatch/pricing ports]
  Edge --> WASM[WASM inference hooks]
  Lab --> SuperApp[optional module publish via port]
```

## API (`:8122` `/v1/innovation/...`)

lab · modules · simulations · twins · edge · iot · robots · drones · blockchain · xr · multimodal · green · quantum · research · admin · outbox

## Events

`SimulationStarted` · `SimulationCompleted` · `InnovationEnabled` · `ResearchExperimentCreated` · `EdgeNodeRegistered` · `IoTDeviceConnected` · `RobotAssigned` · `DroneMissionCreated`

## ER

```mermaid
erDiagram
  INNOVATION_MODULE ||--o{ RESEARCH_EXPERIMENT : incubates
  SIMULATION_RUN ||--o{ SIM_METRIC : produces
  DIGITAL_TWIN ||--o{ TWIN_STATE : snapshots
  EDGE_NODE ||--o{ EDGE_JOB : runs
  IOT_DEVICE ||--o{ TELEMETRY_META : reports
  ROBOT ||--o{ ROBOT_ASSIGNMENT : assigned
  DRONE ||--o{ DRONE_MISSION : flies
  GREEN_METRIC ||--o{ EMISSION_REPORT : aggregates
```
