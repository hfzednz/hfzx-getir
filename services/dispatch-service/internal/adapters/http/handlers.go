package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/dispatch-service/internal/app"
	"github.com/nexora/dispatch-service/internal/domain"
	"github.com/nexora/dispatch-service/internal/ratelimit"
)

// Handler serves dispatch REST endpoints.
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
	const base = "/v1/dispatch"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	mux.HandleFunc("POST "+base+"/jobs", tenant(h.createDispatch))
	mux.HandleFunc("GET "+base+"/jobs", tenant(h.adminList))
	mux.HandleFunc("GET "+base+"/jobs/{id}", tenant(h.getDispatch))
	mux.HandleFunc("POST "+base+"/jobs/{id}/auto-assign", tenant(h.autoAssign))
	mux.HandleFunc("POST "+base+"/jobs/{id}/assign", tenant(h.manualAssign))
	mux.HandleFunc("POST "+base+"/jobs/{id}/reassign", tenant(h.reassign))
	mux.HandleFunc("POST "+base+"/jobs/{id}/pickup/start", tenant(h.startPickup))
	mux.HandleFunc("POST "+base+"/jobs/{id}/pickup/complete", tenant(h.completePickup))
	mux.HandleFunc("POST "+base+"/jobs/{id}/transit/start", tenant(h.startTransit))
	mux.HandleFunc("POST "+base+"/jobs/{id}/arrive", tenant(h.arrive))
	mux.HandleFunc("POST "+base+"/jobs/{id}/complete", tenant(h.completeDelivery))
	mux.HandleFunc("POST "+base+"/jobs/{id}/fail", tenant(h.failDelivery))
	mux.HandleFunc("POST "+base+"/batches", tenant(h.batchCreate))
	mux.HandleFunc("POST "+base+"/fleet/vehicles", tenant(h.upsertVehicle))
	mux.HandleFunc("GET "+base+"/fleet/vehicles", tenant(h.listVehicles))
	mux.HandleFunc("POST "+base+"/couriers/snapshot", tenant(h.upsertCourier))
	mux.HandleFunc("GET "+base+"/admin", tenant(h.adminList))
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

func dispatchDTO(d domain.Dispatch) map[string]any {
	return map[string]any{
		"id": d.ID, "tenantId": d.TenantID, "orderId": d.OrderID,
		"fulfillmentId": d.FulfillmentID, "warehouseId": d.WarehouseID,
		"courierPrincipalId": d.CourierPrincipalID, "vehicleId": d.VehicleID,
		"status": d.Status, "pickup": d.Pickup, "dropoff": d.Dropoff,
		"requiredVehicle": d.RequiredVehicle, "batchId": d.BatchID,
		"routeId": d.RouteID, "etaSeconds": d.ETASeconds,
		"podType": d.PODType, "podReference": d.PODReference,
		"failReason": d.FailReason, "failNote": d.FailNote,
		"createdAt": d.CreatedAt, "updatedAt": d.UpdatedAt,
	}
}

