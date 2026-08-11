package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/tracking-service/internal/app"
	"github.com/nexora/tracking-service/internal/domain"
	"github.com/nexora/tracking-service/internal/ratelimit"
)

// Handler serves tracking REST endpoints.
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
	const base = "/v1/tracking"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/locations", tenant(h.ingestLocation))
	mux.HandleFunc("GET "+base+"/couriers/{id}/live", tenant(h.getLiveCourier))
	mux.HandleFunc("GET "+base+"/orders/{orderId}/timeline", tenant(h.getOrderTimeline))
	mux.HandleFunc("POST "+base+"/orders/{orderId}/timeline", tenant(h.appendTimeline))
	mux.HandleFunc("POST "+base+"/arrival/detect", tenant(h.detectArrival))
	mux.HandleFunc("GET "+base+"/nearby", tenant(h.listNearby))
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

type ingestBody struct {
	CourierID  string   `json:"courierId"`
	Lat        float64  `json:"lat"`
	Lon        float64  `json:"lon"`
	AccuracyM  float64  `json:"accuracyM"`
	HeadingDeg *float64 `json:"headingDeg"`
	SpeedMPS   *float64 `json:"speedMps"`
	RecordedAt *string  `json:"recordedAt"`
	OrderID    *string  `json:"orderId"`
}

func (h *Handler) ingestLocation(w http.ResponseWriter, r *http.Request) {
	var body ingestBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	courierID, err := uuid.Parse(body.CourierID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.IngestLocationInput{
		TenantID: h.tenantID(r), CourierID: courierID,
		Lat: body.Lat, Lon: body.Lon, AccuracyM: body.AccuracyM,
		HeadingDeg: body.HeadingDeg, SpeedMPS: body.SpeedMPS,
	}
	if body.RecordedAt != nil {
		t, err := time.Parse(time.RFC3339, *body.RecordedAt)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.RecordedAt = &t
	}
	if body.OrderID != nil && *body.OrderID != "" {
		oid, err := uuid.Parse(*body.OrderID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.OrderID = &oid
	}
	loc, err := h.Deps.IngestLocation(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, locationDTO(loc))
}

func (h *Handler) getLiveCourier(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	loc, err := h.Deps.GetLiveCourier(r.Context(), app.GetLiveCourierInput{
		TenantID: h.tenantID(r), CourierID: id,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, locationDTO(loc))
}

func (h *Handler) getOrderTimeline(w http.ResponseWriter, r *http.Request) {
	oid, err := uuid.Parse(r.PathValue("orderId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit := 100
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.Deps.GetOrderTimeline(r.Context(), app.GetOrderTimelineInput{
		TenantID: h.tenantID(r), OrderID: oid, Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, e := range list {
		items = append(items, timelineDTO(e))
	}
	writeOK(w, map[string]any{"items": items})
}

type appendTimelineBody struct {
	CourierID  *string                `json:"courierId"`
	Type       string                 `json:"type"`
	Lat        *float64               `json:"lat"`
	Lon        *float64               `json:"lon"`
	Message    string                 `json:"message"`
	Meta       map[string]any         `json:"meta"`
	OccurredAt *string                `json:"occurredAt"`
}

func (h *Handler) appendTimeline(w http.ResponseWriter, r *http.Request) {
	oid, err := uuid.Parse(r.PathValue("orderId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body appendTimelineBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.AppendTimelineInput{
		TenantID: h.tenantID(r), OrderID: oid,
		Type: domain.TimelineEventType(body.Type), Lat: body.Lat, Lon: body.Lon,
		Message: body.Message, Meta: body.Meta,
	}
	if body.CourierID != nil && *body.CourierID != "" {
		cid, err := uuid.Parse(*body.CourierID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.CourierID = &cid
	}
	if body.OccurredAt != nil {
		t, err := time.Parse(time.RFC3339, *body.OccurredAt)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.OccurredAt = &t
	}
	ev, err := h.Deps.AppendTimeline(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, timelineDTO(ev))
}

type detectArrivalBody struct {
	CourierID  string  `json:"courierId"`
	OrderID    string  `json:"orderId"`
	DropoffLat float64 `json:"dropoffLat"`
	DropoffLon float64 `json:"dropoffLon"`
	ThresholdM float64 `json:"thresholdM"`
}

func (h *Handler) detectArrival(w http.ResponseWriter, r *http.Request) {
	var body detectArrivalBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	courierID, err := uuid.Parse(body.CourierID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	orderID, err := uuid.Parse(body.OrderID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.DetectArrival(r.Context(), app.DetectArrivalInput{
		TenantID: h.tenantID(r), CourierID: courierID, OrderID: orderID,
		DropoffLat: body.DropoffLat, DropoffLon: body.DropoffLon, ThresholdM: body.ThresholdM,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"arrived": res.Arrived, "distanceMeters": res.DistanceMeters,
		"thresholdMeters": res.ThresholdMeters, "courierId": res.CourierID,
		"orderId": res.OrderID, "timelineEventId": res.TimelineEventID,
	})
}

func (h *Handler) listNearby(w http.ResponseWriter, r *http.Request) {
	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query().Get("lon"), 64)
	if err1 != nil || err2 != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	radius := 1000.0
	if v := r.URL.Query().Get("radiusM"); v != "" {
		if n, err := strconv.ParseFloat(v, 64); err == nil {
			radius = n
		}
	}
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	list, err := h.Deps.ListNearby(r.Context(), app.ListNearbyInput{
		TenantID: h.tenantID(r), Lat: lat, Lon: lon, RadiusM: radius, Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(list))
	for _, loc := range list {
		items = append(items, locationDTO(loc))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

func locationDTO(loc domain.CourierLocation) map[string]any {
	return map[string]any{
		"tenantId": loc.TenantID, "courierId": loc.CourierID,
		"lat": loc.Lat, "lon": loc.Lon, "accuracyM": loc.AccuracyM,
		"headingDeg": loc.HeadingDeg, "speedMps": loc.SpeedMPS,
		"recordedAt": loc.RecordedAt, "updatedAt": loc.UpdatedAt,
	}
}

func timelineDTO(e domain.TimelineEvent) map[string]any {
	return map[string]any{
		"id": e.ID, "tenantId": e.TenantID, "orderId": e.OrderID,
		"courierId": e.CourierID, "type": e.Type, "lat": e.Lat, "lon": e.Lon,
		"message": e.Message, "meta": e.Meta,
		"occurredAt": e.OccurredAt, "createdAt": e.CreatedAt,
	}
}
