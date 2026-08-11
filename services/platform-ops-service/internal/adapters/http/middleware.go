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
	"github.com/nexora/platform-ops-service/internal/domain"
	"github.com/nexora/platform-ops-service/internal/ratelimit"
)

type ctxKey string

const (
	requestIDKey ctxKey = "request_id"
	tenantIDKey  ctxKey = "tenant_id"
)

func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

func TenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if v, ok := ctx.Value(tenantIDKey).(uuid.UUID); ok && v != uuid.Nil {
		return v, true
	}
	return uuid.Nil, false
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
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey, id)))
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
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), tenantIDKey, tid)))
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
			log.Info("http.request", "method", r.Method, "path", r.URL.Path, "status", rw.status,
				"durationMs", time.Since(start).Milliseconds(), "traceId", RequestIDFromContext(r.Context()))
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
					writeErr(w, r, domain.ErrConflict)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func corsMiddleware(origins []string) func(http.Handler) http.Handler {
	allowAll := len(origins) == 0 || (len(origins) == 1 && origins[0] == "*")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id, X-Tenant-Id, X-Nexora-Tenant")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				ip = strings.TrimSpace(strings.Split(xff, ",")[0])
			} else if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
				ip = host
			}
			ok, err := limiter.Allow(r.Context(), "ip:"+ip, perMinute, time.Minute)
			if err != nil || !ok {
				writeErr(w, r, domain.ErrConflict)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
