package httpclients

import (
	"bytes"
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
	OrderURL      string
	LiveOpsURL    string
	QualityURL    string
	CatalogURL    string
	CrmURL        string
	LedgerURL     string
	PromoURL      string
	PricingURL    string
	InventoryURL  string
	IdentityURL   string
	ProfileURL    string
	SettlementURL string
	NotifyURL     string
	LoyaltyURL    string
	AIURL         string
	TrackingURL   string
}

func ConfigFromEnv() Config {
	return Config{
		OrderURL:      env("ORDER_URL", "http://localhost:8086"),
		LiveOpsURL:    env("LIVEOPS_URL", "http://localhost:8116"),
		QualityURL:    env("QUALITY_URL", "http://localhost:8118"),
		CatalogURL:    env("CATALOG_URL", "http://localhost:8082"),
		CrmURL:        env("CRM_URL", "http://localhost:8102"),
		LedgerURL:     env("LEDGER_URL", "http://localhost:8091"),
		PromoURL:      env("PROMO_URL", "http://localhost:8094"),
		PricingURL:    env("PRICING_URL", "http://localhost:8095"),
		InventoryURL:  env("INVENTORY_URL", "http://localhost:8083"),
		IdentityURL:   env("IDENTITY_URL", "http://localhost:8081"),
		ProfileURL:    env("PROFILE_URL", "http://localhost:8082"),
		SettlementURL: env("SETTLEMENT_URL", "http://localhost:8092"),
		NotifyURL:     env("NOTIFICATION_URL", "http://localhost:8101"),
		LoyaltyURL:    env("LOYALTY_URL", "http://localhost:8093"),
		AIURL:         env("AI_URL", "http://localhost:8106"),
		TrackingURL:   env("TRACKING_URL", "http://localhost:8098"),
	}
}

func env(k, f string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return f
}

type Client struct {
	base          string
	http          *http.Client
	internalToken string
	tokenHeader   string
}

func New(base string) *Client {
	return &Client{base: strings.TrimRight(base, "/"), http: &http.Client{Timeout: 15 * time.Second}}
}

func NewLedger(base string) *Client {
	c := New(base)
	c.internalToken = strings.TrimSpace(os.Getenv("LEDGER_INTERNAL_TOKEN"))
	c.tokenHeader = "X-Ledger-Internal-Token"
	return c
}

func (c *Client) get(ctx context.Context, path, tenant string, out any) error {
	return c.do(ctx, http.MethodGet, path, tenant, nil, out)
}

func (c *Client) post(ctx context.Context, path, tenant string, in, out any) error {
	return c.do(ctx, http.MethodPost, path, tenant, in, out)
}

func (c *Client) patch(ctx context.Context, path, tenant string, in, out any) error {
	return c.do(ctx, http.MethodPatch, path, tenant, in, out)
}

func (c *Client) do(ctx context.Context, method, path, tenant string, in, out any) error {
	var body io.Reader
	if in != nil {
		raw, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if tenant != "" {
		req.Header.Set("X-Tenant-Id", tenant)
	}
	if c.internalToken != "" && c.tokenHeader != "" {
		req.Header.Set(c.tokenHeader, c.internalToken)
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

func (c OrderClient) Cancel(ctx context.Context, tenant, id, reason string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/orders/"+url.PathEscape(id)+"/cancel", tenant, map[string]any{"reason": reason}, &out)
	return out, err
}

func (c OrderClient) Refund(ctx context.Context, tenant, id string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/orders/"+url.PathEscape(id)+"/refunds", tenant, body, &out)
	return out, err
}

func (c OrderClient) DispatchEvent(ctx context.Context, tenant, id, eventType, courierRef string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/orders/"+url.PathEscape(id)+"/events/dispatch", tenant, map[string]any{
		"eventType": eventType, "courierRef": courierRef,
	}, &out)
	return out, err
}

type LiveOpsClient struct{ *Client }

func (c LiveOpsClient) SetFlag(ctx context.Context, tenant, key string, enabled bool) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/liveops/flags", tenant, map[string]any{"key": key, "enabled": enabled}, &out)
	return out, err
}

func (c LiveOpsClient) ListFlags(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/liveops/flags", tenant, &out)
	return out, err
}

type CatalogClient struct{ *Client }

func (c CatalogClient) ListProducts(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	path := "/v1/catalog/products"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}

func (c CatalogClient) GetProduct(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/catalog/products/"+url.PathEscape(id), tenant, &out)
	return out, err
}

type CRMClient struct{ *Client }

func (c CRMClient) ListTickets(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/crm/tickets", tenant, &out)
	return out, err
}

func (c CRMClient) GetTicket(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/crm/tickets/"+url.PathEscape(id), tenant, &out)
	return out, err
}

