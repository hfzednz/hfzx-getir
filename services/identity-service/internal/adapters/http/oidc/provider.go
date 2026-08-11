// Package oidc implements a production-shaped OIDC provider surface with an in-memory code store.
package oidc

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	secjwt "github.com/nexora/identity-service/internal/security/jwt"
)

// Config configures the OIDC provider.
type Config struct {
	Issuer   string
	Audience string
	Keys     *secjwt.KeyManager
	// AccessTTL for issued tokens from /oidc/token.
	AccessTTL time.Duration
}

// Provider is a minimal Authorization Code + PKCE OIDC provider.
type Provider struct {
	cfg   Config
	mu    sync.Mutex
	codes map[string]authCode
}

type authCode struct {
	ClientID            string
	RedirectURI         string
	Subject             string
	Scope               string
	Nonce               string
	CodeChallenge       string
	CodeChallengeMethod string
	ExpiresAt           time.Time
}

// NewProvider constructs an OIDC provider.
func NewProvider(cfg Config) *Provider {
	if cfg.AccessTTL <= 0 {
		cfg.AccessTTL = 15 * time.Minute
	}
	return &Provider{
		cfg:   cfg,
		codes: make(map[string]authCode),
	}
}

// Mount registers OIDC routes on mux.
func (p *Provider) Mount(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/openid-configuration", p.discovery)
	mux.HandleFunc("GET /v1/oidc/.well-known/openid-configuration", p.discovery)
	mux.HandleFunc("GET /v1/oidc/authorize", p.authorize)
	mux.HandleFunc("POST /v1/oidc/token", p.token)
	mux.HandleFunc("GET /v1/oidc/jwks", p.jwks)
	mux.HandleFunc("GET /v1/oidc/userinfo", p.userinfo)
	mux.HandleFunc("POST /v1/oidc/userinfo", p.userinfo)
}

func (p *Provider) discovery(w http.ResponseWriter, r *http.Request) {
	base := strings.TrimRight(p.cfg.Issuer, "/")
	doc := map[string]any{
		"issuer":                                base,
		"authorization_endpoint":                base + "/v1/oidc/authorize",
		"token_endpoint":                        base + "/v1/oidc/token",
		"userinfo_endpoint":                     base + "/v1/oidc/userinfo",
		"jwks_uri":                              base + "/v1/oidc/jwks",
		"response_types_supported":              []string{"code"},
		"subject_types_supported":               []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"scopes_supported":                      []string{"openid", "profile", "email", "offline_access"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "none"},
		"code_challenge_methods_supported":      []string{"S256", "plain"},
		"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
		"claims_supported":                      []string{"sub", "iss", "aud", "exp", "iat", "sid", "tid", "roles"},
	}
	writeJSON(w, http.StatusOK, doc)
}

func (p *Provider) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	scope := q.Get("scope")
	state := q.Get("state")
	nonce := q.Get("nonce")
	challenge := q.Get("code_challenge")
	method := q.Get("code_challenge_method")
	// Dev convenience: login_hint / sub query acts as authenticated subject.
	subject := q.Get("login_hint")
	if subject == "" {
		subject = q.Get("sub")
	}
	if subject == "" {
		subject = "anonymous-" + uuid.NewString()
	}

	if clientID == "" || redirectURI == "" || responseType != "code" {
		http.Error(w, "invalid_request", http.StatusBadRequest)
		return
	}
	if !strings.Contains(scope, "openid") {
		http.Error(w, "invalid_scope", http.StatusBadRequest)
		return
	}
	if challenge == "" {
		http.Error(w, "code_challenge required", http.StatusBadRequest)
		return
	}
	if method == "" {
		method = "plain"
	}
	if method != "S256" && method != "plain" {
		http.Error(w, "unsupported code_challenge_method", http.StatusBadRequest)
		return
	}

	code := randomToken(32)
	p.mu.Lock()
	p.codes[code] = authCode{
		ClientID:            clientID,
		RedirectURI:         redirectURI,
		Subject:             subject,
		Scope:               scope,
		Nonce:               nonce,
		CodeChallenge:       challenge,
		CodeChallengeMethod: method,
		ExpiresAt:           time.Now().UTC().Add(10 * time.Minute),
	}
	p.mu.Unlock()

	u, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "invalid redirect_uri", http.StatusBadRequest)
		return
	}
	vals := u.Query()
	vals.Set("code", code)
	if state != "" {
		vals.Set("state", state)
	}
	u.RawQuery = vals.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func (p *Provider) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "cannot parse form")
		return
	}
	grant := r.FormValue("grant_type")
	switch grant {
	case "authorization_code":
		p.tokenAuthCode(w, r)
	case "refresh_token":
		p.tokenRefresh(w, r)
	default:
		writeOAuthError(w, http.StatusBadRequest, "unsupported_grant_type", "only authorization_code and refresh_token")
	}
}

