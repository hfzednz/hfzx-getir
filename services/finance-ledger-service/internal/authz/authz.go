// Package authz enforces backend-authoritative JWT RBAC at HTTP edges.
package authz

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	ErrNoToken     = errors.New("missing access token")
	ErrInvalidTok  = errors.New("invalid access token")
	ErrForbidden   = errors.New("forbidden")
	ErrTenant      = errors.New("tenant mismatch")
	ErrTicket      = errors.New("invalid sse ticket")
)

type ctxKey struct{}

// Principal is the authenticated access-token subject.
type Principal struct {
	ID       string
	TenantID string
	Roles    []string
}

// Validator validates a Bearer access token.
type Validator interface {
	Validate(ctx context.Context, token string) (Principal, error)
}

// Static is a test double: raw token string → principal.
type Static map[string]Principal

func (s Static) Validate(_ context.Context, token string) (Principal, error) {
	p, ok := s[token]
	if !ok {
		return Principal{}, ErrInvalidTok
	}
	return p, nil
}

// Introspector validates tokens via identity-service introspect.
type Introspector struct {
	IdentityURL string
	HTTP        *http.Client
}

// FromEnv builds an introspector against IDENTITY_URL.
func FromEnv() Validator {
	base := strings.TrimRight(os.Getenv("IDENTITY_URL"), "/")
	if base == "" {
		base = "http://127.0.0.1:8081"
	}
	return &Introspector{IdentityURL: base, HTTP: &http.Client{Timeout: 5 * time.Second}}
}

func (i *Introspector) Validate(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrNoToken
	}
	if i == nil || i.IdentityURL == "" {
		return Principal{}, ErrInvalidTok
	}
	body, _ := json.Marshal(map[string]string{"token": token})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, i.IdentityURL+"/v1/identity/token/introspect", strings.NewReader(string(body)))
	if err != nil {
		return Principal{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := i.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return Principal{}, ErrInvalidTok
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	var out struct {
		Active bool     `json:"active"`
		Sub    string   `json:"sub"`
		Tid    string   `json:"tid"`
		Roles  []string `json:"roles"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || !out.Active || out.Sub == "" {
		return Principal{}, ErrInvalidTok
	}
	return Principal{ID: out.Sub, TenantID: out.Tid, Roles: out.Roles}, nil
}

// Rule matches a URL prefix to allowed roles. First match wins.
type Rule struct {
	Prefix string
	Roles  []string
}

// Options configures Gate.
type Options struct {
	Public []string
	Rules  []Rule
}

// Gate returns middleware that requires a valid JWT and an allowed role.
func Gate(v Validator, opt Options) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			if internalTok := strings.TrimSpace(os.Getenv("LEDGER_INTERNAL_TOKEN")); internalTok != "" {
				if hmac.Equal([]byte(r.Header.Get("X-Ledger-Internal-Token")), []byte(internalTok)) {
					next.ServeHTTP(w, r)
					return
				}
			}
			path := r.URL.Path
			for _, p := range opt.Public {
				if path == p || strings.HasPrefix(path, p) {
					next.ServeHTTP(w, r)
					return
				}
			}
			if v == nil {
				writeAuth(w, http.StatusUnauthorized, "unauthorized", "authentication required")
				return
			}
			tok := BearerToken(r)
			if tok == "" {
				writeAuth(w, http.StatusUnauthorized, "unauthorized", "missing access token")
				return
			}
			p, err := v.Validate(r.Context(), tok)
			if err != nil {
				writeAuth(w, http.StatusUnauthorized, "unauthorized", "invalid access token")
				return
			}
			allowed := rolesFor(path, opt.Rules)
			if len(allowed) == 0 || !hasAny(p.Roles, allowed) {
				writeAuth(w, http.StatusForbidden, "forbidden", "insufficient role")
				return
			}
			hdrTid := strings.TrimSpace(r.Header.Get("X-Tenant-Id"))
			if hdrTid == "" {
				hdrTid = strings.TrimSpace(r.Header.Get("X-Nexora-Tenant"))
			}
			if hdrTid != "" && p.TenantID != "" && hdrTid != p.TenantID {
				writeAuth(w, http.StatusNotFound, "not_found", "not found")
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func rolesFor(path string, rules []Rule) []string {
	for _, rule := range rules {
		if rule.Prefix == path || strings.HasPrefix(path, rule.Prefix) {
			return rule.Roles
		}
	}
	return nil
}

func hasAny(have, need []string) bool {
	set := make(map[string]struct{}, len(have))
	for _, h := range have {
		set[h] = struct{}{}
	}
	for _, n := range need {
		if _, ok := set[n]; ok {
			return true
		}
	}
	return false
}

// PrincipalFrom returns the authenticated principal.
func PrincipalFrom(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}

// BearerToken extracts the Authorization Bearer token.
func BearerToken(r *http.Request) string {
	h := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(h) >= 7 && strings.EqualFold(h[:7], "Bearer ") {
		return strings.TrimSpace(h[7:])
	}
	return ""
}

func writeAuth(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": code, "message": msg, "retriable": false},
	})
}

// SSEClaims is a short-lived topic subscription ticket (not a JWT).
type SSEClaims struct {
	Tenant string `json:"tid"`
	Sub    string `json:"sub"`
	Topic  string `json:"topic"`
	Exp    int64  `json:"exp"`
}

// IssueSSETicket mints an HMAC ticket bound to tenant, principal, and topic.
func IssueSSETicket(secret, tenant, sub, topic string, ttl time.Duration) (string, error) {
	if secret == "" || tenant == "" || sub == "" || topic == "" {
		return "", ErrTicket
	}
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	c := SSEClaims{Tenant: tenant, Sub: sub, Topic: topic, Exp: time.Now().Add(ttl).Unix()}
	raw, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	return base64.RawURLEncoding.EncodeToString(raw) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// ParseSSETicket validates an HMAC ticket.
func ParseSSETicket(secret, ticket string) (SSEClaims, error) {
	if secret == "" || ticket == "" {
		return SSEClaims{}, ErrTicket
	}
	parts := strings.Split(ticket, ".")
	if len(parts) != 2 {
		return SSEClaims{}, ErrTicket
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return SSEClaims{}, ErrTicket
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return SSEClaims{}, ErrTicket
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return SSEClaims{}, ErrTicket
	}
	var c SSEClaims
	if err := json.Unmarshal(raw, &c); err != nil {
		return SSEClaims{}, ErrTicket
	}
	if c.Exp < time.Now().Unix() || c.Topic == "" || c.Sub == "" {
		return SSEClaims{}, ErrTicket
	}
	return c, nil
}
