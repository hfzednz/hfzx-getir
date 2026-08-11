# NEXORA — System Context & Container Diagrams

> Companion to `docs/constitution/MASTER_BLUEPRINT.md`  
> Format: Mermaid (source of truth). Export to PNG/SVG in CI later if needed.

---

## 1. System Context

```mermaid
C4Context
title NEXORA System Context
Person(customer, "Customer", "Orders essentials in minutes")
Person(courier, "Courier", "Delivers orders")
Person(picker, "Warehouse Associate", "Picks and packs")
Person(ops, "City Ops / Admin", "Runs cities")
Person(super, "Super Admin", "Platform governance")

System(nexora, "NEXORA Platform", "Quick commerce OS")

System_Ext(psp, "Payment Service Provider", "Card/wallet capture")
System_Ext(sms, "SMS/OTP Provider", "Auth & alerts")
System_Ext(maps, "Maps / Routing Provider", "Navigation & distance")
System_Ext(fcm, "Push Provider", "FCM/APNs")
System_Ext(object, "Object Storage/CDN", "Media delivery")

Rel(customer, nexora, "Uses apps & APIs")
Rel(courier, nexora, "Uses courier app")
Rel(picker, nexora, "Uses warehouse app")
Rel(ops, nexora, "Uses admin")
Rel(super, nexora, "Uses super admin")
Rel(nexora, psp, "Charges/refunds")
Rel(nexora, sms, "Sends OTP/SMS")
Rel(nexora, maps, "ETA/routes")
Rel(nexora, fcm, "Push")
Rel(nexora, object, "Stores/serves media")
```

---

## 2. High-Level Container View

```mermaid
flowchart TB
  subgraph Clients
    CA[Customer Flutter]
    CO[Courier Flutter]
    WH[Warehouse Flutter]
    AD[Admin Web]
    SA[Super Admin Web]
  end

  subgraph Edge
    CDN[CDN]
    GW[API Gateway / WAF]
    RT[Realtime Gateway]
  end

  subgraph BFF
    BC[bff-customer]
    BR[bff-courier]
    BW[bff-warehouse]
    BA[bff-admin]
  end

  subgraph Domains
    ID[Identity]
    CAT[Catalog]
    INV[Inventory]
    CART[Cart]
    ORD[Order]
    PAY[Payment]
    WHSVC[Warehouse]
    DISP[Dispatch]
    TRK[Tracking]
    SRCH[Search]
    AI[AI Services]
    NOTIF[Notifications]
  end

  subgraph DataPlane
    PG[(PostgreSQL)]
    RD[(Redis)]
    KF[(Kafka)]
    OS[(OpenSearch)]
    CH[(ClickHouse)]
    S3[(Object Storage)]
  end

  CA --> CDN
  CA --> GW
  CO --> GW
  WH --> GW
  AD --> GW
  SA --> GW
  CA --> RT
  CO --> RT
  AD --> RT

  GW --> BC & BR & BW & BA
  BC & BR & BW & BA --> ID & CAT & INV & CART & ORD & PAY & WHSVC & DISP & TRK & SRCH & AI & NOTIF
  ID & CAT & INV & CART & ORD & PAY & WHSVC & DISP --> PG
  CART & DISP & TRK --> RD
  ORD & INV & WHSVC & DISP --> KF
  SRCH --> OS
  KF --> CH
  CAT & TRK --> S3
  RT --> TRK
```

---

## 3. Fulfillment Flow (Order Lifecycle)

```mermaid
stateDiagram-v2
  [*] --> DraftCart
  DraftCart --> CheckoutStarted
  CheckoutStarted --> StockReserved
  StockReserved --> PaymentPending
  PaymentPending --> Paid: payment_ok
  PaymentPending --> Cancelled: payment_fail
  Paid --> Picking
  Picking --> Packed
  Packed --> AwaitingCourier
  AwaitingCourier --> PickedUp
  PickedUp --> InTransit
  InTransit --> Delivered
  Delivered --> [*]
  Picking --> PartiallyCancelled: stock_issue
  Paid --> Cancelled: customer_cancel_window
  InTransit --> FailedDelivery
  FailedDelivery --> ReturnedToStore
```

---

## 4. Event Backbone Topics (core)

| Topic | Producers | Consumers |
|-------|-----------|-----------|
| `identity.user.events` | identity | CRM, analytics |
| `catalog.product.events` | catalog | search-indexer, cache |
| `inventory.stock.events` | inventory | search, admin, recommendations |
| `order.order.events` | order | warehouse, dispatch, notify, analytics, finance |
| `warehouse.task.events` | warehouse | dispatch, tracking, admin |
| `dispatch.assignment.events` | dispatch | tracking, courier BFF, notify |
| `tracking.location.events` | tracking | realtime-gateway, ETA |
| `payment.payment.events` | payment | order, finance, fraud |
| `notification.delivery.events` | notification | analytics |

---

## 5. City as Scale Unit

```mermaid
flowchart LR
  subgraph CityA[City: IST]
    SA1[Store A1]
    SA2[Store A2]
    CA[Couriers A]
  end
  subgraph CityB[City: ANK]
    SB1[Store B1]
    CB[Couriers B]
  end
  CTRL[Control Plane / Super Admin]
  CTRL --> CityA
  CTRL --> CityB
```

All domain partitions, configs, fees, SLAs, and flags are addressable by `city_id`.
