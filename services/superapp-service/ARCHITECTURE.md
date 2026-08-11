# NEXORA Super App & Modular Ecosystem Platform

> Binding under Master Blueprint §7 (`superapp-service`).  
> Stack: Go control plane · Flutter shell/plugin SDK · PostgreSQL · Redis · Kafka · WASM/micro-frontend manifests · OTel.  
> **Hard rules:** Does **not** own domain SoT (orders/payments/catalog), open-platform API keys/webhooks SoT, liveops flag SoT, identity/wallet/loyalty balances.  
> Owns mini-app/plugin/widget registry, manifests, install lifecycle, permissions sandbox metadata, store listings, monetization *rules*, shell module resolution.

## Mission

Host independent Mini Apps, Plugins, Widgets, and Partner modules inside a single Super App shell — dynamically installable, remotely configurable, sandboxed, and production-ready.

## Architecture

```mermaid
flowchart LR
  Shell[Flutter Super App Shell] --> SUP[superapp-service :8121]
  SUP --> Registry[(Module Registry)]
  SUP --> Store[Plugin Store]
  SUP --> Perms[Permission PDP]
  Shell --> Hooks[Extension Hooks]
  Hooks --> Domains[Existing BFFs/Services]
  SUP --> LiveOps[liveops port]
  SUP --> Open[open-platform port]
  SUP --> Outbox --> Kafka
```

## Boundaries

| Owns | Does not own |
|------|----------------|
| Mini-app / plugin / widget catalog | Product catalog SoT |
| Install/update/remove lifecycle | Payment capture |
| Manifest + signing metadata | Wallet balances |
| Permission grants for modules | IAM sessions |
| Store ratings for plugins | Review SoT for products |
| Monetization commission *rules* | Settlement execution |
| Shell module resolution API | LiveOps experiment SoT |

## Folder structure

```text
services/superapp-service/
packages/flutter/superapp_shell/
packages/flutter/superapp_plugin_sdk/
docs/superapp/
qa/superapp/
```

## Dependency graph

```mermaid
flowchart TB
  Shell[superapp_shell] --> API[superapp-service]
  SDK[superapp_plugin_sdk] --> Shell
  API --> Registry[(Module Registry)]
  API --> Store[(Plugin Store)]
  API --> LiveOps
  API --> AI
  API --> Kafka[superapp.events]
  Shell --> Hooks --> BFF[Existing BFFs]
  BFF --> Domain[Domain SoTs]
  Partners --> OpenPlatform[open-platform-service]
```

## API (`:8121` `/v1/superapp/...`)

modules · mini-apps · plugins · widgets · store · installs · permissions · hooks · monetization · resolve · admin · outbox

## Events

`PluginInstalled` · `PluginRemoved` · `PluginUpdated` · `MiniAppLaunched` · `WidgetAdded` · `ModuleActivated` · `PermissionGranted`

## ER

```mermaid
erDiagram
  MODULE ||--o{ MODULE_VERSION : versions
  MODULE ||--o{ INSTALL : installed_by
  MODULE ||--o{ PERMISSION_GRANT : grants
  PLUGIN_LISTING ||--o{ RATING : rated
  WIDGET ||--o{ HOME_SLOT : placed
  MONETIZATION_RULE ||--o{ REVENUE_SHARE : applies
```
