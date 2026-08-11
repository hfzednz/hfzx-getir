package httpadapter

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/location-service/internal/domain"
	"github.com/nexora/location-service/internal/ratelimit"
)

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	tenantIDKey  ctxKey = "tenant_id"
	userIDKey    ctxKey = "nexora_user"
)

// RequestIDFromContext returns the request/trace id.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

// TenantIDFromContext returns the tenant id from context.
func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if v, ok := ctx.Value(tenantIDKey).(uuid.UUID); ok && v != uuid.Nil {
		return v, true
	}
	return uuid.Nil, false
}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func withTenantID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, tenantIDKey, id)
}

func withUserID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey, id)
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

func tenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
		if raw == "" {
			raw = strings.TrimSpace(r.Header.Get("X-Nexora-Tenant"))
		}
		if raw == "" {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		tid, err := uuid.Parse(raw)
		if err != nil {
			writeErr(w, r, domain.ErrInvalidArgument)
			return
		}
		ctx := withTenantID(r.Context(), tid)
		if u := strings.TrimSpace(r.Header.Get("X-Nexora-User")); u != "" {
			if uid, err := uuid.Parse(u); err == nil {
				ctx = withUserID(ctx, uid)
			}
		}
		next.ServeHTTP(w, r.WithContext(ctx))
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
			// Privacy: never log full precise coords at info level.
			log.Info("http.request",
				"method", r.Method, "path", r.URL.Path, "status", rw.status,
				"durationMs", time.Since(start).Milliseconds(),
				"traceId", RequestIDFromContext(r.Context()),
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
					log.Error("http.panic", "recover", rec, "stack", string(debug.Stack()))
					writeErr(w, r, &simpleErr{msg: "internal panic"})
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type simpleErr struct{ msg string }

func (e *simpleErr) Error() string { return e.msg }

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
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id, X-Tenant-Id, X-Nexora-Tenant, X-Nexora-User, Idempotency-Key")
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

func rateLimitMiddleware(limiter ratelimit.Limiter, perMinute int) func(http.Handler) http.Handler {
	if limiter == nil || perMinute <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	window := time.Minute
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "ip:" + clientIP(r)
			ok, err := limiter.Allow(r.Context(), key, perMinute, window)
			if err != nil {
				writeErr(w, r, err)
				return
			}
			if !ok {
				writeErr(w, r, domain.ErrConflict)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
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
