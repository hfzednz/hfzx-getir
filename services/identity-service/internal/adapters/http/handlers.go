package httpadapter

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

// Handler holds use-case dependencies for REST routes.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
}

func (h *Handler) authCtx(r *http.Request) ports.AuthContext {
	return ports.AuthContext{
		UserAgent: r.UserAgent(),
		RequestID: RequestIDFromContext(r.Context()),
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

func parseUUID(s string) (uuid.UUID, error) {
	if s == "" {
		return uuid.Nil, domain.ErrInvalidArgument
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, domain.ErrInvalidArgument
	}
	return id, nil
}

func tokenPairJSON(p ports.TokenPair) map[string]any {
	expIn := int64(0)
	if !p.ExpiresAt.IsZero() {
		expIn = int64(time.Until(p.ExpiresAt).Seconds())
		if expIn < 0 {
			expIn = 0
		}
	}
	return map[string]any{
		"accessToken":  p.AccessToken,
		"refreshToken": p.RefreshToken,
		"tokenType":    "Bearer",
		"expiresIn":    expIn,
		"sessionId":    p.SessionID.String(),
	}
}

func authResultJSON(res app.AuthResult) map[string]any {
	if res.MFARequired {
		out := map[string]any{
			"mfaRequired": true,
			"expiresIn":   300,
		}
		if res.MFAChallengeID != nil {
			out["challengeId"] = res.MFAChallengeID.String()
		}
		return out
	}
	out := tokenPairJSON(res.Tokens)
	out["principalId"] = res.Principal.ID.String()
	out["riskScore"] = res.RiskScore
	return out
}

// --- Auth ---

func (h *Handler) otpStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Phone    string `json:"phone"`
		TenantID string `json:"tenantId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	id, err := h.Deps.StartOTP(r.Context(), app.StartOTPInput{
		TenantID: tid, Phone: body.Phone, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"challengeId": id.String(), "expiresIn": 300})
}

func (h *Handler) otpVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChallengeID string `json:"challengeId"`
		Code        string `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cid, err := parseUUID(body.ChallengeID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.VerifyOTP(r.Context(), app.VerifyOTPInput{
		ChallengeID: cid, Code: body.Code, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

func (h *Handler) passwordLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TenantID string `json:"tenantId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.Login(r.Context(), app.LoginInput{
		TenantID: tid, Email: body.Email, Password: body.Password, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

func (h *Handler) magicLinkStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email    string `json:"email"`
		TenantID string `json:"tenantId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	token, err := h.Deps.StartMagicLink(r.Context(), app.StartMagicLinkInput{
		TenantID: tid, Email: body.Email, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"challengeId": token, "expiresIn": 900})
}

func (h *Handler) magicLinkConsume(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token    string `json:"token"`
		TenantID string `json:"tenantId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	res, err := h.Deps.ConsumeMagicLink(r.Context(), app.ConsumeMagicLinkInput{
		Token: body.Token, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

func (h *Handler) socialStart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider    string `json:"provider"`
		RedirectURI string `json:"redirectUri"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.Provider == "" {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	writeOK(w, map[string]string{
		"authorizationUrl": "https://accounts.example/" + body.Provider + "/authorize?redirect_uri=" + body.RedirectURI,
		"state":            uuid.NewString(),
	})
}

func (h *Handler) socialCallback(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Provider    string `json:"provider"`
		Code        string `json:"code"`
		RedirectURI string `json:"redirectUri"`
		TenantID    string `json:"tenantId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.HandleSocialCallback(r.Context(), app.SocialCallbackInput{
		TenantID: tid, Provider: ports.SocialProvider(body.Provider),
		Code: body.Code, RedirectURI: body.RedirectURI, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

func (h *Handler) webauthnRegisterOptions(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID string `json:"principalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	opts, err := h.Deps.BeginWebAuthnRegistration(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, opts)
}

func (h *Handler) webauthnRegisterVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID string                                  `json:"principalId"`
		Attestation *webauthn.AuthenticatorAttestationResponse `json:"attestation"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.FinishWebAuthnRegistration(r.Context(), pid, body.Attestation); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]string{"status": "registered", "principalId": pid.String()})
}

func (h *Handler) webauthnAuthOptions(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Identifier string `json:"identifier"`
		TenantID   string `json:"tenantId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	opts, err := h.Deps.BeginWebAuthnLogin(r.Context(), tid, body.Identifier)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, opts)
}

func (h *Handler) webauthnAuthVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TenantID  string                                 `json:"tenantId"`
		Assertion *webauthn.AuthenticatorAssertionResponse `json:"assertion"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.FinishWebAuthnLogin(r.Context(), tid, body.Assertion, h.authCtx(r))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

func (h *Handler) guest(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TenantID string `json:"tenantId"`
	}
	_ = decodeJSON(r, &body)
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.CreateGuestSession(r.Context(), app.CreateGuestSessionInput{
		TenantID: tid, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

// --- MFA ---

func (h *Handler) mfaTotpEnroll(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID string `json:"principalId"`
		AccountName string `json:"accountName"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out, err := h.Deps.EnrollTOTP(r.Context(), pid, body.AccountName)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"factorId":    out.FactorID.String(),
		"secret":      out.Secret,
		"otpauth":     out.OTPAuthURI,
		"backupCodes": out.BackupCodes,
	})
}

