package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/security-service/internal/app"
	"github.com/nexora/security-service/internal/domain"
	"github.com/nexora/security-service/internal/ratelimit"
)

type Handler struct{ Deps *app.Deps }

type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
}

func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps}
	mux := http.NewServeMux()
	const base = "/v1/security"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/policies", tenant(h.upsertPolicy))
	mux.HandleFunc("GET "+base+"/policies", tenant(h.listPolicies))
	mux.HandleFunc("POST "+base+"/policies/evaluate", tenant(h.evaluatePolicy))

	mux.HandleFunc("POST "+base+"/audits", tenant(h.recordAudit))
	mux.HandleFunc("GET "+base+"/audits", tenant(h.searchAudits))

	mux.HandleFunc("POST "+base+"/secrets", tenant(h.registerSecret))
	mux.HandleFunc("GET "+base+"/secrets", tenant(h.listSecrets))
	mux.HandleFunc("POST "+base+"/secrets/{id}/rotate", tenant(h.rotateSecret))
	mux.HandleFunc("POST "+base+"/secrets/{id}/renew-cert", tenant(h.renewCert))

	mux.HandleFunc("POST "+base+"/threats/ingest", tenant(h.ingestThreat))
	mux.HandleFunc("GET "+base+"/threats", tenant(h.listThreats))

	mux.HandleFunc("POST "+base+"/vulns", tenant(h.ingestFinding))
	mux.HandleFunc("GET "+base+"/vulns", tenant(h.listVulns))

	mux.HandleFunc("POST "+base+"/incidents", tenant(h.openIncident))
	mux.HandleFunc("POST "+base+"/incidents/{id}/close", tenant(h.closeIncident))
	mux.HandleFunc("GET "+base+"/incidents", tenant(h.listIncidents))

	mux.HandleFunc("POST "+base+"/compliance/controls", tenant(h.upsertControl))
	mux.HandleFunc("POST "+base+"/compliance/evidence", tenant(h.addEvidence))
	mux.HandleFunc("POST "+base+"/compliance/audits/run", tenant(h.runCompliance))
	mux.HandleFunc("GET "+base+"/compliance/controls", tenant(h.listControls))

	mux.HandleFunc("POST "+base+"/data-assets", tenant(h.upsertAsset))
	mux.HandleFunc("GET "+base+"/data-assets", tenant(h.listAssets))
	mux.HandleFunc("POST "+base+"/privacy/requests", tenant(h.createPrivacy))
	mux.HandleFunc("POST "+base+"/privacy/requests/{id}/complete", tenant(h.completePrivacy))

	mux.HandleFunc("POST "+base+"/risks", tenant(h.upsertRisk))
	mux.HandleFunc("GET "+base+"/risks", tenant(h.listRisks))

	mux.HandleFunc("POST "+base+"/access-requests", tenant(h.requestAccess))
	mux.HandleFunc("POST "+base+"/access-requests/{id}/decide", tenant(h.decideAccess))
	mux.HandleFunc("GET "+base+"/access-requests", tenant(h.listAccess))

	mux.HandleFunc("POST "+base+"/devices", tenant(h.upsertDevice))
	mux.HandleFunc("POST "+base+"/zero-trust/evaluate", tenant(h.zeroTrust))

	mux.HandleFunc("POST "+base+"/ai/prompt-check", tenant(h.aiPrompt))
	mux.HandleFunc("POST "+base+"/fraud/signals", tenant(h.fraudSignal))

	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.stats))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.outbox))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr: cfg.Addr, Handler: NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func requireTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return tid, true
}

func (h *Handler) upsertPolicy(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.SecurityPolicy
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.UpsertPolicy(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) listPolicies(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Policies.List(r.Context(), tid, r.URL.Query().Get("kind"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) evaluatePolicy(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		PolicyKey string         `json:"policyKey"`
		Subject   string         `json:"subject"`
		Action    string         `json:"action"`
		Resource  string         `json:"resource"`
		Input     map[string]any `json:"input"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	dec, err := h.Deps.EvaluatePolicy(r.Context(), tid, body.PolicyKey, body.Subject, body.Action, body.Resource, body.Input)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, dec)
}

func (h *Handler) recordAudit(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.AuditEvent
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	e, err := h.Deps.RecordAudit(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, e)
}

func (h *Handler) searchAudits(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.Deps.Audits.Search(r.Context(), tid, r.URL.Query().Get("action"), r.URL.Query().Get("actor"), limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) registerSecret(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.SecretMeta
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.RegisterSecret(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) listSecrets(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Secrets.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) rotateSecret(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.RotateSecret(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) renewCert(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	s, err := h.Deps.RenewCertificate(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) ingestThreat(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Subject  string             `json:"subject"`
		Features map[string]float64 `json:"features"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.IngestThreat(r.Context(), tid, body.Subject, body.Features)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, t)
}

func (h *Handler) listThreats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Threats.List(r.Context(), tid, r.URL.Query().Get("status"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) ingestFinding(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ScanFinding
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	f, err := h.Deps.IngestFinding(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, f)
}

func (h *Handler) listVulns(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Vulns.List(r.Context(), tid, r.URL.Query().Get("status"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) openIncident(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Incident
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	i, err := h.Deps.OpenIncident(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, i)
}

func (h *Handler) closeIncident(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Postmortem string `json:"postmortem"`
	}
	_ = decodeJSON(r, &body)
	i, err := h.Deps.CloseIncident(r.Context(), tid, id, body.Postmortem)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, i)
}

func (h *Handler) listIncidents(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Incidents.List(r.Context(), tid, r.URL.Query().Get("status"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertControl(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ComplianceControl
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.UpsertControl(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) addEvidence(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ComplianceEvidence
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	e, err := h.Deps.AddEvidence(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, e)
}

func (h *Handler) runCompliance(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Framework string `json:"framework"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	run, err := h.Deps.RunComplianceAudit(r.Context(), tid, body.Framework)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, run)
}

func (h *Handler) listControls(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Compliance.ListControls(r.Context(), tid, r.URL.Query().Get("framework"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertAsset(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.DataAsset
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.UpsertDataAsset(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) listAssets(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.DataGov.ListAssets(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) createPrivacy(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PrivacyRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.CreatePrivacyRequest(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) completePrivacy(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.CompletePrivacyRequest(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) upsertRisk(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.RiskItem
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	item, err := h.Deps.UpsertRisk(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, item)
}

func (h *Handler) listRisks(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Risks.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) requestAccess(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.AccessRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.RequestAccess(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) decideAccess(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Approve bool `json:"approve"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	a, err := h.Deps.DecideAccess(r.Context(), tid, id, body.Approve)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) listAccess(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Access.ListPending(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertDevice(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.DeviceTrust
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	d, err := h.Deps.UpsertDeviceTrust(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, d)
}

func (h *Handler) zeroTrust(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		Subject     string  `json:"subject"`
		DeviceID    string  `json:"deviceId"`
		ContextRisk float64 `json:"contextRisk"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ZeroTrustEvaluate(r.Context(), tid, body.Subject, body.DeviceID, body.ContextRisk)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, res)
}

func (h *Handler) aiPrompt(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		ModelKey string `json:"modelKey"`
		Prompt   string `json:"prompt"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	e, err := h.Deps.CheckAIPrompt(r.Context(), tid, body.ModelKey, body.Prompt)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, e)
}

func (h *Handler) fraudSignal(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.FraudSignal
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.RecordFraudSignal(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	s, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) outbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
