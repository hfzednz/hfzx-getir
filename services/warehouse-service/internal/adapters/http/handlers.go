package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
	"github.com/nexora/warehouse-service/internal/ratelimit"
)

// Handler serves warehouse REST endpoints.
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
	const base = "/v1/warehouse"

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET "+base+"/health", h.health)
	mux.HandleFunc("GET "+base+"/ready", h.ready)

	// Fulfillment
	mux.HandleFunc("POST "+base+"/fulfillments/receive", tenant(h.receiveFulfillment))
	mux.HandleFunc("GET "+base+"/fulfillments", tenant(h.listFulfillments))
	mux.HandleFunc("GET "+base+"/fulfillments/{id}", tenant(h.getFulfillment))
	mux.HandleFunc("POST "+base+"/fulfillments/{id}/cancel", tenant(h.cancelFulfillment))

	// Tasks
	mux.HandleFunc("GET "+base+"/tasks", tenant(h.listTasks))
	mux.HandleFunc("GET "+base+"/tasks/{id}", tenant(h.getTask))
	mux.HandleFunc("POST "+base+"/tasks/claim", tenant(h.claimPickTask))
	mux.HandleFunc("POST "+base+"/tasks/{id}/reassign", tenant(h.reassignTask))
	mux.HandleFunc("POST "+base+"/tasks/{id}/cancel", tenant(h.cancelTask))
	mux.HandleFunc("POST "+base+"/tasks/{id}/escalate", tenant(h.escalateTask))

	// Picking
	mux.HandleFunc("POST "+base+"/picking/{taskId}/start", tenant(h.startPick))
	mux.HandleFunc("POST "+base+"/picking/sessions/{sessionId}/scan", tenant(h.scanPickLine))
	mux.HandleFunc("POST "+base+"/picking/{taskId}/complete", tenant(h.completePick))

	// Packing
	mux.HandleFunc("POST "+base+"/packing/{taskId}/claim", tenant(h.claimPack))
	mux.HandleFunc("POST "+base+"/packing/sessions/{sessionId}/verify-weight", tenant(h.verifyWeight))
	mux.HandleFunc("POST "+base+"/packing/sessions/{sessionId}/seal", tenant(h.sealPack))
	mux.HandleFunc("POST "+base+"/packing/sessions/{sessionId}/label", tenant(h.generateLabel))

	// Dispatch
	mux.HandleFunc("GET "+base+"/dispatch/queue", tenant(h.listDispatchQueue))
	mux.HandleFunc("POST "+base+"/dispatch/{id}/verify", tenant(h.dispatchVerify))
	mux.HandleFunc("POST "+base+"/dispatch/{id}/handoff", tenant(h.handoffConfirm))

	// QC
	mux.HandleFunc("POST "+base+"/qc", tenant(h.createQC))
	mux.HandleFunc("POST "+base+"/qc/{id}/pass", tenant(h.qcPass))
	mux.HandleFunc("POST "+base+"/qc/{id}/fail", tenant(h.qcFail))

	// Workforce
	mux.HandleFunc("POST "+base+"/workforce/employees", tenant(h.registerEmployee))
	mux.HandleFunc("POST "+base+"/workforce/clock-in", tenant(h.clockIn))
	mux.HandleFunc("POST "+base+"/workforce/clock-out", tenant(h.clockOut))
	mux.HandleFunc("POST "+base+"/workforce/breaks/start", tenant(h.startBreak))
	mux.HandleFunc("POST "+base+"/workforce/breaks/end", tenant(h.endBreak))
	mux.HandleFunc("GET "+base+"/workforce/{employeeId}/performance", tenant(h.workforcePerf))

	// Equipment / stations
	mux.HandleFunc("POST "+base+"/stations", tenant(h.createStation))
	mux.HandleFunc("GET "+base+"/stations", tenant(h.listStations))
	mux.HandleFunc("POST "+base+"/equipment", tenant(h.registerEquipment))
	mux.HandleFunc("GET "+base+"/equipment", tenant(h.listEquipment))
	mux.HandleFunc("POST "+base+"/equipment/{id}/heartbeat", tenant(h.equipmentHeartbeat))

	// Sensors stub
	mux.HandleFunc("POST "+base+"/sensors/ingest", tenant(h.sensorIngest))

	// AI
	mux.HandleFunc("POST "+base+"/ai/optimize-route", tenant(h.optimizeRoute))

	// Admin
	mux.HandleFunc("GET "+base+"/admin/dashboard", tenant(h.dashboard))

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

