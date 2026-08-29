// Package httpclients implements BFF ports via downstream HTTP APIs.
package httpclients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nexora/bff-customer/internal/domain"
	"github.com/nexora/bff-customer/internal/reqctx"
)

// Base is a shared HTTP caller with tenant header propagation.
type Base struct {
	HTTP    *http.Client
	BaseURL string
}

func newBase(baseURL string) Base {
	return Base{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

type apiErrBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func mapHTTPStatus(status int, body []byte) error {
	var ae apiErrBody
	_ = json.Unmarshal(body, &ae)
	code := strings.ToLower(ae.Error.Code)
	switch {
	case status == http.StatusUnauthorized || code == "unauthorized":
		return domain.ErrUnauthorized
	case status == http.StatusNotFound || code == "not_found":
		return domain.ErrNotFound
	case status == http.StatusConflict || code == "conflict":
		return domain.ErrConflict
	case status == http.StatusBadRequest || code == "invalid_argument":
		return domain.ErrInvalidArgument
	default:
		return domain.ErrUpstream
	}
}

func (b Base) do(ctx context.Context, method, path, tenantID string, reqBody any, out any) error {
	if b.BaseURL == "" {
		return domain.ErrUpstream
	}
	var rdr io.Reader
	if reqBody != nil {
		raw, err := json.Marshal(reqBody)
		if err != nil {
			return domain.ErrInvalidArgument
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, b.BaseURL+path, rdr)
	if err != nil {
		return domain.ErrUpstream
	}
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if tenantID != "" {
		req.Header.Set("X-Tenant-Id", tenantID)
	}
	if rid := reqctx.RequestID(ctx); rid != "" {
		req.Header.Set("X-Request-Id", rid)
	}
	if uid := reqctx.UserID(ctx); uid != "" {
		req.Header.Set("X-Nexora-User", uid)
	}
	resp, err := b.HTTP.Do(req)
	if err != nil {
		return domain.ErrUpstream
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return mapHTTPStatus(resp.StatusCode, body)
	}
	if out == nil || len(body) == 0 || resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return domain.ErrUpstream
	}
	return nil
}

func (b Base) get(ctx context.Context, path, tenantID string, out any) error {
	return b.do(ctx, http.MethodGet, path, tenantID, nil, out)
}

func (b Base) post(ctx context.Context, path, tenantID string, reqBody, out any) error {
	return b.do(ctx, http.MethodPost, path, tenantID, reqBody, out)
}

func (b Base) patch(ctx context.Context, path, tenantID string, reqBody, out any) error {
	return b.do(ctx, http.MethodPatch, path, tenantID, reqBody, out)
}

func asString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case fmt.Stringer:
		return t.String()
	case float64:
		return fmt.Sprintf("%.0f", t)
	default:
		if v == nil {
			return ""
		}
		return fmt.Sprint(v)
	}
}

func asInt64(v any) int64 {
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	case json.Number:
		n, _ := t.Int64()
		return n
	default:
		return 0
	}
}

func asFloat64(v any) float64 {
	switch t := v.(type) {
	case float64:
		return t
	case int64:
		return float64(t)
	case int:
		return float64(t)
	default:
		return 0
	}
}

func asBool(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	default:
		return false
	}
}
