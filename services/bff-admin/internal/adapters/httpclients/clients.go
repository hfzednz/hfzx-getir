package httpclients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type Config struct {
	OrderURL   string
	LiveOpsURL string
	QualityURL string
}

func ConfigFromEnv() Config {
	return Config{
		OrderURL:   env("ORDER_URL", "http://localhost:8086"),
		LiveOpsURL: env("LIVEOPS_URL", "http://localhost:8116"),
		QualityURL: env("QUALITY_URL", "http://localhost:8118"),
	}
}

func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}

type Client struct {
	base   string
	http   *http.Client
}

func New(base string) *Client {
	return &Client{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) get(ctx context.Context, path, tenant string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if tenant != "" {
		req.Header.Set("X-Tenant-Id", tenant)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upstream %s: %d %s", path, resp.StatusCode, string(b))
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

func (c *Client) post(ctx context.Context, path, tenant string, in, out any) error {
	body, _ := json.Marshal(in)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if tenant != "" {
		req.Header.Set("X-Tenant-Id", tenant)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upstream %s: %d %s", path, resp.StatusCode, string(b))
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

type OrderClient struct{ *Client }

func (c OrderClient) List(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	path := "/v1/orders"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}

func (c OrderClient) Get(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/orders/"+url.PathEscape(id), tenant, &out)
	return out, err
}

type LiveOpsClient struct{ *Client }

func (c LiveOpsClient) SetFlag(ctx context.Context, tenant, key string, enabled bool) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/liveops/flags/"+url.PathEscape(key), tenant, map[string]any{"enabled": enabled}, &out)
	return out, err
}
