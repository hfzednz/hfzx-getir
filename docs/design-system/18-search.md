# 18 — Search Experience

## Goals

Fast path to product in seconds. Typo-tolerant. Honest empty states. Multi-modal (text, voice, barcode, image, NL) with one results canvas.

---

## Instant search

- Debounce 150–200ms
- Min chars: 1 for CJK/short SKUs configurable; default 2 for Latin
- Cancel in-flight on new keystroke
- Skeleton sparingly — prefer previous suggestions

## Layout (mobile)

1. Search field focused on entry  
2. Before query: **Recent** + **Trending**  
3. While typing: **Suggestions** (products, categories, queries)  
4. On submit / suggestion tap: **Results** with filters sheet  

## Suggestions

- Product hits with thumbnail + price
- Category shortcuts
- Query completions
- “Search in {category}”

## Recent / Trending

- Recent: local Drift/Hive; clear all
- Trending: city-scoped; privacy-safe aggregates

## Voice search

- Trailing mic on `NxSearchField`
- Permission rationale
- Partial transcript live in field
- Hand off to same suggestion pipeline

## Barcode / QR search

- Scanner sheet; haptic on hit
- Direct PDP if unique SKU; else results

## Image search

- Camera / gallery → uploading state → candidate grid
- Confidence low → “Not sure — try text”

## Semantic / NL search

- Example chips: “milk lactose free under 50₺”
- Parse feedback: show applied filters as chips (editable)
- Fallback to keyword if NL unavailable

## Filters & sort

- Sheet with chips summary on bar
- Counts on facets when known
- Clear all

## Empty / error

- Illustration + tips + category shortcuts
- No results: keep query editable; suggest corrections

## Analytics hooks (spec only)

Exposure events for suggestion impressions; never log raw voice audio in product analytics.
