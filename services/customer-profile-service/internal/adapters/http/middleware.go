package httpadapter

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/customer-profile-service/internal/domain"
	"github.com/nexora/customer-profile-service/internal/observability"
)

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	principalKey ctxKey = "principal"
)

// TrustedPrincipal is the gateway-injected identity (no IAM validation here).
type TrustedPrincipal struct {
	PrincipalID uuid.UUID
	TenantID    uuid.UUID
}

// RequestIDFromContext returns the request/trace id.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// PrincipalFromContext returns the trusted principal from X-Nexora-User.
func PrincipalFromContext(ctx context.Context) (TrustedPrincipal, bool) {
	p, ok := ctx.Value(principalKey).(TrustedPrincipal)
	return p, ok
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func withPrincipal(ctx context.Context, p TrustedPrincipal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

func chain(h http.Handler, mws ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(withRequestID(r.Context(), id)))
	})
}

func loggingMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			observability.Default.HTTPRequests.Add(1)
			log.Info("http.request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"durationMs", time.Since(start).Milliseconds(),
				"traceId", RequestIDFromContext(r.Context()),
				"remote", clientIP(r),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

func recoverMiddleware(log *slog.Logger) func(http.Handler) http.Handler {
	if log == nil {
		log = slog.Default()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("http.panic",
						"recover", rec,
						"stack", string(debug.Stack()),
						"traceId", RequestIDFromContext(r.Context()),
					)
					writeErr(w, r, errors.New("internal panic"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	allowAll := len(origins) == 0 || (len(origins) == 1 && origins[0] == "*")
	allowed := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		allowed[o] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" {
				if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Headers",
				"Authorization, Content-Type, X-Request-Id, X-Tenant-Id, X-Nexora-User, X-Nexora-Tenant")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// memoryLimiter is a simple per-key fixed-window rate limiter.
type memoryLimiter struct {
	mu      sync.Mutex
	windows map[string]*window
}

type window struct {
	count int
	reset time.Time
}

func newMemoryLimiter() *memoryLimiter {
	return &memoryLimiter{windows: make(map[string]*window)}
}

func (l *memoryLimiter) allow(key string, limit int, dur time.Duration) bool {
	if limit <= 0 {
		return true
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	w, ok := l.windows[key]
	if !ok || now.After(w.reset) {
		l.windows[key] = &window{count: 1, reset: now.Add(dur)}
		return true
	}
	if w.count >= limit {
		return false
	}
	w.count++
	return true
}

func rateLimitMiddleware(limiter *memoryLimiter, perMinute int) func(http.Handler) http.Handler {
	if limiter == nil || perMinute <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "ip:" + clientIP(r)
			if !limiter.allow(key, perMinute, time.Minute) {
				writeErr(w, r, domain.ErrRateLimited)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// trustedPrincipalMiddleware parses X-Nexora-User (+ tenant header) into context.
// It does not reject missing headers — /me handlers enforce presence.
func trustedPrincipalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawUser := strings.TrimSpace(r.Header.Get("X-Nexora-User"))
		rawTenant := strings.TrimSpace(r.Header.Get("X-Nexora-Tenant"))
		if rawTenant == "" {
			rawTenant = strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
		}
		if rawUser != "" {
			pid, err := uuid.Parse(rawUser)
			if err == nil {
				tp := TrustedPrincipal{PrincipalID: pid}
				if rawTenant != "" {
					if tid, err := uuid.Parse(rawTenant); err == nil {
						tp.TenantID = tid
					}
				}
				r = r.WithContext(withPrincipal(r.Context(), tp))
			}
		}
		next.ServeHTTP(w, r)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
