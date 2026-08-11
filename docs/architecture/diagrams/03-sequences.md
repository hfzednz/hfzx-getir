# NEXORA — Sequence Diagrams (Critical Paths)

Companion to Master Blueprint §8 and §29.

---

## 1. OTP Authentication

```mermaid
sequenceDiagram
  autonumber
  actor U as User
  participant App as Mobile App
  participant BFF as bff-*
  participant ID as identity-service
  participant OTP as OTP Provider
  participant RD as Redis

  U->>App: Enter phone
  App->>BFF: POST /v1/auth/otp/start
  BFF->>ID: StartOtp(phone, device)
  ID->>RD: rate_limit + store hash
  ID->>OTP: Send code
  OTP-->>U: SMS
  U->>App: Enter code
  App->>BFF: POST /v1/auth/otp/verify
  BFF->>ID: VerifyOtp
  ID->>RD: compare & consume
  ID-->>BFF: access + refresh
  BFF-->>App: session payload
  App->>App: Secure storage save
```

---

## 2. Add to Cart with Soft Stock Hold

```mermaid
sequenceDiagram
  participant App
  participant BFF as bff-customer
  participant Cart as cart-service
  participant Inv as inventory-service
  participant RD as Redis

  App->>BFF: POST /v1/cart/items
  BFF->>Cart: AddItem
  Cart->>Inv: ReserveSoft(store, sku, qty, ttl)
  alt insufficient
    Inv-->>Cart: UNAVAILABLE
    Cart-->>BFF: 409 STOCK_UNAVAILABLE
    BFF-->>App: show substitutes
  else ok
    Inv->>RD: reservation key TTL
    Inv-->>Cart: reservation_id
    Cart-->>BFF: cart snapshot
    BFF-->>App: updated cart
  end
```

---

## 3. Live Tracking Fanout

```mermaid
sequenceDiagram
  participant Courier as Courier App
  participant BFF as bff-courier
  participant TRK as tracking-service
  participant KF as Kafka
  participant RT as realtime-gateway
  participant Cust as Customer App
  participant Admin as Admin Web

  Courier->>BFF: POST /v1/location (batched)
  BFF->>TRK: IngestLocation
  TRK->>KF: tracking.location.events
  KF->>RT: consume
  RT-->>Cust: WS order.location
  RT-->>Admin: WS city.live
```

---

## 4. Warehouse Scan Offline Sync

```mermaid
sequenceDiagram
  participant WH as Warehouse App
  participant Drift as Drift Local DB
  participant Outbox as Mutation Outbox
  participant BFF as bff-warehouse
  participant WHS as warehouse-service

  WH->>Drift: apply scan locally
  WH->>Outbox: enqueue mutation
  Note over WH,Outbox: offline possible
  Outbox->>BFF: flush when online (Idempotency-Key)
  BFF->>WHS: ApplyScan
  alt accepted
    WHS-->>BFF: new version
    BFF-->>Outbox: ack
    Outbox->>Drift: mark synced
  else conflict
    WHS-->>BFF: 409 version conflict
    BFF-->>Outbox: refetch truth
    Outbox->>Drift: rebase + user prompt if needed
  end
```
