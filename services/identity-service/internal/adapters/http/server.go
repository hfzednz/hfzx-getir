// Package httpadapter exposes the identity-service REST API (stdlib ServeMux, Go 1.22+).
package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/nexora/identity-service/internal/adapters/http/oidc"
	"github.com/nexora/identity-service/internal/app"
	"github.com/nexora/identity-service/internal/ratelimit"
)

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	OIDC               *oidc.Provider
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

// NewHandler returns a fully wired http.Handler.
func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{
		Deps:  cfg.Deps,
		Ready: cfg.Ready,
		Live:  cfg.Live,
	}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /ready", h.ready)
	mux.HandleFunc("GET /v1/identity/health", h.health)
	mux.HandleFunc("GET /v1/identity/ready", h.ready)

	const base = "/v1/identity"

	mux.HandleFunc("POST "+base+"/auth/otp/start", h.otpStart)
	mux.HandleFunc("POST "+base+"/auth/otp/verify", h.otpVerify)
	mux.HandleFunc("POST "+base+"/auth/password/login", h.passwordLogin)
	mux.HandleFunc("POST "+base+"/auth/magic-link/start", h.magicLinkStart)
	mux.HandleFunc("POST "+base+"/auth/magic-link/consume", h.magicLinkConsume)
	mux.HandleFunc("POST "+base+"/auth/social/start", h.socialStart)
	mux.HandleFunc("POST "+base+"/auth/social/callback", h.socialCallback)
	mux.HandleFunc("POST "+base+"/auth/webauthn/register/options", h.webauthnRegisterOptions)
	mux.HandleFunc("POST "+base+"/auth/webauthn/register/verify", h.webauthnRegisterVerify)
	mux.HandleFunc("POST "+base+"/auth/webauthn/authenticate/options", h.webauthnAuthOptions)
	mux.HandleFunc("POST "+base+"/auth/webauthn/authenticate/verify", h.webauthnAuthVerify)
	mux.HandleFunc("POST "+base+"/auth/guest", h.guest)

	mux.HandleFunc("POST "+base+"/mfa/totp/enroll", h.mfaTotpEnroll)
	mux.HandleFunc("POST "+base+"/mfa/challenge", h.mfaChallenge)
	mux.HandleFunc("POST "+base+"/mfa/verify", h.mfaVerify)

	mux.HandleFunc("POST "+base+"/token/refresh", h.tokenRefresh)
	mux.HandleFunc("POST "+base+"/token/revoke", h.tokenRevoke)
	mux.HandleFunc("POST "+base+"/token/introspect", h.tokenIntrospect)

	mux.HandleFunc("GET "+base+"/sessions", h.listSessions)
	mux.HandleFunc("POST "+base+"/sessions/{id}/revoke", h.revokeSession)
	mux.HandleFunc("GET "+base+"/devices", h.listDevices)
	mux.HandleFunc("POST "+base+"/devices/{id}/trust", h.trustDevice)
	mux.HandleFunc("POST "+base+"/devices/{id}/revoke", h.revokeDevice)

	mux.HandleFunc("GET "+base+"/principals", h.listPrincipals)
	mux.HandleFunc("POST "+base+"/principals", h.createPrincipal)
	mux.HandleFunc("GET "+base+"/principals/{id}", h.getPrincipal)
	mux.HandleFunc("GET "+base+"/principals/{id}/roles", h.listPrincipalRoles)
	mux.HandleFunc("POST "+base+"/principals/{id}/roles", h.assignPrincipalRole)

	mux.HandleFunc("GET "+base+"/roles", h.listRoles)
	mux.HandleFunc("GET "+base+"/permissions", h.listPermissions)

	mux.HandleFunc("POST "+base+"/privacy/export", h.privacyExport)
	mux.HandleFunc("POST "+base+"/privacy/delete", h.privacyDelete)

	if cfg.OIDC != nil {
		cfg.OIDC.Mount(mux)
	}

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
