package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/routing-service/internal/app"
	"github.com/nexora/routing-service/internal/domain"
	"github.com/nexora/routing-service/internal/ratelimit"
)

// Handler serves routing REST endpoints.
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
	const base = "/v1/routing"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/routes", tenant(h.createRoute))
	mux.HandleFunc("GET "+base+"/routes/{id}", tenant(h.getRoute))
	mux.HandleFunc("POST "+base+"/routes/{id}/optimize", tenant(h.optimize))
	mux.HandleFunc("POST "+base+"/routes/{id}/recalculate-eta", tenant(h.recalculateETA))
	mux.HandleFunc("POST "+base+"/traffic-hints", tenant(h.updateTrafficHint))
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

type waypointBody struct {
	Kind    string  `json:"kind"`
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	OrderID *string `json:"orderId"`
	Label   string  `json:"label"`
}

type createRouteBody struct {
	DispatchID  *string        `json:"dispatchId"`
	CourierID   *string        `json:"courierId"`
	WarehouseID *string        `json:"warehouseId"`
	SpeedMPS    float64        `json:"speedMps"`
	Waypoints   []waypointBody `json:"waypoints"`
}

func (h *Handler) createRoute(w http.ResponseWriter, r *http.Request) {
	var body createRouteBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	wps := make([]app.WaypointInput, 0, len(body.Waypoints))
	for _, wp := range body.Waypoints {
		in := app.WaypointInput{
			Kind: domain.WaypointKind(wp.Kind), Lat: wp.Lat, Lon: wp.Lon, Label: wp.Label,
		}
		if wp.OrderID != nil && *wp.OrderID != "" {
			oid, err := uuid.Parse(*wp.OrderID)
			if err != nil {
				writeErr(w, r, domain.ErrInvalidArgument)
				return
			}
			in.OrderID = &oid
		}
		wps = append(wps, in)
	}
	in := app.CreateRouteInput{
		TenantID: h.tenantID(r), Waypoints: wps, SpeedMPS: body.SpeedMPS,
	}
	if body.DispatchID != nil {
		id, err := uuid.Parse(*body.DispatchID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.DispatchID = &id
	}
	if body.CourierID != nil {
		id, err := uuid.Parse(*body.CourierID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.CourierID = &id
	}
	if body.WarehouseID != nil {
		id, err := uuid.Parse(*body.WarehouseID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.WarehouseID = &id
	}
	route, err := h.Deps.CreateRoute(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, routeDTO(route))
}

func (h *Handler) getRoute(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	route, err := h.Deps.GetRoute(r.Context(), app.GetRouteInput{
		TenantID: h.tenantID(r), RouteID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, routeDTO(route))
}

func (h *Handler) optimize(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	route, err := h.Deps.Optimize(r.Context(), app.OptimizeInput{
		TenantID: h.tenantID(r), RouteID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, routeDTO(route))
}

type recalcBody struct {
	CurrentLat *float64 `json:"currentLat"`
	CurrentLon *float64 `json:"currentLon"`
	Reason     string   `json:"reason"`
}

func (h *Handler) recalculateETA(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body recalcBody
	if r.ContentLength != 0 {
		if err := decodeJSON(r, &body); err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
	}
	route, err := h.Deps.RecalculateETA(r.Context(), app.RecalculateETAInput{
		TenantID: h.tenantID(r), RouteID: id,
		CurrentLat: body.CurrentLat, CurrentLon: body.CurrentLon, Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, routeDTO(route))
}

type trafficHintBody struct {
	HintID     *string  `json:"hintId"`
	RegionKey  string   `json:"regionKey"`
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	RadiusM    float64  `json:"radiusM"`
	Factor     float64  `json:"factor"`
	ValidFrom  *string  `json:"validFrom"`
	ValidUntil *string  `json:"validUntil"`
}

func (h *Handler) updateTrafficHint(w http.ResponseWriter, r *http.Request) {
	var body trafficHintBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.UpdateTrafficHintInput{
		TenantID: h.tenantID(r), RegionKey: body.RegionKey,
		Lat: body.Lat, Lon: body.Lon, RadiusM: body.RadiusM, Factor: body.Factor,
	}
	if body.HintID != nil && *body.HintID != "" {
		id, err := uuid.Parse(*body.HintID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.HintID = &id
	}
	if body.ValidFrom != nil {
		t, err := time.Parse(time.RFC3339, *body.ValidFrom)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.ValidFrom = &t
	}
	if body.ValidUntil != nil {
		t, err := time.Parse(time.RFC3339, *body.ValidUntil)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.ValidUntil = &t
	}
	hint, err := h.Deps.UpdateTrafficHint(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"id": hint.ID, "tenantId": hint.TenantID, "regionKey": hint.RegionKey,
		"lat": hint.Lat, "lon": hint.Lon, "radiusM": hint.RadiusM, "factor": hint.Factor,
		"validFrom": hint.ValidFrom, "validUntil": hint.ValidUntil,
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

func routeDTO(r domain.Route) map[string]any {
	wps := make([]map[string]any, 0, len(r.Waypoints))
	for _, w := range r.Waypoints {
		wps = append(wps, map[string]any{
			"id": w.ID, "sequence": w.Sequence, "kind": w.Kind,
			"lat": w.Lat, "lon": w.Lon, "orderId": w.OrderID, "label": w.Label, "etaAt": w.ETAAt,
		})
	}
	legs := make([]map[string]any, 0, len(r.Legs))
	for _, l := range r.Legs {
		legs = append(legs, map[string]any{
			"id": l.ID, "fromSequence": l.FromSequence, "toSequence": l.ToSequence,
			"distanceMeters": l.DistanceMeters, "durationSeconds": l.DurationSeconds,
		})
	}
	return map[string]any{
		"id": r.ID, "tenantId": r.TenantID, "dispatchId": r.DispatchID,
		"courierId": r.CourierID, "warehouseId": r.WarehouseID, "status": r.Status,
		"waypoints": wps, "legs": legs,
		"distanceMeters": r.DistanceMeters, "durationSeconds": r.DurationSeconds,
		"etaAt": r.ETAAt, "trafficFactor": r.TrafficFactor, "weatherFactor": r.WeatherFactor,
		"speedMps": r.SpeedMPS, "createdAt": r.CreatedAt, "updatedAt": r.UpdatedAt,
	}
}