type pointBody struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (h *Handler) createDispatch(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OrderID         string    `json:"orderId"`
		FulfillmentID   string    `json:"fulfillmentId"`
		WarehouseID     string    `json:"warehouseId"`
		Pickup          pointBody `json:"pickup"`
		Dropoff         pointBody `json:"dropoff"`
		RequiredVehicle string    `json:"requiredVehicle"`
		City            string    `json:"city"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	orderID, err := uuid.Parse(body.OrderID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var fulfillmentID, warehouseID uuid.UUID
	if body.FulfillmentID != "" {
		fulfillmentID, _ = uuid.Parse(body.FulfillmentID)
	}
	if body.WarehouseID != "" {
		warehouseID, _ = uuid.Parse(body.WarehouseID)
	}
	d, err := h.Deps.CreateDispatch(r.Context(), app.CreateDispatchInput{
		TenantID: h.tenantID(r), OrderID: orderID, FulfillmentID: fulfillmentID,
		WarehouseID: warehouseID,
		Pickup: domain.Point{Lat: body.Pickup.Lat, Lng: body.Pickup.Lng},
		Dropoff: domain.Point{Lat: body.Dropoff.Lat, Lng: body.Dropoff.Lng},
		RequiredVehicle: domain.VehicleType(body.RequiredVehicle), City: body.City,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, dispatchDTO(d))
}

func (h *Handler) getDispatch(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	d, err := h.Deps.GetDispatch(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) adminList(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.AdminList(r.Context(), app.AdminListInput{
		TenantID: h.tenantID(r),
		Status:   domain.DispatchStatus(r.URL.Query().Get("status")),
		Limit:    limit, Offset: offset,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, d := range items {
		out = append(out, dispatchDTO(d))
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) parseID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) autoAssign(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	d, err := h.Deps.AutoAssign(r.Context(), app.AutoAssignInput{TenantID: h.tenantID(r), DispatchID: id})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) manualAssign(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	var body struct {
		CourierPrincipalID string `json:"courierPrincipalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cid, err := uuid.Parse(body.CourierPrincipalID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	d, err := h.Deps.ManualAssign(r.Context(), app.ManualAssignInput{
		TenantID: h.tenantID(r), DispatchID: id, CourierPrincipalID: cid,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) reassign(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	var body struct {
		CourierPrincipalID string `json:"courierPrincipalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cid, err := uuid.Parse(body.CourierPrincipalID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	d, err := h.Deps.Reassign(r.Context(), app.ReassignInput{
		TenantID: h.tenantID(r), DispatchID: id, CourierPrincipalID: cid,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) startPickup(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	d, err := h.Deps.StartPickup(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) completePickup(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	d, err := h.Deps.CompletePickup(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) startTransit(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	d, err := h.Deps.StartTransit(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) arrive(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	d, err := h.Deps.Arrive(r.Context(), h.tenantID(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) completeDelivery(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	var body struct {
		PODType      string `json:"podType"`
		PODReference string `json:"podReference"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	d, err := h.Deps.CompleteDelivery(r.Context(), app.CompleteDeliveryInput{
		TenantID: h.tenantID(r), DispatchID: id,
		PODType: domain.PODType(body.PODType), PODReference: body.PODReference,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) failDelivery(w http.ResponseWriter, r *http.Request) {
	id, ok := h.parseID(w, r)
	if !ok {
		return
	}
	var body struct {
		Reason  string `json:"reason"`
		Note    string `json:"note"`
		Requeue bool   `json:"requeue"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	d, err := h.Deps.FailDelivery(r.Context(), app.FailDeliveryInput{
		TenantID: h.tenantID(r), DispatchID: id,
		Reason: domain.FailReason(body.Reason), Note: body.Note, Requeue: body.Requeue,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dispatchDTO(d))
}

func (h *Handler) batchCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Label string `json:"label"`
		Items []struct {
			OrderID         string    `json:"orderId"`
			FulfillmentID   string    `json:"fulfillmentId"`
			WarehouseID     string    `json:"warehouseId"`
			Pickup          pointBody `json:"pickup"`
			Dropoff         pointBody `json:"dropoff"`
			RequiredVehicle string    `json:"requiredVehicle"`
			City            string    `json:"city"`
		} `json:"items"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items := make([]app.CreateDispatchInput, 0, len(body.Items))
	for _, it := range body.Items {
		oid, err := uuid.Parse(it.OrderID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		var fid, wid uuid.UUID
		if it.FulfillmentID != "" {
			fid, _ = uuid.Parse(it.FulfillmentID)
		}
		if it.WarehouseID != "" {
			wid, _ = uuid.Parse(it.WarehouseID)
		}
		items = append(items, app.CreateDispatchInput{
			OrderID: oid, FulfillmentID: fid, WarehouseID: wid,
			Pickup: domain.Point{Lat: it.Pickup.Lat, Lng: it.Pickup.Lng},
			Dropoff: domain.Point{Lat: it.Dropoff.Lat, Lng: it.Dropoff.Lng},
			RequiredVehicle: domain.VehicleType(it.RequiredVehicle), City: it.City,
		})
	}
	batch, created, err := h.Deps.BatchCreate(r.Context(), app.BatchCreateInput{
		TenantID: h.tenantID(r), Label: body.Label, Items: items,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	jobs := make([]map[string]any, 0, len(created))
	for _, d := range created {
		jobs = append(jobs, dispatchDTO(d))
	}
	writeCreated(w, map[string]any{
		"id": batch.ID, "label": batch.Label, "dispatchIds": batch.DispatchIDs,
		"jobs": jobs, "createdAt": batch.CreatedAt,
	})
}

func (h *Handler) upsertVehicle(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID       *string `json:"id"`
		Plate    string  `json:"plate"`
		Type     string  `json:"type"`
		Capacity int     `json:"capacity"`
		Active   *bool   `json:"active"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var idPtr *uuid.UUID
	if body.ID != nil && *body.ID != "" {
		id, err := uuid.Parse(*body.ID)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		idPtr = &id
	}
	v, err := h.Deps.UpsertVehicle(r.Context(), app.UpsertVehicleInput{
		TenantID: h.tenantID(r), ID: idPtr, Plate: body.Plate,
		Type: domain.VehicleType(body.Type), Capacity: body.Capacity, Active: body.Active,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if body.ID == nil {
		writeCreated(w, map[string]any{
			"id": v.ID, "tenantId": v.TenantID, "plate": v.Plate,
			"type": v.Type, "capacity": v.Capacity, "active": v.Active,
			"createdAt": v.CreatedAt, "updatedAt": v.UpdatedAt,
		})
		return
	}
	writeOK(w, map[string]any{
		"id": v.ID, "tenantId": v.TenantID, "plate": v.Plate,
		"type": v.Type, "capacity": v.Capacity, "active": v.Active,
		"createdAt": v.CreatedAt, "updatedAt": v.UpdatedAt,
	})
}

func (h *Handler) listVehicles(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, total, err := h.Deps.ListVehicles(r.Context(), h.tenantID(r), limit, offset)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, v := range items {
		out = append(out, map[string]any{
			"id": v.ID, "plate": v.Plate, "type": v.Type,
			"capacity": v.Capacity, "active": v.Active,
		})
	}
	writeOK(w, map[string]any{"items": out, "total": total})
}

func (h *Handler) upsertCourier(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CourierPrincipalID string  `json:"courierPrincipalId"`
		Available          bool    `json:"available"`
		Lat                float64 `json:"lat"`
		Lng                float64 `json:"lng"`
		CurrentLoad        int     `json:"currentLoad"`
		MaxCapacity        int     `json:"maxCapacity"`
		Rating             float64 `json:"rating"`
		VehicleType        string  `json:"vehicleType"`
		OnShift            bool    `json:"onShift"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cid, err := uuid.Parse(body.CourierPrincipalID)
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.UpsertCourierSnapshot(r.Context(), app.UpsertCourierInput{
		TenantID: h.tenantID(r), CourierPrincipalID: cid,
		Available: body.Available, Lat: body.Lat, Lng: body.Lng,
		CurrentLoad: body.CurrentLoad, MaxCapacity: body.MaxCapacity,
		Rating: body.Rating, VehicleType: domain.VehicleType(body.VehicleType),
		OnShift: body.OnShift,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"id": c.ID, "courierPrincipalId": c.CourierPrincipalID,
		"available": c.Available, "lat": c.Lat, "lng": c.Lng,
		"currentLoad": c.CurrentLoad, "maxCapacity": c.MaxCapacity,
		"rating": c.Rating, "vehicleType": c.VehicleType, "onShift": c.OnShift,
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
