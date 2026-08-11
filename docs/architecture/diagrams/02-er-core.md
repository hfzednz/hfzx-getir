# NEXORA — Core ER Diagrams (Logical)

> Logical model per bounded context. Physical DBs are separated per service.  
> Money fields: `*_amount_minor` + `currency_code`. IDs: UUIDv7/ULID.

---

## 1. Identity & Profile

```mermaid
erDiagram
  USER ||--o{ DEVICE : has
  USER ||--o{ SESSION : has
  USER ||--o{ AUTH_FACTOR : has
  USER ||--|| CUSTOMER_PROFILE : extends
  CUSTOMER_PROFILE ||--o{ ADDRESS : has
  CUSTOMER_PROFILE ||--o{ CONSENT : has

  USER {
    ulid id PK
    string phone UK
    string email UK
    string status
    timestamptz created_at
  }
  DEVICE {
    ulid id PK
    ulid user_id FK
    string device_fingerprint
    string push_token
    string platform
  }
  SESSION {
    ulid id PK
    ulid user_id FK
    string refresh_hash
    timestamptz expires_at
  }
  ADDRESS {
    ulid id PK
    ulid profile_id FK
    string label
    float lat
    float lng
    string city_id
    bool is_default
  }
```

---

## 2. Catalog & Pricing

```mermaid
erDiagram
  CATEGORY ||--o{ CATEGORY : parent
  CATEGORY ||--o{ PRODUCT : contains
  PRODUCT ||--o{ PRODUCT_VARIANT : has
  PRODUCT ||--o{ PRODUCT_MEDIA : has
  PRODUCT ||--o{ PRODUCT_ATTRIBUTE : has
  PRODUCT_VARIANT ||--o{ CITY_PRICE : priced_in

  PRODUCT {
    ulid id PK
    string sku_base UK
    string name
    string brand
    bool is_active
  }
  PRODUCT_VARIANT {
    ulid id PK
    ulid product_id FK
    string sku UK
    string barcode
    int unit_size
  }
  CITY_PRICE {
    ulid id PK
    ulid variant_id FK
    string city_id
    int amount_minor
    string currency_code
    timestamptz effective_from
  }
```

---

## 3. Inventory

```mermaid
erDiagram
  DARK_STORE ||--o{ STOCK_ITEM : holds
  PRODUCT_VARIANT ||--o{ STOCK_ITEM : tracked_as
  STOCK_ITEM ||--o{ STOCK_RESERVATION : reserves
  STOCK_ITEM ||--o{ STOCK_LEDGER : movements

  DARK_STORE {
    ulid id PK
    string city_id
    string name
    string status
  }
  STOCK_ITEM {
    ulid id PK
    ulid store_id FK
    ulid variant_id FK
    int on_hand
    int reserved
    int available
    int version
  }
  STOCK_RESERVATION {
    ulid id PK
    ulid stock_item_id FK
    ulid cart_or_order_id
    int qty
    string status
    timestamptz expires_at
  }
  STOCK_LEDGER {
    ulid id PK
    ulid stock_item_id FK
    string reason
    int delta
    ulid actor_id
    timestamptz created_at
  }
```

---

## 4. Cart → Order → Payment

```mermaid
erDiagram
  CART ||--o{ CART_ITEM : contains
  CART ||--o| CHECKOUT_SESSION : opens
  CHECKOUT_SESSION ||--|| ORDER : places
  ORDER ||--o{ ORDER_ITEM : contains
  ORDER ||--o{ ORDER_STATUS_HISTORY : tracks
  ORDER ||--o{ PAYMENT_ATTEMPT : pays
  ORDER ||--o| FULFILLMENT : fulfills

  CART {
    ulid id PK
    ulid user_id
    string city_id
    ulid store_id
    string status
    int version
  }
  ORDER {
    ulid id PK
    string public_code UK
    ulid user_id
    string city_id
    ulid store_id
    string status
    int subtotal_minor
    int delivery_fee_minor
    int discount_minor
    int tip_minor
    int total_minor
    string currency_code
    string idempotency_key UK
  }
  PAYMENT_ATTEMPT {
    ulid id PK
    ulid order_id FK
    string psp_ref
    string status
    int amount_minor
  }
```

---

## 5. Warehouse & Dispatch

```mermaid
erDiagram
  ORDER ||--|| PICK_TASK : generates
  PICK_TASK ||--o{ PICK_TASK_LINE : has
  PICK_TASK ||--o| PACK_TASK : next
  PACK_TASK ||--o| HANDOFF : stages
  ORDER ||--o| DISPATCH_ASSIGNMENT : assigns
  DISPATCH_ASSIGNMENT ||--|| COURIER : to
  DISPATCH_ASSIGNMENT ||--o{ TRACKING_EVENT : emits

  PICK_TASK {
    ulid id PK
    ulid order_id FK
    ulid store_id
    ulid picker_id
    string status
    int sla_seconds
  }
  DISPATCH_ASSIGNMENT {
    ulid id PK
    ulid order_id FK
    ulid courier_id FK
    string status
    timestamptz assigned_at
  }
  TRACKING_EVENT {
    ulid id PK
    ulid order_id
    string type
    float lat
    float lng
    timestamptz at
  }
```

---

## 6. Wallet & Loyalty

```mermaid
erDiagram
  WALLET ||--o{ WALLET_ENTRY : ledger
  LOYALTY_ACCOUNT ||--o{ LOYALTY_LEDGER : points
  LOYALTY_ACCOUNT ||--o{ LOYALTY_TIER_HISTORY : tiers

  WALLET {
    ulid id PK
    ulid user_id UK
    string currency_code
    int balance_minor
  }
  WALLET_ENTRY {
    ulid id PK
    ulid wallet_id FK
    int delta_minor
    string reason
    ulid reference_id
  }
  LOYALTY_ACCOUNT {
    ulid id PK
    ulid user_id UK
    int points_balance
    string tier
  }
```
