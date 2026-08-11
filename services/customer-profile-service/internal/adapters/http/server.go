// Package httpadapter exposes the customer-profile-service REST API (stdlib ServeMux, Go 1.22+).
package httpadapter

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/nexora/customer-profile-service/internal/app"
)

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
	Ready              func(*http.Request) error
	Live               func(*http.Request) error
}

// Handler holds dependencies for route handlers.
type Handler struct {
	Deps  *app.Deps
	Ready func(*http.Request) error
	Live  func(*http.Request) error
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
	mux.HandleFunc("GET /v1/profile/health", h.health)
	mux.HandleFunc("GET /v1/profile/ready", h.ready)

	const base = "/v1/profile"

	// Me (trusted principal from X-Nexora-User)
	mux.HandleFunc("GET "+base+"/me", h.getMe)
	mux.HandleFunc("PATCH "+base+"/me", h.patchMe)

	// Profiles
	mux.HandleFunc("POST "+base+"/customers", h.provisionCustomer)
	mux.HandleFunc("GET "+base+"/customers/{id}", h.getCustomer)
	mux.HandleFunc("PATCH "+base+"/customers/{id}", h.patchCustomer)

	// Addresses
	mux.HandleFunc("GET "+base+"/customers/{id}/addresses", h.listAddresses)
	mux.HandleFunc("POST "+base+"/customers/{id}/addresses", h.createAddress)
	mux.HandleFunc("PATCH "+base+"/customers/{id}/addresses/{addressId}", h.patchAddress)
	mux.HandleFunc("DELETE "+base+"/customers/{id}/addresses/{addressId}", h.deleteAddress)
	mux.HandleFunc("POST "+base+"/customers/{id}/addresses/{addressId}/default", h.setDefaultAddress)

	// Preferences
	mux.HandleFunc("GET "+base+"/customers/{id}/preferences", h.getPreferences)
	mux.HandleFunc("PUT "+base+"/customers/{id}/preferences", h.putPreferences)

	// Avatar
	mux.HandleFunc("POST "+base+"/customers/{id}/avatar", h.setAvatar)
	mux.HandleFunc("DELETE "+base+"/customers/{id}/avatar", h.deleteAvatar)

	// Tags
	mux.HandleFunc("GET "+base+"/customers/{id}/tags", h.listTags)
	mux.HandleFunc("POST "+base+"/customers/{id}/tags", h.addTag)
	mux.HandleFunc("DELETE "+base+"/customers/{id}/tags/{tagId}", h.removeTag)

	// Household
	mux.HandleFunc("GET "+base+"/customers/{id}/household", h.getHousehold)
	mux.HandleFunc("POST "+base+"/customers/{id}/household", h.createHousehold)
	mux.HandleFunc("POST "+base+"/customers/{id}/household/members", h.addHouseholdMember)
	mux.HandleFunc("PATCH "+base+"/customers/{id}/household/sharing", h.updateHouseholdSharing)

	// Consents
	mux.HandleFunc("GET "+base+"/customers/{id}/consents", h.listConsents)
	mux.HandleFunc("PUT "+base+"/customers/{id}/consents", h.setConsent)

	// CRM
	mux.HandleFunc("GET "+base+"/customers/{id}/360", h.getCustomer360)
	mux.HandleFunc("GET "+base+"/customers/{id}/notes", h.listNotes)
	mux.HandleFunc("POST "+base+"/customers/{id}/notes", h.addNote)
	mux.HandleFunc("GET "+base+"/customers/{id}/timeline", h.listTimeline)
	mux.HandleFunc("POST "+base+"/customers/{id}/timeline", h.appendTimeline)

	// Segments
	mux.HandleFunc("GET "+base+"/segments", h.listSegments)
	mux.HandleFunc("POST "+base+"/segments", h.upsertSegment)
	mux.HandleFunc("GET "+base+"/customers/{id}/segments", h.listProfileSegments)
	mux.HandleFunc("POST "+base+"/customers/{id}/segments/{segmentId}/assign", h.assignSegment)
	mux.HandleFunc("POST "+base+"/customers/{id}/segments/{segmentId}/evaluate", h.evaluateSegment)

	// Personalization
	mux.HandleFunc("GET "+base+"/customers/{id}/personalization", h.getPersonalization)
	mux.HandleFunc("PUT "+base+"/customers/{id}/personalization", h.putPersonalization)

	// AI model
	mux.HandleFunc("GET "+base+"/customers/{id}/ai-model", h.getAIModel)
	mux.HandleFunc("POST "+base+"/customers/{id}/ai-model/recompute", h.recomputeAIModel)

	// Privacy
	mux.HandleFunc("POST "+base+"/customers/{id}/privacy/export", h.requestExport)
	mux.HandleFunc("POST "+base+"/customers/{id}/privacy/delete", h.requestDeletion)

	// Admin
	mux.HandleFunc("GET "+base+"/admin/search", h.adminSearch)
	mux.HandleFunc("POST "+base+"/admin/merge", h.adminMerge)
	mux.HandleFunc("GET "+base+"/admin/duplicates", h.adminDuplicates)

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(newMemoryLimiter(), cfg.RateLimitPerMinute),
		trustedPrincipalMiddleware,
	)
}

// NewServer builds an *http.Server with sensible timeouts.
func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}