func (c CRMClient) EscalateTicket(ctx context.Context, tenant, id, reason string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/crm/tickets/"+url.PathEscape(id)+"/escalate", tenant, map[string]any{"reason": reason}, &out)
	return out, err
}

func (c CRMClient) ResolveTicket(ctx context.Context, tenant, id, note string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/crm/tickets/"+url.PathEscape(id)+"/resolve", tenant, map[string]any{"note": note}, &out)
	return out, err
}

type LedgerClient struct{ *Client }

func (c LedgerClient) ListJournals(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/ledger/journals", tenant, &out)
	return out, err
}

type PromoClient struct{ *Client }

func (c PromoClient) ListCampaigns(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	path := "/v1/promo/campaigns"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}

func (c PromoClient) GetCampaign(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/promo/campaigns/"+url.PathEscape(id), tenant, &out)
	return out, err
}

func (c PromoClient) CreateCampaign(ctx context.Context, tenant string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/promo/campaigns", tenant, body, &out)
	return out, err
}

func (c PromoClient) ListCoupons(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	path := "/v1/promo/coupons"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}

func (c PromoClient) GetCoupon(ctx context.Context, tenant, code string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/promo/coupons/"+url.PathEscape(code), tenant, &out)
	return out, err
}

func (c PromoClient) CreateCoupon(ctx context.Context, tenant string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/promo/coupons", tenant, body, &out)
	return out, err
}

func (c PromoClient) UpdateCoupon(ctx context.Context, tenant, code string, body map[string]any) (map[string]any, error) {
	var out map[string]any
	err := c.patch(ctx, "/v1/promo/coupons/"+url.PathEscape(code), tenant, body, &out)
	return out, err
}

type PricingClient struct{ *Client }

func (c PricingClient) AdminList(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/pricing/admin", tenant, &out)
	return out, err
}

type InventoryClient struct{ *Client }

func (c InventoryClient) ListWarehouses(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/inventory/warehouses", tenant, &out)
	return out, err
}

func (c InventoryClient) GetWarehouse(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/inventory/warehouses/"+url.PathEscape(id), tenant, &out)
	return out, err
}

func (c InventoryClient) ListStock(ctx context.Context, tenant, warehouseID string) (map[string]any, error) {
	path := "/v1/inventory/stock?warehouseId=" + url.QueryEscape(warehouseID) + "&limit=200"
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}

type IdentityClient struct{ *Client }

func (c IdentityClient) ListRoles(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/identity/roles", tenant, &out)
	return out, err
}

type ProfileClient struct{ *Client }

func (c ProfileClient) Search(ctx context.Context, tenant, q string, limit int) (map[string]any, error) {
	if limit <= 0 {
		limit = 50
	}
	path := "/v1/profile/admin/search?tenantId=" + url.QueryEscape(tenant) + "&limit=" + fmt.Sprint(limit)
	if q != "" {
		path += "&q=" + url.QueryEscape(q)
	}
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}

type SettlementClient struct{ *Client }

func (c SettlementClient) ListBatches(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/settlements/batches", tenant, &out)
	return out, err
}

func (c SettlementClient) Approve(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/settlements/batches/"+url.PathEscape(id)+"/approve", tenant, map[string]any{}, &out)
	return out, err
}

func (c SettlementClient) Execute(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/settlements/batches/"+url.PathEscape(id)+"/execute", tenant, map[string]any{}, &out)
	return out, err
}

type NotifyClient struct{ *Client }

func (c NotifyClient) Inbox(ctx context.Context, tenant, principalID string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/notifications/inbox/"+url.PathEscape(principalID), tenant, &out)
	return out, err
}

func (c NotifyClient) MarkRead(ctx context.Context, tenant, id string) (map[string]any, error) {
	var out map[string]any
	err := c.post(ctx, "/v1/notifications/inbox/"+url.PathEscape(id)+"/read", tenant, map[string]any{}, &out)
	return out, err
}

func (c NotifyClient) AdminStats(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/notifications/admin/stats", tenant, &out)
	return out, err
}

type LoyaltyClient struct{ *Client }

func (c LoyaltyClient) ListRewards(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/loyalty/rewards", tenant, &out)
	return out, err
}

type AIClient struct{ *Client }

func (c AIClient) AdminStats(ctx context.Context, tenant string) (map[string]any, error) {
	var out map[string]any
	err := c.get(ctx, "/v1/ai/admin/stats", tenant, &out)
	return out, err
}

type TrackingClient struct{ *Client }

func (c TrackingClient) Nearby(ctx context.Context, tenant string, lat, lon float64) (map[string]any, error) {
	path := fmt.Sprintf("/v1/tracking/nearby?lat=%f&lon=%f&radiusM=5000&limit=50", lat, lon)
	var out map[string]any
	err := c.get(ctx, path, tenant, &out)
	return out, err
}
