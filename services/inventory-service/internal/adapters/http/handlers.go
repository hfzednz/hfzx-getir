package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
	"github.com/nexora/inventory-service/internal/ratelimit"
)

// Handler serves inventory REST endpoints.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

// NewHandler returns a fully wired http.Handler.
func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps, Ready: cfg.Ready, Live: cfg.Live}
	mux := http.NewServeMux()
	const base = "/v1/inventory"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("GET "+base+"/warehouses", tenant(h.listWarehouses))
	mux.HandleFunc("POST "+base+"/warehouses", tenant(h.createWarehouse))
	mux.HandleFunc("GET "+base+"/warehouses/{id}", tenant(h.getWarehouse))
	mux.HandleFunc("PATCH "+base+"/warehouses/{id}", tenant(h.updateWarehouse))
	mux.HandleFunc("DELETE "+base+"/warehouses/{id}", tenant(h.deleteWarehouse))

	mux.HandleFunc("GET "+base+"/warehouses/{id}/locations", tenant(h.listLocations))
	mux.HandleFunc("POST "+base+"/warehouses/{id}/locations", tenant(h.createLocation))
	mux.HandleFunc("GET "+base+"/locations/{locationId}", tenant(h.getLocation))
	mux.HandleFunc("PATCH "+base+"/locations/{locationId}", tenant(h.updateLocation))
	mux.HandleFunc("DELETE "+base+"/locations/{locationId}", tenant(h.deleteLocation))

	mux.HandleFunc("GET "+base+"/stock", tenant(h.listStock))
	mux.HandleFunc("GET "+base+"/stock/{id}", tenant(h.getStock))
	mux.HandleFunc("POST "+base+"/stock/ensure", tenant(h.ensureBalance))
	mux.HandleFunc("POST "+base+"/stock/adjust", tenant(h.adjust))
	mux.HandleFunc("POST "+base+"/stock/receive", tenant(h.receive))
	mux.HandleFunc("POST "+base+"/stock/damage", tenant(h.damage))
	mux.HandleFunc("POST "+base+"/stock/waste", tenant(h.waste))
	mux.HandleFunc("GET "+base+"/stock/{id}/movements", tenant(h.listMovements))

	mux.HandleFunc("POST "+base+"/reservations/soft", tenant(h.softReserve))
	mux.HandleFunc("POST "+base+"/reservations/{id}/confirm-hard", tenant(h.confirmHard))
	mux.HandleFunc("POST "+base+"/reservations/{id}/extend", tenant(h.extend))
	mux.HandleFunc("POST "+base+"/reservations/{id}/release", tenant(h.release))
	mux.HandleFunc("POST "+base+"/reservations/{id}/consume", tenant(h.consume))
	mux.HandleFunc("GET "+base+"/reservations/{id}", tenant(h.getReservation))

	mux.HandleFunc("GET "+base+"/atp", tenant(h.queryATP))

	mux.HandleFunc("POST "+base+"/transfers", tenant(h.createTransfer))
	mux.HandleFunc("POST "+base+"/transfers/{id}/approve", tenant(h.approveTransfer))
	mux.HandleFunc("POST "+base+"/transfers/{id}/complete", tenant(h.completeTransfer))
	mux.HandleFunc("GET "+base+"/transfers/{id}", tenant(h.getTransfer))

	mux.HandleFunc("POST "+base+"/counts", tenant(h.startCount))
	mux.HandleFunc("POST "+base+"/counts/{id}/submit", tenant(h.submitCount))
	mux.HandleFunc("POST "+base+"/counts/{id}/approve", tenant(h.approveCount))
	mux.HandleFunc("GET "+base+"/counts/{id}", tenant(h.getCount))

	mux.HandleFunc("GET "+base+"/lots/near-expiry", tenant(h.nearExpiry))
	mux.HandleFunc("POST "+base+"/lots/fefo-allocate", tenant(h.fefoAllocate))

	mux.HandleFunc("POST "+base+"/returns", tenant(h.receiveReturn))
	mux.HandleFunc("GET "+base+"/returns/{id}", tenant(h.getReturn))

	mux.HandleFunc("GET "+base+"/forecasts", tenant(h.listForecasts))
	mux.HandleFunc("PUT "+base+"/forecasts", tenant(h.upsertForecast))
	mux.HandleFunc("POST "+base+"/forecasts/generate", tenant(h.generateForecast))
	mux.HandleFunc("GET "+base+"/forecasts/{id}", tenant(h.getForecast))

	mux.HandleFunc("GET "+base+"/search", tenant(h.search))
	mux.HandleFunc("POST "+base+"/search/reindex", tenant(h.reindex))

	mux.HandleFunc("GET "+base+"/admin/explorer", tenant(h.explorer))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

