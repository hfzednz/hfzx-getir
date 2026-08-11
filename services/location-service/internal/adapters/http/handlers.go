package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/app"
	"github.com/nexora/location-service/internal/domain"
	"github.com/nexora/location-service/internal/ratelimit"
)

// Handler serves location REST endpoints.
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
	const base = "/v1/location"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/geocode", tenant(h.forwardGeocode))
	mux.HandleFunc("POST "+base+"/geocode/reverse", tenant(h.reverseGeocode))
	mux.HandleFunc("POST "+base+"/geocode/autocomplete", tenant(h.autocomplete))

	mux.HandleFunc("POST "+base+"/addresses/validate", tenant(h.validateAddress))
	mux.HandleFunc("POST "+base+"/addresses/normalize", tenant(h.normalizeAddress))
	mux.HandleFunc("POST "+base+"/addresses/enrich", tenant(h.enrichAddress))

	mux.HandleFunc("POST "+base+"/pois", tenant(h.upsertPOI))
	mux.HandleFunc("POST "+base+"/spatial/nearby", tenant(h.nearby))
	mux.HandleFunc("POST "+base+"/spatial/radius", tenant(h.radiusSearch))
	mux.HandleFunc("POST "+base+"/spatial/bbox", tenant(h.bboxSearch))
	mux.HandleFunc("POST "+base+"/spatial/nearest", tenant(h.nearest))

	mux.HandleFunc("POST "+base+"/routes", tenant(h.proxyRoute))
	mux.HandleFunc("POST "+base+"/eta", tenant(h.proxyETA))
	mux.HandleFunc("POST "+base+"/zones/serviceability", tenant(h.zoneServiceability))

	mux.HandleFunc("POST "+base+"/history", tenant(h.ingestHistory))
	mux.HandleFunc("GET "+base+"/history", tenant(h.getHistory))

	mux.HandleFunc("GET "+base+"/maps/offline/{region}", tenant(h.getOfflineManifest))
	mux.HandleFunc("POST "+base+"/maps/offline", tenant(h.upsertOfflineManifest))

	mux.HandleFunc("POST "+base+"/ai/heat", tenant(h.upsertHeat))
	mux.HandleFunc("GET "+base+"/ai/heatmap", tenant(h.demandHeatmap))
	mux.HandleFunc("GET "+base+"/admin/coverage", tenant(h.coverageStats))

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

func (h *Handler) forwardGeocode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query string `json:"query"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ForwardGeocode(r.Context(), app.ForwardGeocodeInput{
		TenantID: h.tenantID(r), Query: body.Query,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, geocodeDTO(res))
}

func (h *Handler) reverseGeocode(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ReverseGeocode(r.Context(), app.ReverseGeocodeInput{
		TenantID: h.tenantID(r), Lat: body.Lat, Lng: body.Lng,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, geocodeDTO(res))
}

func (h *Handler) autocomplete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.Autocomplete(r.Context(), app.AutocompleteInput{
		TenantID: h.tenantID(r), Query: body.Query, Limit: body.Limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(res))
	for _, g := range res {
		items = append(items, geocodeDTO(g))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) validateAddress(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Line1   string  `json:"line1"`
		Lat     float64 `json:"lat"`
		Lng     float64 `json:"lng"`
		PlaceID string  `json:"placeId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ValidateAddress(r.Context(), app.ValidateAddressInput{
		TenantID: h.tenantID(r), Line1: body.Line1, Lat: body.Lat, Lng: body.Lng, PlaceID: body.PlaceID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"address": addressDTO(res.Address), "feasibility": res.Feasibility,
	})
}

