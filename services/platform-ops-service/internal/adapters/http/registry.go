package httpadapter

import (
	"encoding/json"
	"net/http"

	"github.com/nexora/platform-ops-service/internal/domain"
)

func decodeBody(r *http.Request, dst any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(dst)
}

func actorFrom(r *http.Request) string {
	if v := r.Header.Get("X-Actor-Email"); v != "" {
		return v
	}
	if v := r.Header.Get("X-Principal-Id"); v != "" {
		return v
	}
	return "super_admin"
}

func (h *Handler) listTenants(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	out, err := h.Deps.ListTenants(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getTenant(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	out, err := h.Deps.GetTenant(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) createTenant(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	var body struct {
		Name          string `json:"name"`
		Slug          string `json:"slug"`
		CompanyID     string `json:"companyId"`
		IsolationMode string `json:"isolationMode"`
		Region        string `json:"region"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.CreateTenant(r.Context(), body.Name, body.Slug, body.CompanyID, body.IsolationMode, body.Region, actorFrom(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, t)
}

func (h *Handler) patchTenantIsolation(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	var body struct {
		IsolationMode string `json:"isolationMode"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.PatchTenantIsolation(r.Context(), r.PathValue("id"), body.IsolationMode, actorFrom(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) proposeTenant(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	var body struct {
		TenantID       string `json:"tenantId"`
		Action         string `json:"action"`
		Reason         string `json:"reason"`
		RequesterID    string `json:"requesterId"`
		RequesterEmail string `json:"requesterEmail"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.ProposeTenantAction(r.Context(), body.TenantID, body.Action, body.Reason, body.RequesterID, body.RequesterEmail)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) resolveTenantProposal(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	var body struct {
		Decision   string `json:"decision"`
		ApproverID string `json:"approverId"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.ResolveTenantProposal(r.Context(), r.PathValue("id"), body.Decision, actorFrom(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) listCompanies(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	out, err := h.Deps.ListCompanies(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) getCompany(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	out, err := h.Deps.GetCompany(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) createCompany(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	var body struct {
		LegalName        string `json:"legalName"`
		TradeName        string `json:"tradeName"`
		CountryCode      string `json:"countryCode"`
		PrimaryCurrency  string `json:"primaryCurrency"`
	}
	if err := decodeBody(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	c, err := h.Deps.CreateCompany(r.Context(), body.LegalName, body.TradeName, body.CountryCode, body.PrimaryCurrency, actorFrom(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) patchCompany(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	var patch map[string]any
	if err := decodeBody(r, &patch); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	out, err := h.Deps.PatchCompany(r.Context(), r.PathValue("id"), patch, actorFrom(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) deleteCompany(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	if err := h.Deps.DeleteCompany(r.Context(), r.PathValue("id"), actorFrom(r)); err != nil {
		writeErr(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) rolesSnapshot(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	writeOK(w, h.Deps.RolesSnapshot(r.Context()))
}

func (h *Handler) orgSnapshot(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	out, err := h.Deps.OrgSnapshot(r.Context())
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

func (h *Handler) auditSnapshot(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireTenant(w, r); !ok {
		return
	}
	out, err := h.Deps.AuditSnapshot(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, out)
}