// NewServer builds an *http.Server with sensible timeouts.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	if h.Live != nil {
		if err := h.Live(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) ready(w http.ResponseWriter, r *http.Request) {
	if h.Ready != nil {
		if err := h.Ready(r); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeOK(w, map[string]string{"status": "ready"})
}

func (h *Handler) tenant(r *http.Request) uuid.UUID {
	tid, _ := TenantIDFromContext(r.Context())
	return tid
}

func parseUUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}

func idemKey(r *http.Request, bodyKey string) string {
	if v := r.Header.Get("Idempotency-Key"); v != "" {
		return v
	}
	return bodyKey
}

func parseActor(r *http.Request) *uuid.UUID {
	if v := r.Header.Get("X-Actor-Id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return &id
		}
	}
	return nil
}

// --- Warehouses ---

func (h *Handler) createWarehouse(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code     string                 `json:"code"`
		Name     string                 `json:"name"`
		RegionID *uuid.UUID             `json:"regionId"`
		Timezone string                 `json:"timezone"`
		Status   domain.WarehouseStatus `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.CreateWarehouse(r.Context(), app.CreateWarehouseInput{
		TenantID: h.tenant(r), Code: body.Code, Name: body.Name,
		RegionID: body.RegionID, Timezone: body.Timezone, Status: body.Status,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) listWarehouses(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListWarehouses(r.Context(), ports.WarehouseFilter{
		TenantID: h.tenant(r), Query: r.URL.Query().Get("q"), Limit: limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) getWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetWarehouse(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) updateWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Name     *string                 `json:"name"`
		RegionID *uuid.UUID              `json:"regionId"`
		Timezone *string                 `json:"timezone"`
		Status   *domain.WarehouseStatus `json:"status"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.UpdateWarehouse(r.Context(), app.UpdateWarehouseInput{
		TenantID: h.tenant(r), ID: id, Name: body.Name, RegionID: body.RegionID,
		Timezone: body.Timezone, Status: body.Status,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) deleteWarehouse(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.DeleteWarehouse(r.Context(), h.tenant(r), id); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "deleted"})
}

// --- Locations ---

func (h *Handler) createLocation(w http.ResponseWriter, r *http.Request) {
	whID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ParentID   *uuid.UUID        `json:"parentId"`
		Kind       domain.LocationKind `json:"kind"`
		ZoneType   *domain.ZoneType  `json:"zoneType"`
		Code       string            `json:"code"`
		Name       string            `json:"name"`
		IsPickable bool              `json:"isPickable"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.CreateLocation(r.Context(), app.CreateLocationInput{
		TenantID: h.tenant(r), WarehouseID: whID, ParentID: body.ParentID,
		Kind: body.Kind, ZoneType: body.ZoneType, Code: body.Code, Name: body.Name,
		IsPickable: body.IsPickable,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) listLocations(w http.ResponseWriter, r *http.Request) {
	whID, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListLocations(r.Context(), whID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) getLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "locationId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetLocation(r.Context(), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) updateLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "locationId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Name       *string          `json:"name"`
		ZoneType   *domain.ZoneType `json:"zoneType"`
		IsPickable *bool            `json:"isPickable"`
		IsActive   *bool            `json:"isActive"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.UpdateLocation(r.Context(), app.UpdateLocationInput{
		ID: id, Name: body.Name, ZoneType: body.ZoneType, IsPickable: body.IsPickable, IsActive: body.IsActive,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) deleteLocation(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "locationId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.DeleteLocation(r.Context(), id); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "deleted"})
}

// --- Stock ---

func (h *Handler) ensureBalance(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID  `json:"warehouseId"`
		VariantID   uuid.UUID  `json:"variantId"`
		SKUCode     string     `json:"skuCode"`
		LocationID  *uuid.UUID `json:"locationId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.EnsureBalance(r.Context(), app.EnsureBalanceInput{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		SKUCode: body.SKUCode, LocationID: body.LocationID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) listStock(w http.ResponseWriter, r *http.Request) {
	whRaw := r.URL.Query().Get("warehouseId")
	whID, err := uuid.Parse(whRaw)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListBalances(r.Context(), h.tenant(r), whID, limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

func (h *Handler) getStock(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetBalance(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) adjust(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID    uuid.UUID  `json:"warehouseId"`
		VariantID      uuid.UUID  `json:"variantId"`
		SKUCode        string     `json:"skuCode"`
		LocationID     *uuid.UUID `json:"locationId"`
		Delta          int64      `json:"delta"`
		IdempotencyKey string     `json:"idempotencyKey"`
		Reason         string     `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	bal, mov, err := h.Deps.Adjust(r.Context(), app.AdjustStockInput{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		SKUCode: body.SKUCode, LocationID: body.LocationID, Delta: body.Delta,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r), Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"balance": bal, "movement": mov})
}

func (h *Handler) receive(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID    uuid.UUID  `json:"warehouseId"`
		VariantID      uuid.UUID  `json:"variantId"`
		SKUCode        string     `json:"skuCode"`
		LocationID     *uuid.UUID `json:"locationId"`
		Qty            int64      `json:"qty"`
		LotCode        string     `json:"lotCode"`
		IdempotencyKey string     `json:"idempotencyKey"`
		Reason         string     `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	bal, mov, err := h.Deps.Receive(r.Context(), app.ReceiveStockCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		SKUCode: body.SKUCode, LocationID: body.LocationID, Qty: body.Qty, LotCode: body.LotCode,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r), Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"balance": bal, "movement": mov})
}

func (h *Handler) damage(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID    uuid.UUID  `json:"warehouseId"`
		VariantID      uuid.UUID  `json:"variantId"`
		SKUCode        string     `json:"skuCode"`
		LocationID     *uuid.UUID `json:"locationId"`
		Qty            int64      `json:"qty"`
		IdempotencyKey string     `json:"idempotencyKey"`
		Reason         string     `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	bal, mov, err := h.Deps.Damage(r.Context(), app.DamageStockCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		SKUCode: body.SKUCode, LocationID: body.LocationID, Qty: body.Qty,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r), Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"balance": bal, "movement": mov})
}