func (p *Provider) tokenAuthCode(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	verifier := r.FormValue("code_verifier")

	p.mu.Lock()
	ac, ok := p.codes[code]
	if ok {
		delete(p.codes, code)
	}
	p.mu.Unlock()
	if !ok || time.Now().UTC().After(ac.ExpiresAt) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "code invalid or expired")
		return
	}
	if ac.ClientID != clientID || ac.RedirectURI != redirectURI {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "client/redirect mismatch")
		return
	}
	if !verifyPKCE(verifier, ac.CodeChallenge, ac.CodeChallengeMethod) {
		writeOAuthError(w, http.StatusBadRequest, "invalid_grant", "pkce verification failed")
		return
	}
	if p.cfg.Keys == nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "signing keys not configured")
		return
	}

	now := time.Now().UTC()
	access, err := p.cfg.Keys.IssueAccessToken(secjwt.AccessClaims{
		Subject:  ac.Subject,
		Session:  uuid.NewString(),
		Tenant:   "00000000-0000-0000-0000-000000000000",
		Roles:    []string{},
		AMR:      []string{"pwd"},
		ACR:      "urn:nexora:acr:1",
		Issuer:   p.cfg.Issuer,
		Audience: p.cfg.Audience,
		Expires:  now.Add(p.cfg.AccessTTL),
		IssuedAt: now,
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	idToken, err := p.cfg.Keys.IssueAccessToken(secjwt.AccessClaims{
		Subject:  ac.Subject,
		Session:  uuid.NewString(),
		Tenant:   "00000000-0000-0000-0000-000000000000",
		Issuer:   p.cfg.Issuer,
		Audience: clientID,
		Expires:  now.Add(p.cfg.AccessTTL),
		IssuedAt: now,
		JTI:      uuid.NewString(),
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	_ = ac.Nonce
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    int64(p.cfg.AccessTTL.Seconds()),
		"refresh_token": "oidc.refresh." + uuid.NewString(),
		"id_token":      idToken,
		"scope":         ac.Scope,
	})
}

func (p *Provider) tokenRefresh(w http.ResponseWriter, r *http.Request) {
	rt := r.FormValue("refresh_token")
	if rt == "" {
		writeOAuthError(w, http.StatusBadRequest, "invalid_request", "refresh_token required")
		return
	}
	if p.cfg.Keys == nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "signing keys not configured")
		return
	}
	now := time.Now().UTC()
	access, err := p.cfg.Keys.IssueAccessToken(secjwt.AccessClaims{
		Subject:  "refresh-subject",
		Session:  uuid.NewString(),
		Tenant:   "00000000-0000-0000-0000-000000000000",
		Issuer:   p.cfg.Issuer,
		Audience: p.cfg.Audience,
		Expires:  now.Add(p.cfg.AccessTTL),
		IssuedAt: now,
	})
	if err != nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":  access,
		"token_type":    "Bearer",
		"expires_in":    int64(p.cfg.AccessTTL.Seconds()),
		"refresh_token": "oidc.refresh." + uuid.NewString(),
	})
}

func (p *Provider) jwks(w http.ResponseWriter, r *http.Request) {
	if p.cfg.Keys == nil {
		http.Error(w, "keys unavailable", http.StatusServiceUnavailable)
		return
	}
	doc, err := p.cfg.Keys.ExportJWKS()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, doc)
}

func (p *Provider) userinfo(w http.ResponseWriter, r *http.Request) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		w.Header().Set("WWW-Authenticate", `Bearer realm="oidc"`)
		writeOAuthError(w, http.StatusUnauthorized, "invalid_token", "missing bearer token")
		return
	}
	raw := strings.TrimSpace(auth[7:])
	if p.cfg.Keys == nil {
		writeOAuthError(w, http.StatusInternalServerError, "server_error", "keys unavailable")
		return
	}
	claims, err := p.cfg.Keys.ParseAndValidate(raw, p.cfg.Issuer, p.cfg.Audience)
	if err != nil {
		// Also accept tokens with client audience.
		claims, err = p.cfg.Keys.ParseAndValidate(raw, p.cfg.Issuer, "")
		if err != nil {
			writeOAuthError(w, http.StatusUnauthorized, "invalid_token", err.Error())
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"sub": claims.Subject,
		"sid": claims.Session,
		"tid": claims.Tenant,
	})
}

func verifyPKCE(verifier, challenge, method string) bool {
	if verifier == "" || challenge == "" {
		return false
	}
	switch method {
	case "plain":
		return subtleEqual(verifier, challenge)
	case "S256":
		sum := sha256.Sum256([]byte(verifier))
		computed := base64.RawURLEncoding.EncodeToString(sum[:])
		return subtleEqual(computed, challenge)
	default:
		return false
	}
}

func subtleEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeOAuthError(w http.ResponseWriter, status int, code, desc string) {
	writeJSON(w, status, map[string]string{
		"error":             code,
		"error_description": desc,
	})
}

// ErrNotConfigured is returned when keys are missing.
var ErrNotConfigured = errors.New("oidc: not configured")
