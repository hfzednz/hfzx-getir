# 19 — Map Experience

## Principles

Operational truth over spectacle. Markers and ETA must match state machines. Zone fills subtle (`opacity.mapZone` 0.14).

---

## Modes

| Mode | Surface | Elements |
|------|---------|----------|
| Address refine | Customer | Draggable pin, accuracy circle, confirm CTA |
| Live tracking | Customer | Courier marker, store pin, destination, route, ETA chip |
| Courier navigation | Courier | External nav handoff preferred; in-app overview map |
| Warehouse context | Rare | Store outline only |
| Ops live map | Admin | Many couriers, heat, stores, SLA markers |
| Delivery zones | Admin editor | Polygons edit; snap; validate |

---

## Pins & markers

| Entity | Visual |
|--------|--------|
| Customer destination | Teal pin |
| Dark store | Graphite store glyph |
| Courier | Accent/brand circular avatar+arrow heading |
| Selected | Scale 1.1 + elevation |
| Cluster (admin) | Count badge |

## Heatmaps (admin)

- Demand / delay heat using brand→warning scale
- Legend required
- Toggleable; off by default for performance

## ETA on map

- `NxEtaCard` floating top or bottom-leading
- Updates without full map reload
- Breathe motion on live only

## Delivery zones

- Polygon stroke `border.strong` / brand
- Fill mapZone opacity
- Unserviceable: danger stroke when pin outside

## Animations

- Camera fit bounds padding 48–72
- Marker interpolation
- Reduced motion: jump cut camera

## Controls

- Recenter, layer toggles (admin), zoom (+/− desktop)
- Never obscure CTA with legal attribution — pad bottom

## Privacy

- Courier precise location only while assigned & in transit
- Historical paths retained per policy; UI defaults to current segment