func (h *Handler) waste(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID    uuid.UUID  `json:"warehouseId"`
		VariantID      uuid.UUID  `json:"variantId"`
		SKUCode        string     `json:"skuCode"`
		LocationID     *uuid.UUID `json:"locationId"`
		Qty            int64      `json:"qty"`
		IdempotencyKey string     `json:"idempotencyKey"`
		Reason         string     `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	bal, mov, err := h.Deps.Waste(r.Context(), app.WasteStockCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		SKUCode: body.SKUCode, LocationID: body.LocationID, Qty: body.Qty,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r), Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"balance": bal, "movement": mov})
}

func (h *Handler) listMovements(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListMovements(r.Context(), id, limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items, "total": total})
}

// --- Reservations ---

func (h *Handler) softReserve(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID    *uuid.UUID `json:"warehouseId"`
		ExternalRef    string     `json:"externalRef"`
		Priority       int        `json:"priority"`
		TTLSeconds     int        `json:"ttlSeconds"`
		IdempotencyKey string     `json:"idempotencyKey"`
		Lines          []struct {
			WarehouseID uuid.UUID  `json:"warehouseId"`
			VariantID   uuid.UUID  `json:"variantId"`
			SKUCode     string     `json:"skuCode"`
			LocationID  *uuid.UUID `json:"locationId"`
			Qty         int64      `json:"qty"`
			UseFEFO     bool       `json:"useFefo"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.SoftReserveLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.SoftReserveLine{
			WarehouseID: l.WarehouseID, VariantID: l.VariantID, SKUCode: l.SKUCode,
			LocationID: l.LocationID, Qty: l.Qty, UseFEFO: l.UseFEFO,
		})
	}
	var ttl time.Duration
	if body.TTLSeconds > 0 {
		ttl = time.Duration(body.TTLSeconds) * time.Second
	}
	out, err := h.Deps.SoftReserve(r.Context(), app.SoftReserveCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, ExternalRef: body.ExternalRef,
		Priority: body.Priority, TTL: ttl, Lines: lines,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) confirmHard(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.ConfirmHard(r.Context(), app.ConfirmHardCmd{
		TenantID: h.tenant(r), ReservationID: id, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) extend(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ExtendSeconds  int    `json:"extendSeconds"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	var ext time.Duration
	if body.ExtendSeconds > 0 {
		ext = time.Duration(body.ExtendSeconds) * time.Second
	}
	out, err := h.Deps.Extend(r.Context(), app.ExtendReservationCmd{
		TenantID: h.tenant(r), ReservationID: id, ExtendBy: ext,
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) release(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.Release(r.Context(), app.ReleaseReservationCmd{
		TenantID: h.tenant(r), ReservationID: id, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) consume(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.Consume(r.Context(), app.ConsumeReservationCmd{
		TenantID: h.tenant(r), ReservationID: id,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getReservation(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetReservation(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- ATP ---

func (h *Handler) queryATP(w http.ResponseWriter, r *http.Request) {
	q := domain.ATPQuery{TenantID: h.tenant(r), SKUCode: r.URL.Query().Get("skuCode"), Policy: domain.DefaultATPPolicy()}
	if v := r.URL.Query().Get("variantId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		q.VariantID = id
	}
	if v := r.URL.Query().Get("warehouseId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		q.WarehouseID = &id
	}
	if v := r.URL.Query().Get("regionId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		q.RegionID = &id
	}
	if r.URL.Query().Get("includeIncoming") == "true" {
		q.Policy.IncludeIncoming = true
	}
	items, err := h.Deps.QueryATP(r.Context(), q)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Transfers ---

func (h *Handler) createTransfer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code            string     `json:"code"`
		FromWarehouseID uuid.UUID  `json:"fromWarehouseId"`
		ToWarehouseID   uuid.UUID  `json:"toWarehouseId"`
		FromLocationID  *uuid.UUID `json:"fromLocationId"`
		ToLocationID    *uuid.UUID `json:"toLocationId"`
		Reason          string     `json:"reason"`
		Lines           []struct {
			VariantID uuid.UUID  `json:"variantId"`
			SKUCode   string     `json:"skuCode"`
			LotID     *uuid.UUID `json:"lotId"`
			Qty       int64      `json:"qty"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.CreateTransferLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.CreateTransferLine{
			VariantID: l.VariantID, SKUCode: l.SKUCode, LotID: l.LotID, Qty: l.Qty,
		})
	}
	out, err := h.Deps.CreateTransfer(r.Context(), app.CreateTransferCmd{
		TenantID: h.tenant(r), Code: body.Code,
		FromWarehouseID: body.FromWarehouseID, ToWarehouseID: body.ToWarehouseID,
		FromLocationID: body.FromLocationID, ToLocationID: body.ToLocationID,
		Reason: body.Reason, RequestedBy: parseActor(r), Lines: lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) approveTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ApproveTransfer(r.Context(), app.ApproveTransferCmd{
		TenantID: h.tenant(r), TransferID: id, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) completeTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.CompleteTransfer(r.Context(), app.CompleteTransferCmd{
		TenantID: h.tenant(r), TransferID: id,
		IdempotencyKey: idemKey(r, body.IdempotencyKey), ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getTransfer(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetTransfer(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Counts ---

func (h *Handler) startCount(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID              `json:"warehouseId"`
		LocationID  *uuid.UUID             `json:"locationId"`
		Type        domain.CountSessionType `json:"type"`
		Notes       string                 `json:"notes"`
		Lines       []struct {
			VariantID  uuid.UUID  `json:"variantId"`
			SKUCode    string     `json:"skuCode"`
			LocationID *uuid.UUID `json:"locationId"`
			LotID      *uuid.UUID `json:"lotId"`
			SystemQty  int64      `json:"systemQty"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.StartCountLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.StartCountLine{
			VariantID: l.VariantID, SKUCode: l.SKUCode, LocationID: l.LocationID,
			LotID: l.LotID, SystemQty: l.SystemQty,
		})
	}
	out, err := h.Deps.StartCount(r.Context(), app.StartCountCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, LocationID: body.LocationID,
		Type: body.Type, ActorID: parseActor(r), Notes: body.Notes, Lines: lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) submitCount(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Lines []struct {
			LineID     uuid.UUID `json:"lineId"`
			CountedQty int64     `json:"countedQty"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.SubmitCountLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.SubmitCountLine{LineID: l.LineID, CountedQty: l.CountedQty})
	}
	out, err := h.Deps.SubmitCount(r.Context(), app.SubmitCountCmd{
		TenantID: h.tenant(r), SessionID: id, ActorID: parseActor(r), Lines: lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) approveCount(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		IdempotencyKey string `json:"idempotencyKey"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.ApproveCount(r.Context(), app.ApproveCountCmd{
		TenantID: h.tenant(r), SessionID: id, ActorID: parseActor(r),
		IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getCount(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetCount(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Lots ---

func (h *Handler) nearExpiry(w http.ResponseWriter, r *http.Request) {
	days, _ := strconv.Atoi(r.URL.Query().Get("withinDays"))
	var wh *uuid.UUID
	if v := r.URL.Query().Get("warehouseId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		wh = &id
	}
	items, err := h.Deps.ListNearExpiry(r.Context(), h.tenant(r), wh, days)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) fefoAllocate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID `json:"warehouseId"`
		VariantID   uuid.UUID `json:"variantId"`
		Qty         int64     `json:"qty"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.AllocateFEFO(r.Context(), app.AllocateFEFOCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID, Qty: body.Qty,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"allocations": out})
}

// --- Returns ---

func (h *Handler) receiveReturn(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID    uuid.UUID               `json:"warehouseId"`
		Source         domain.ReturnSource     `json:"source"`
		Disposition    domain.ReturnDisposition `json:"disposition"`
		ExternalRef    string                  `json:"externalRef"`
		Reason         string                  `json:"reason"`
		IdempotencyKey string                  `json:"idempotencyKey"`
		Lines          []struct {
			VariantID   uuid.UUID                `json:"variantId"`
			SKUCode     string                   `json:"skuCode"`
			LotID       *uuid.UUID               `json:"lotId"`
			LocationID  *uuid.UUID               `json:"locationId"`
			Qty         int64                    `json:"qty"`
			Disposition *domain.ReturnDisposition `json:"disposition"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.ReceiveReturnLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.ReceiveReturnLine{
			VariantID: l.VariantID, SKUCode: l.SKUCode, LotID: l.LotID,
			LocationID: l.LocationID, Qty: l.Qty, Disposition: l.Disposition,
		})
	}
	out, err := h.Deps.ReceiveReturn(r.Context(), app.ReceiveReturnCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, Source: body.Source,
		Disposition: body.Disposition, ExternalRef: body.ExternalRef, ActorID: parseActor(r),
		Reason: body.Reason, Lines: lines, IdempotencyKey: idemKey(r, body.IdempotencyKey),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) getReturn(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetReturn(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Forecast ---

func (h *Handler) upsertForecast(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID     uuid.UUID `json:"warehouseId"`
		VariantID       uuid.UUID `json:"variantId"`
		SKUCode         string    `json:"skuCode"`
		HorizonStart    time.Time `json:"horizonStart"`
		HorizonEnd      time.Time `json:"horizonEnd"`
		PredictedDemand float64   `json:"predictedDemand"`
		PredictedATP    *float64  `json:"predictedAtp"`
		Confidence      *float64  `json:"confidence"`
		ModelID         string    `json:"modelId"`
		ModelVersion    string    `json:"modelVersion"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.UpsertForecast(r.Context(), app.UpsertForecastCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		SKUCode: body.SKUCode, HorizonStart: body.HorizonStart, HorizonEnd: body.HorizonEnd,
		PredictedDemand: body.PredictedDemand, PredictedATP: body.PredictedATP,
		Confidence: body.Confidence, ModelID: body.ModelID, ModelVersion: body.ModelVersion,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) generateForecast(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID `json:"warehouseId"`
		VariantID   uuid.UUID `json:"variantId"`
		HorizonDays int       `json:"horizonDays"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GenerateForecast(r.Context(), app.GenerateForecastCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, VariantID: body.VariantID,
		HorizonDays: body.HorizonDays,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getForecast(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetForecast(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) listForecasts(w http.ResponseWriter, r *http.Request) {
	whID, err := uuid.Parse(r.URL.Query().Get("warehouseId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	vid, err := uuid.Parse(r.URL.Query().Get("variantId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.ListForecasts(r.Context(), h.tenant(r), whID, vid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

// --- Search / admin ---

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	q := ports.SearchQuery{
		TenantID: h.tenant(r), Query: r.URL.Query().Get("q"),
		SKUCode: r.URL.Query().Get("skuCode"), Limit: limit, Offset: offset,
	}
	if v := r.URL.Query().Get("warehouseId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		q.WarehouseID = &id
	}
	if v := r.URL.Query().Get("variantId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		q.VariantID = &id
	}
	out, err := h.Deps.SearchStock(r.Context(), q)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) reindex(w http.ResponseWriter, r *http.Request) {
	if err := h.Deps.ReindexStock(r.Context(), h.tenant(r)); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "ok"})
}

func (h *Handler) explorer(w http.ResponseWriter, r *http.Request) {
	writeOK(w, map[string]any{
		"service": "inventory-service",
		"areas": []string{
			"warehouses", "locations", "stock", "reservations", "atp",
			"transfers", "counts", "lots", "returns", "forecasts", "search",
		},
	})
}
