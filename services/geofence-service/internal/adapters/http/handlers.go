package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/geofence-service/internal/app"
	"github.com/nexora/geofence-service/internal/domain"
	"github.com/nexora/geofence-service/internal/ratelimit"
)

// Handler serves geofence REST endpoints.
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
	const base = "/v1/geofence"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/zones", tenant(h.createZone))
	mux.HandleFunc("GET "+base+"/zones", tenant(h.listZones))
	mux.HandleFunc("GET "+base+"/zones/{id}", tenant(h.getZone))
	mux.HandleFunc("PUT "+base+"/zones/{id}", tenant(h.updateZone))
	mux.HandleFunc("DELETE "+base+"/zones/{id}", tenant(h.deleteZone))
	mux.HandleFunc("POST "+base+"/contains", tenant(h.contains))
	mux.HandleFunc("POST "+base+"/serviceability", tenant(h.serviceability))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.publishOutbox))

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

func (h *Handler) tenantID(r *http.Request) uuid.UUID {
	tid, _ := TenantIDFromContext(r.Context())
	return tid
}

func zoneDTO(z domain.Zone) map[string]any {
	return map[string]any{
		"id": z.ID, "tenantId": z.TenantID, "name": z.Name, "city": z.City,
		"kind": z.Kind, "vertices": z.Vertices,
		"centerLat": z.CenterLat, "centerLng": z.CenterLng, "radiusM": z.RadiusM,
		"active": z.Active, "createdAt": z.CreatedAt, "updatedAt": z.UpdatedAt,
	}
}

type pointBody struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type createZoneBody struct {
	Name      string       `json:"name"`
	City      string       `json:"city"`
	Kind      string       `json:"kind"`
	Vertices  []pointBody  `json:"vertices"`
	CenterLat *float64     `json:"centerLat"`
	CenterLng *float64     `json:"centerLng"`
	RadiusM   *float64     `json:"radiusM"`
	Active    *bool        `json:"active"`
}

func (h *Handler) createZone(w http.ResponseWriter, r *http.Request) {
	var body createZoneBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	verts := make([]domain.Point, 0, len(body.Vertices))
	for _, v := range body.Vertices {
		verts = append(verts, domain.Point{Lat: v.Lat, Lng: v.Lng})
	}
	z, err := h.Deps.CreateZone(r.Context(), app.CreateZoneInput{
		TenantID: h.tenantID(r), Name: body.Name, City: body.City,
		Kind: domain.ZoneKind(body.Kind), Vertices: verts,
		CenterLat: body.CenterLat, CenterLng: body.CenterLng, RadiusM: body.RadiusM,
		Active: body.Active,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, zoneDTO(z))
}

func (h *Handler) listZones(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListZones(r.Context(), app.ListZonesInput{
		TenantID: h.tenantID(r),
		City:     r.URL.Query().Get("city"),
		Kind:     domain.ZoneKind(r.URL.Query().Get("kind")),
		Limit:    limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, z := range items {
		out = append(out, zoneDTO(z))
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) getZone(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	z, err := h.Deps.GetZone(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, zoneDTO(z))
}

func (h *Handler) updateZone(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body createZoneBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	kind := domain.ZoneKind(body.Kind)
	verts := make([]domain.Point, 0, len(body.Vertices))
	for _, v := range body.Vertices {
		verts = append(verts, domain.Point{Lat: v.Lat, Lng: v.Lng})
	}
	in := app.UpdateZoneInput{
		TenantID: h.tenantID(r), ID: id,
		Name: &body.Name, City: &body.City, Kind: &kind,
		Vertices: verts, CenterLat: body.CenterLat, CenterLng: body.CenterLng,
		RadiusM: body.RadiusM, Active: body.Active, SetGeo: true,
	}
	z, err := h.Deps.UpdateZone(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, zoneDTO(z))
}

func (h *Handler) deleteZone(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if err := h.Deps.DeleteZone(r.Context(), h.tenantID(r), id); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "deleted"})
}

func (h *Handler) contains(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ZoneID string    `json:"zoneId"`
		Point  pointBody `json:"point"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	zid, err := uuid.Parse(body.ZoneID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.Contains(r.Context(), app.ContainsInput{
		TenantID: h.tenantID(r), ZoneID: zid,
		Point: domain.Point{Lat: body.Point.Lat, Lng: body.Point.Lng},
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"inside": res.Inside, "zoneId": res.ZoneID, "kind": res.Kind})
}

func (h *Handler) serviceability(w http.ResponseWriter, r *http.Request) {
	var body struct {
		City  string    `json:"city"`
		Point pointBody `json:"point"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.CheckServiceability(r.Context(), app.ServiceabilityInput{
		TenantID: h.tenantID(r), City: body.City,
		Point: domain.Point{Lat: body.Point.Lat, Lng: body.Point.Lng},
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"serviceable": res.Serviceable, "blocked": res.Blocked, "reason": res.Reason,
		"matchingZones": res.MatchingZones, "restrictedZones": res.RestrictedZones,
		"deliveryZones": res.DeliveryZones,
	})
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