func parseActor(r *http.Request) *uuid.UUID {
	if v := r.Header.Get("X-Actor-Id"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			return &id
		}
	}
	return nil
}

func queryUUID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.URL.Query().Get(name))
}

// --- Fulfillment ---

func (h *Handler) receiveFulfillment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID     uuid.UUID `json:"warehouseId"`
		ExternalOrderID string    `json:"externalOrderId"`
		Strategy        string    `json:"strategy"`
		Priority        int       `json:"priority"`
		IdempotencyKey  string    `json:"idempotencyKey"`
		Lines           []struct {
			VariantID    uuid.UUID `json:"variantId"`
			SKUCode      string    `json:"skuCode"`
			Barcode      string    `json:"barcode"`
			LocationCode string    `json:"locationCode"`
			Qty          int64     `json:"qty"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]app.ReceiveLineInput, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, app.ReceiveLineInput{
			VariantID: l.VariantID, SKUCode: l.SKUCode, Barcode: l.Barcode,
			LocationCode: l.LocationCode, Qty: l.Qty,
		})
	}
	idem := r.Header.Get("Idempotency-Key")
	if idem == "" {
		idem = body.IdempotencyKey
	}
	out, err := h.Deps.ReceiveFulfillment(r.Context(), app.ReceiveFulfillmentCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, ExternalOrderID: body.ExternalOrderID,
		Strategy: domain.PickStrategy(body.Strategy), Priority: body.Priority,
		Lines: lines, IdempotencyKey: idem, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) getFulfillment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetFulfillment(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) listFulfillments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	f := ports.FulfillmentFilter{TenantID: h.tenant(r), Limit: limit, Offset: offset}
	if wh := r.URL.Query().Get("warehouseId"); wh != "" {
		id, err := uuid.Parse(wh)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		f.WarehouseID = &id
	}
	if st := r.URL.Query().Get("status"); st != "" {
		s := domain.FulfillmentStatus(st)
		f.Status = &s
	}
	list, total, err := h.Deps.ListFulfillments(r.Context(), f)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list, "total": total})
}

func (h *Handler) cancelFulfillment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.CancelFulfillment(r.Context(), app.CancelFulfillmentCmd{
		TenantID: h.tenant(r), FulfillmentID: id, Reason: body.Reason, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Tasks ---

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	wh, err := queryUUID(r, "warehouseId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	f := ports.TaskFilter{TenantID: h.tenant(r), WarehouseID: wh, Limit: limit}
	if t := r.URL.Query().Get("type"); t != "" {
		tt := domain.TaskType(t)
		f.Type = &tt
	}
	if s := r.URL.Query().Get("status"); s != "" {
		st := domain.TaskStatus(s)
		f.Status = &st
	}
	list, total, err := h.Deps.ListTasks(r.Context(), f)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list, "total": total})
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.GetTask(r.Context(), h.tenant(r), id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) claimPickTask(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID  `json:"warehouseId"`
		TaskID      *uuid.UUID `json:"taskId"`
		PickerID    uuid.UUID  `json:"pickerId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ClaimPickTask(r.Context(), app.ClaimPickTaskCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, TaskID: body.TaskID, PickerID: body.PickerID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) reassignTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		NewAssignee uuid.UUID `json:"newAssigneeId"`
		Note        string    `json:"note"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ReassignTask(r.Context(), app.ReassignTaskCmd{
		TenantID: h.tenant(r), TaskID: id, NewAssignee: body.NewAssignee, Note: body.Note, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) cancelTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.CancelTask(r.Context(), app.CancelTaskCmd{
		TenantID: h.tenant(r), TaskID: id, Reason: body.Reason, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) escalateTask(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.EscalateTask(r.Context(), app.EscalateTaskCmd{
		TenantID: h.tenant(r), TaskID: id, Note: body.Note, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Picking ---

func (h *Handler) startPick(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUIDParam(r, "taskId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		PickerID uuid.UUID `json:"pickerId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.StartPick(r.Context(), app.StartPickCmd{
		TenantID: h.tenant(r), TaskID: taskID, PickerID: body.PickerID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) scanPickLine(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUIDParam(r, "sessionId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		LineID   uuid.UUID `json:"lineId"`
		Barcode  string    `json:"barcode"`
		Qty      int64     `json:"qty"`
		PickerID uuid.UUID `json:"pickerId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ScanPickLine(r.Context(), app.ScanPickLineCmd{
		TenantID: h.tenant(r), SessionID: sessionID, LineID: body.LineID,
		Barcode: body.Barcode, Qty: body.Qty, PickerID: body.PickerID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) completePick(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUIDParam(r, "taskId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		PickerID uuid.UUID `json:"pickerId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.CompletePick(r.Context(), app.CompletePickCmd{
		TenantID: h.tenant(r), TaskID: taskID, PickerID: body.PickerID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Packing ---

func (h *Handler) claimPack(w http.ResponseWriter, r *http.Request) {
	taskID, err := parseUUIDParam(r, "taskId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		PackerID  uuid.UUID `json:"packerId"`
		StationID uuid.UUID `json:"stationId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ClaimPack(r.Context(), app.ClaimPackCmd{
		TenantID: h.tenant(r), TaskID: taskID, PackerID: body.PackerID, StationID: body.StationID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) verifyWeight(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUIDParam(r, "sessionId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		ActualWeightG int64     `json:"actualWeightG"`
		PackerID      uuid.UUID `json:"packerId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.VerifyWeight(r.Context(), app.VerifyWeightCmd{
		TenantID: h.tenant(r), PackSessionID: sessionID, ActualWeightG: body.ActualWeightG, PackerID: body.PackerID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) sealPack(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUIDParam(r, "sessionId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		PackerID uuid.UUID `json:"packerId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.Seal(r.Context(), app.SealPackCmd{
		TenantID: h.tenant(r), PackSessionID: sessionID, PackerID: body.PackerID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) generateLabel(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parseUUIDParam(r, "sessionId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		PackerID uuid.UUID `json:"packerId"`
		Format   string    `json:"format"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	label, unit, err := h.Deps.GenerateLabel(r.Context(), app.GenerateLabelCmd{
		TenantID: h.tenant(r), PackSessionID: sessionID, PackerID: body.PackerID, Format: body.Format,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"label": label, "dispatchUnit": unit})
}

// --- Dispatch ---

func (h *Handler) listDispatchQueue(w http.ResponseWriter, r *http.Request) {
	wh, err := queryUUID(r, "warehouseId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	list, total, err := h.Deps.ListDispatchQueue(r.Context(), h.tenant(r), wh, limit, 0)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": list, "total": total})
}

func (h *Handler) dispatchVerify(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		TrackingCode string `json:"trackingCode"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.DispatchVerify(r.Context(), app.DispatchVerifyCmd{
		TenantID: h.tenant(r), DispatchUnitID: id, TrackingCode: body.TrackingCode, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) handoffConfirm(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		CourierRef string `json:"courierRef"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.HandoffConfirm(r.Context(), app.HandoffConfirmCmd{
		TenantID: h.tenant(r), DispatchUnitID: id, CourierRef: body.CourierRef, ActorID: parseActor(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- QC ---

func (h *Handler) createQC(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID   uuid.UUID  `json:"warehouseId"`
		FulfillmentID uuid.UUID  `json:"fulfillmentId"`
		InspectorID   *uuid.UUID `json:"inspectorId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.CreateQCInspection(r.Context(), app.CreateQCCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, FulfillmentID: body.FulfillmentID, InspectorID: body.InspectorID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) qcPass(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Notes       string     `json:"notes"`
		InspectorID *uuid.UUID `json:"inspectorId"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.QCPass(r.Context(), app.QCPassCmd{
		TenantID: h.tenant(r), InspectionID: id, Notes: body.Notes, InspectorID: body.InspectorID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) qcFail(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Notes       string     `json:"notes"`
		DefectCodes []string   `json:"defectCodes"`
		InspectorID *uuid.UUID `json:"inspectorId"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.QCFail(r.Context(), app.QCFailCmd{
		TenantID: h.tenant(r), InspectionID: id, Notes: body.Notes, DefectCodes: body.DefectCodes, InspectorID: body.InspectorID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Workforce ---

func (h *Handler) registerEmployee(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID `json:"warehouseId"`
		ExternalRef string    `json:"externalRef"`
		DisplayName string    `json:"displayName"`
		Role        string    `json:"role"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.RegisterEmployee(r.Context(), app.RegisterEmployeeCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, ExternalRef: body.ExternalRef,
		DisplayName: body.DisplayName, Role: domain.EmployeeRole(body.Role),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) clockIn(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID `json:"warehouseId"`
		EmployeeID  uuid.UUID `json:"employeeId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ClockIn(r.Context(), app.ClockInCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, EmployeeID: body.EmployeeID,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) clockOut(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID uuid.UUID `json:"employeeId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ClockOut(r.Context(), app.ClockOutCmd{TenantID: h.tenant(r), EmployeeID: body.EmployeeID})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) startBreak(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID uuid.UUID `json:"employeeId"`
		Reason     string    `json:"reason"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.StartBreak(r.Context(), app.StartBreakCmd{
		TenantID: h.tenant(r), EmployeeID: body.EmployeeID, Reason: body.Reason,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) endBreak(w http.ResponseWriter, r *http.Request) {
	var body struct {
		EmployeeID uuid.UUID `json:"employeeId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.EndBreak(r.Context(), app.EndBreakCmd{TenantID: h.tenant(r), EmployeeID: body.EmployeeID})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) workforcePerf(w http.ResponseWriter, r *http.Request) {
	empID, err := parseUUIDParam(r, "employeeId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	wh, err := queryUUID(r, "warehouseId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.WorkforcePerformance(r.Context(), h.tenant(r), wh, empID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

// --- Stations / Equipment ---

func (h *Handler) createStation(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID `json:"warehouseId"`
		Code        string    `json:"code"`
		Name        string    `json:"name"`
		Type        string    `json:"type"`
		Zone        string    `json:"zone"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.CreateStation(r.Context(), app.CreateStationCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, Code: body.Code, Name: body.Name,
		Type: domain.StationType(body.Type), Zone: body.Zone,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) listStations(w http.ResponseWriter, r *http.Request) {
	wh, err := queryUUID(r, "warehouseId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ListStations(r.Context(), h.tenant(r), wh)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) registerEquipment(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID      `json:"warehouseId"`
		Code        string         `json:"code"`
		Kind        string         `json:"kind"`
		Metadata    map[string]any `json:"metadata"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.RegisterEquipment(r.Context(), app.RegisterEquipmentCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, Code: body.Code, Kind: body.Kind, Metadata: body.Metadata,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, out)
}

func (h *Handler) listEquipment(w http.ResponseWriter, r *http.Request) {
	wh, err := queryUUID(r, "warehouseId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.ListEquipment(r.Context(), h.tenant(r), wh)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) equipmentHeartbeat(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Status string `json:"status"`
	}
	_ = decodeJSON(r, &body)
	out, err := h.Deps.EquipmentHeartbeat(r.Context(), app.EquipmentHeartbeatCmd{
		TenantID: h.tenant(r), EquipmentID: id, Status: domain.EquipmentStatus(body.Status),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) sensorIngest(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	writeOK(w, map[string]any{"accepted": true, "stub": true, "payload": body})
}

// --- AI / Admin ---

func (h *Handler) optimizeRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		WarehouseID uuid.UUID `json:"warehouseId"`
		Lines       []struct {
			LineID       uuid.UUID `json:"lineId"`
			LocationCode string    `json:"locationCode"`
			SKUCode      string    `json:"skuCode"`
			Sequence     int       `json:"sequence"`
		} `json:"lines"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	lines := make([]ports.RouteLine, 0, len(body.Lines))
	for _, l := range body.Lines {
		lines = append(lines, ports.RouteLine{
			LineID: l.LineID, LocationCode: l.LocationCode, SKUCode: l.SKUCode, Sequence: l.Sequence,
		})
	}
	out, err := h.Deps.OptimizeRoute(r.Context(), app.OptimizeRouteCmd{
		TenantID: h.tenant(r), WarehouseID: body.WarehouseID, Lines: lines,
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"lines": out})
}

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	wh, err := queryUUID(r, "warehouseId")
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.Dashboard(r.Context(), h.tenant(r), wh)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}