func (h *Handler) mfaChallenge(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID string `json:"principalId"`
		SessionID   string `json:"sessionId"`
		FactorType  string `json:"factorType"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	ft := domain.MFAFactorType(body.FactorType)
	if ft == "" {
		ft = domain.MFAFactorTOTP
	}
	chID := uuid.New()
	sid, _ := uuid.Parse(body.SessionID)
	now := time.Now().UTC()
	if err := h.Deps.OAuth.SaveMFAChallenge(r.Context(), ports.MFAChallenge{
		ID: chID, PrincipalID: pid, SessionHint: sid, FactorType: ft,
		ExpiresAt: now.Add(5 * time.Minute), CreatedAt: now,
	}); err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"challengeId": chID.String(), "expiresIn": 300, "mfaRequired": true})
}

func (h *Handler) mfaVerify(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ChallengeID string `json:"challengeId"`
		Code        string `json:"code"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	cid, err := parseUUID(body.ChallengeID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.VerifyMFAChallenge(r.Context(), app.VerifyMFAChallengeInput{
		ChallengeID: cid, Code: body.Code, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, authResultJSON(res))
}

// --- Tokens ---

func (h *Handler) tokenRefresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refreshToken"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pair, err := h.Deps.Refresh(r.Context(), body.RefreshToken)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, tokenPairJSON(pair))
}

func (h *Handler) tokenRevoke(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token     string `json:"token"`
		SessionID string `json:"sessionId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.SessionID != "" {
		sid, err := parseUUID(body.SessionID)
		if err != nil {
			writeErr(w, r, err)
			return
		}
		if err := h.Deps.Logout(r.Context(), app.LogoutInput{
			SessionID: sid, Reason: "revoke", AuthCtx: h.authCtx(r),
		}); err != nil {
			writeErr(w, r, err)
			return
		}
	}
	writeNoContent(w)
}

func (h *Handler) tokenIntrospect(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	if body.Token == "" || h.Deps.JWTKeys == nil {
		writeOK(w, map[string]any{"active": false})
		return
	}
	claims, err := h.Deps.JWTKeys.ParseAndValidate(body.Token, h.Deps.Issuer, h.Deps.Audience)
	if err != nil {
		writeOK(w, map[string]any{"active": false})
		return
	}
	writeOK(w, map[string]any{
		"active":    true,
		"tokenType": "access_token",
		"sub":       claims.Subject,
		"sid":       claims.Session,
		"tid":       claims.Tenant,
		"roles":     claims.Roles,
		"exp":       claims.Expires.Unix(),
		"iat":       claims.IssuedAt.Unix(),
	})
}

// --- Sessions / Devices ---

func (h *Handler) listSessions(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(r.URL.Query().Get("principalId"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items, err := h.Deps.ListSessions(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, s := range items {
		out = append(out, map[string]any{
			"id": s.ID.String(), "principalId": s.PrincipalID.String(),
			"tenantId": s.TenantID.String(), "acr": s.ACR,
			"createdAt": s.CreatedAt, "lastSeenAt": s.LastSeenAt, "revoked": s.IsRevoked(),
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) revokeSession(w http.ResponseWriter, r *http.Request) {
	sid, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.RevokeSession(r.Context(), app.RevokeSessionInput{
		SessionID: sid, AllowAdmin: true, Reason: "user_revoke", AuthCtx: h.authCtx(r),
	}); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) listDevices(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(r.URL.Query().Get("principalId"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items, err := h.Deps.ListDevices(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, d := range items {
		out = append(out, map[string]any{
			"id": d.ID.String(), "principalId": d.PrincipalID.String(),
			"name": d.Name, "platform": d.Platform, "trusted": d.IsTrusted(),
			"revoked": d.IsRevoked(), "lastSeenAt": d.LastSeenAt,
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) trustDevice(w http.ResponseWriter, r *http.Request) {
	did, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body struct {
		PrincipalID string `json:"principalId"`
	}
	_ = decodeJSON(r, &body)
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if _, err := h.Deps.TrustDevice(r.Context(), pid, did); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) revokeDevice(w http.ResponseWriter, r *http.Request) {
	did, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body struct {
		PrincipalID string `json:"principalId"`
	}
	_ = decodeJSON(r, &body)
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.RevokeDevice(r.Context(), pid, did); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

// --- Principals ---

func (h *Handler) listPrincipals(w http.ResponseWriter, r *http.Request) {
	tid, err := parseUUID(r.URL.Query().Get("tenantId"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	q := r.URL.Query().Get("q")
	items, err := h.Deps.SearchPrincipals(r.Context(), tid, q, limit)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, map[string]any{
			"id": p.ID.String(), "tenantId": p.TenantID.String(),
			"kind": p.Kind, "status": p.Status, "displayName": p.DisplayName,
			"createdAt": p.CreatedAt,
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) getPrincipal(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	p, err := h.Deps.Principals.GetByID(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{
		"id": p.ID.String(), "tenantId": p.TenantID.String(),
		"kind": p.Kind, "status": p.Status, "displayName": p.DisplayName,
		"createdAt": p.CreatedAt, "updatedAt": p.UpdatedAt,
	})
}

func (h *Handler) createPrincipal(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TenantID    string `json:"tenantId"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	tid, err := parseUUID(body.TenantID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	res, err := h.Deps.Register(r.Context(), app.RegisterInput{
		TenantID: tid, Email: body.Email, Password: body.Password,
		DisplayName: body.DisplayName, AuthCtx: h.authCtx(r),
	})
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{
		"id": res.Principal.ID.String(), "tenantId": res.Principal.TenantID.String(),
		"kind": res.Principal.Kind, "status": res.Principal.Status,
		"displayName": res.Principal.DisplayName,
		"tokens":      tokenPairJSON(res.Tokens),
	})
}

// --- RBAC ---

func (h *Handler) listRoles(w http.ResponseWriter, r *http.Request) {
	if h.Deps == nil || h.Deps.Roles == nil {
		writeOK(w, map[string]any{"items": []any{}, "total": 0})
		return
	}
	names := []string{
		"customer", "courier", "picker", "packer", "dispatcher", "city_ops",
		"support_agent", "finance_analyst", "admin", "super_admin",
		"service_account", "partner", "supplier",
	}
	out := make([]map[string]any, 0, len(names))
	for _, name := range names {
		role, err := h.Deps.Roles.GetRoleByName(r.Context(), nil, name)
		if err != nil {
			continue
		}
		out = append(out, map[string]any{
			"id": role.ID.String(), "name": role.Name, "kind": string(role.Kind),
			"description": role.Description, "createdAt": role.CreatedAt, "updatedAt": role.UpdatedAt,
		})
	}
	writeOK(w, map[string]any{"items": out, "total": len(out)})
}

func (h *Handler) listPermissions(w http.ResponseWriter, r *http.Request) {
	pidStr := r.URL.Query().Get("principalId")
	if pidStr == "" {
		writeOK(w, map[string]any{"items": []any{}})
		return
	}
	pid, err := parseUUID(pidStr)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items, err := h.Deps.ListEffectivePermissions(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, p := range items {
		out = append(out, map[string]any{
			"id": p.ID.String(), "resource": p.Resource, "action": p.Action,
			"permission": p.String(), "description": p.Description,
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) listPrincipalRoles(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	items, err := h.Deps.Roles.ListPrincipalRoles(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, b := range items {
		out = append(out, map[string]any{
			"id": b.ID.String(), "principalId": b.PrincipalID.String(),
			"roleId": b.RoleID.String(), "createdAt": b.CreatedAt,
		})
	}
	writeOK(w, map[string]any{"items": out})
}

func (h *Handler) assignPrincipalRole(w http.ResponseWriter, r *http.Request) {
	pid, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	var body struct {
		RoleID string `json:"roleId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	rid, err := parseUUID(body.RoleID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if _, err := h.Deps.AssignRole(r.Context(), app.AssignRoleInput{
		PrincipalID: pid, RoleID: rid, AuthCtx: h.authCtx(r),
	}); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}

// --- Privacy ---

func (h *Handler) privacyExport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID string `json:"principalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	exp, err := h.Deps.ExportPrincipalData(r.Context(), pid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, exp)
}

func (h *Handler) privacyDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PrincipalID string `json:"principalId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	pid, err := parseUUID(body.PrincipalID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	if err := h.Deps.RequestDeletion(r.Context(), pid, h.authCtx(r)); err != nil {
		writeErr(w, r, err)
		return
	}
	writeNoContent(w)
}