func (h *Handler) normalizeAddress(w http.ResponseWriter, r *http.Request) {
	var body addressBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.NormalizeAddress(r.Context(), app.NormalizeAddressInput{
		TenantID: h.tenantID(r), Address: body.toDomain(),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, addressDTO(res))
}

func (h *Handler) enrichAddress(w http.ResponseWriter, r *http.Request) {
	var body addressBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.EnrichAddress(r.Context(), app.EnrichAddressInput{
		TenantID: h.tenantID(r), Address: body.toDomain(),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, addressDTO(res))
}

func (h *Handler) upsertPOI(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID     *string        `json:"id"`
		Kind   string         `json:"kind"`
		RefID  string         `json:"refId"`
		Name   string         `json:"name"`
		Lat    float64        `json:"lat"`
		Lng    float64        `json:"lng"`
		Meta   map[string]any `json:"meta"`
		Active *bool          `json:"active"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	in := app.UpsertPOIInput{
		TenantID: h.tenantID(r), Kind: domain.POIKind(body.Kind),
		RefID: body.RefID, Name: body.Name, Lat: body.Lat, Lng: body.Lng,
		Meta: body.Meta, Active: body.Active,
	}
	if body.ID != nil && *body.ID != "" {
		id, err := uuid.Parse(*body.ID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		in.ID = id
	}
	p, err := h.Deps.UpsertPOI(r.Context(), in)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, poiDTO(p))
}

func (h *Handler) nearby(w http.ResponseWriter, r *http.Request) {
	var body spatialBody
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var kind *domain.POIKind
	if body.Kind != "" {
		k := domain.POIKind(body.Kind)
		kind = &k
	}
	hits, err := h.Deps.Nearby(r.Context(), app.NearbyInput{
		TenantID: h.tenantID(r), Lat: body.Lat, Lng: body.Lng,
		RadiusM: body.RadiusM, Kind: kind, Limit: body.Limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": poiListDTO(hits)})
}

func (h *Handler) radiusSearch(w http.ResponseWriter, r *http.Request) {
	h.nearby(w, r)
}

func (h *Handler) bboxSearch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MinLat float64 `json:"minLat"`
		MinLng float64 `json:"minLng"`
		MaxLat float64 `json:"maxLat"`
		MaxLng float64 `json:"maxLng"`
		Kind   string  `json:"kind"`
		Limit  int     `json:"limit"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var kind *domain.POIKind
	if body.Kind != "" {
		k := domain.POIKind(body.Kind)
		kind = &k
	}
	hits, err := h.Deps.BBoxSearch(r.Context(), app.BBoxSearchInput{
		TenantID: h.tenantID(r),
		BBox: domain.BBox{
			MinLat: body.MinLat, MinLng: body.MinLng,
			MaxLat: body.MaxLat, MaxLng: body.MaxLng,
		},
		Kind: kind, Limit: body.Limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": poiListDTO(hits)})
}

func (h *Handler) nearest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Kind  string  `json:"kind"`
		Lat   float64 `json:"lat"`
		Lng   float64 `json:"lng"`
		Limit int     `json:"limit"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	hits, err := h.Deps.NearestOfKind(r.Context(), app.NearestOfKindInput{
		TenantID: h.tenantID(r), Kind: domain.POIKind(body.Kind),
		Lat: body.Lat, Lng: body.Lng, Limit: body.Limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": poiListDTO(hits)})
}

func (h *Handler) proxyRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Origin    latLngBody   `json:"origin"`
		Dest      latLngBody   `json:"dest"`
		Waypoints []latLngBody `json:"waypoints"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	wps := make([]domain.LatLng, 0, len(body.Waypoints))
	for _, wp := range body.Waypoints {
		wps = append(wps, domain.LatLng{Lat: wp.Lat, Lng: wp.Lng})
	}
	res, err := h.Deps.ProxyRoute(r.Context(), app.ProxyRouteInput{
		TenantID: h.tenantID(r),
		Origin:   domain.LatLng{Lat: body.Origin.Lat, Lng: body.Origin.Lng},
		Dest:     domain.LatLng{Lat: body.Dest.Lat, Lng: body.Dest.Lng},
		Waypoints: wps,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"routeId": res.RouteID, "distanceMeters": res.DistanceMeters,
		"durationSeconds": res.DurationSeconds, "provider": res.Provider,
	})
}

func (h *Handler) proxyETA(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Origin latLngBody `json:"origin"`
		Dest   latLngBody `json:"dest"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ProxyETA(r.Context(), app.ProxyETAInput{
		TenantID: h.tenantID(r),
		Origin:   domain.LatLng{Lat: body.Origin.Lat, Lng: body.Origin.Lng},
		Dest:     domain.LatLng{Lat: body.Dest.Lat, Lng: body.Dest.Lng},
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"distanceMeters": res.DistanceMeters, "durationSeconds": res.DurationSeconds,
		"provider": res.Provider,
	})
}

func (h *Handler) zoneServiceability(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lng"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.CheckZoneServiceability(r.Context(), app.CheckZoneServiceabilityInput{
		TenantID: h.tenantID(r), Lat: body.Lat, Lng: body.Lng,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) ingestHistory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		SubjectType string  `json:"subjectType"`
		SubjectID   string  `json:"subjectId"`
		Lat         float64 `json:"lat"`
		Lng         float64 `json:"lng"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.IngestHistory(r.Context(), app.IngestHistoryInput{
		TenantID: h.tenantID(r), SubjectType: domain.SubjectType(body.SubjectType),
		SubjectID: body.SubjectID, Lat: body.Lat, Lng: body.Lng,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, historyDTO(res))
}

func (h *Handler) getHistory(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	res, err := h.Deps.GetHistory(r.Context(), app.GetHistoryInput{
		TenantID: h.tenantID(r), SubjectType: domain.SubjectType(q.Get("subjectType")),
		SubjectID: q.Get("subjectId"), Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(res))
	for _, row := range res {
		items = append(items, historyDTO(row))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) getOfflineManifest(w http.ResponseWriter, r *http.Request) {
	res, err := h.Deps.GetOfflineManifest(r.Context(), app.GetOfflineManifestInput{
		TenantID: h.tenantID(r), Region: r.PathValue("region"),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, manifestDTO(res))
}

func (h *Handler) upsertOfflineManifest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Region    string `json:"region"`
		Version   string `json:"version"`
		URL       string `json:"url"`
		SizeBytes int64  `json:"sizeBytes"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.UpsertOfflineManifest(r.Context(), app.UpsertOfflineManifestInput{
		TenantID: h.tenantID(r), Region: body.Region, Version: body.Version,
		URL: body.URL, SizeBytes: body.SizeBytes,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, manifestDTO(res))
}

func (h *Handler) upsertHeat(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GridCell    string  `json:"gridCell"`
		DemandScore float64 `json:"demandScore"`
		Density     float64 `json:"density"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.UpsertHeatCell(r.Context(), app.UpsertHeatCellInput{
		TenantID: h.tenantID(r), GridCell: body.GridCell,
		DemandScore: body.DemandScore, Density: body.Density,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, heatDTO(res))
}

func (h *Handler) demandHeatmap(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	cells, err := h.Deps.DemandHeatmap(r.Context(), app.DemandHeatmapInput{
		TenantID: h.tenantID(r), Limit: limit,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items := make([]map[string]any, 0, len(cells))
	for _, c := range cells {
		items = append(items, heatDTO(c))
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) coverageStats(w http.ResponseWriter, r *http.Request) {
	res, err := h.Deps.AdminCoverageStats(r.Context(), app.AdminCoverageStatsInput{
		TenantID: h.tenantID(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) publishOutbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}

type latLngBody struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type spatialBody struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	RadiusM float64 `json:"radiusM"`
	Kind    string  `json:"kind"`
	Limit   int     `json:"limit"`
}

type addressBody struct {
	ID         *string `json:"id"`
	Line1      string  `json:"line1"`
	Building   string  `json:"building"`
	Entrance   string  `json:"entrance"`
	Floor      string  `json:"floor"`
	Apt        string  `json:"apt"`
	Landmark   string  `json:"landmark"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
	PlaceID    string  `json:"placeId"`
	Confidence float64 `json:"confidence"`
	RiskScore  float64 `json:"riskScore"`
	Components struct {
		Country      string `json:"country"`
		City         string `json:"city"`
		District     string `json:"district"`
		Neighborhood string `json:"neighborhood"`
		Street       string `json:"street"`
		PostalCode   string `json:"postalCode"`
	} `json:"components"`
}

func (b addressBody) toDomain() domain.NormalizedAddress {
	a := domain.NormalizedAddress{
		Line1: b.Line1, Building: b.Building, Entrance: b.Entrance,
		Floor: b.Floor, Apt: b.Apt, Landmark: b.Landmark,
		Lat: b.Lat, Lng: b.Lng, PlaceID: b.PlaceID,
		Confidence: domain.ConfidenceScore(b.Confidence), RiskScore: b.RiskScore,
		Components: domain.AddressComponents{
			Country: b.Components.Country, City: b.Components.City,
			District: b.Components.District, Neighborhood: b.Components.Neighborhood,
			Street: b.Components.Street, PostalCode: b.Components.PostalCode,
		},
	}
	if b.ID != nil {
		if id, err := uuid.Parse(*b.ID); err == nil {
			a.ID = id
		}
	}
	return a
}

func geocodeDTO(g domain.GeocodeResult) map[string]any {
	return map[string]any{
		"placeId": g.PlaceID, "formatted": g.Formatted,
		"lat": g.Lat, "lng": g.Lng, "confidence": g.Confidence,
		"components": g.Components, "provider": g.Provider, "cached": g.Cached,
	}
}

func addressDTO(a domain.NormalizedAddress) map[string]any {
	return map[string]any{
		"id": a.ID.String(), "tenantId": a.TenantID.String(),
		"line1": a.Line1, "building": a.Building, "entrance": a.Entrance,
		"floor": a.Floor, "apt": a.Apt, "landmark": a.Landmark,
		"lat": a.Lat, "lng": a.Lng, "placeId": a.PlaceID,
		"confidence": a.Confidence, "riskScore": a.RiskScore,
		"components": a.Components,
	}
}

func poiDTO(p domain.POI) map[string]any {
	return map[string]any{
		"id": p.ID.String(), "tenantId": p.TenantID.String(),
		"kind": string(p.Kind), "refId": p.RefID, "name": p.Name,
		"lat": p.Lat, "lng": p.Lng, "meta": p.Meta, "active": p.Active,
	}
}

func poiListDTO(items []domain.POI) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, poiDTO(p))
	}
	return out
}

func historyDTO(h domain.LocationHistory) map[string]any {
	return map[string]any{
		"id": h.ID.String(), "tenantId": h.TenantID.String(),
		"subjectType": string(h.SubjectType), "subjectId": h.SubjectID,
		"lat": h.Lat, "lng": h.Lng, "recordedAt": h.RecordedAt,
	}
}

func manifestDTO(m domain.OfflineManifest) map[string]any {
	return map[string]any{
		"id": m.ID.String(), "tenantId": m.TenantID.String(),
		"region": m.Region, "version": m.Version, "url": m.URL, "sizeBytes": m.SizeBytes,
	}
}

func heatDTO(c domain.HeatCell) map[string]any {
	return map[string]any{
		"id": c.ID.String(), "tenantId": c.TenantID.String(),
		"gridCell": c.GridCell, "demandScore": c.DemandScore, "density": c.Density,
		"updatedAt": c.UpdatedAt,
	}
}
