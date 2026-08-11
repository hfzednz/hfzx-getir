package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AppKind string

const (
	AppPublic   AppKind = "public"
	AppPartner  AppKind = "partner"
	AppInternal AppKind = "internal"
)

type DeveloperApp struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	Name         string    `json:"name"`
	Kind         AppKind   `json:"kind"`
	OwnerEmail   string    `json:"ownerEmail"`
	Scopes       []string  `json:"scopes"`
	OAuthClientID string   `json:"oauthClientId,omitempty"` // opaque ref to identity-service
	Sandbox      bool      `json:"sandbox"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
}

type ApiKey struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	AppID      uuid.UUID  `json:"appId"`
	Name       string     `json:"name"`
	Prefix     string     `json:"prefix"`
	SecretHash string     `json:"-"`
	SecretOnce string     `json:"secret,omitempty"` // only on create
	Scopes     []string   `json:"scopes"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	Revoked    bool       `json:"revoked"`
}

type Surface string

const (
	SurfacePublic  Surface = "public"
	SurfacePrivate Surface = "private"
	SurfacePartner Surface = "partner"
)

type CatalogEntry struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Key         string    `json:"key"` // e.g. orders
	Title       string    `json:"title"`
	Surface     Surface   `json:"surface"`
	BasePath    string    `json:"basePath"`
	ServiceRef  string    `json:"serviceRef"` // e.g. order-service — opaque
	OpenAPIPath string    `json:"openApiPath"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ApiVersion struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenantId"`
	CatalogKey  string     `json:"catalogKey"`
	Version     string     `json:"version"` // v1, v1.2
	Status      string     `json:"status"` // current|deprecated|retired
	ReleasedAt  time.Time  `json:"releasedAt"`
	DeprecatedAt *time.Time `json:"deprecatedAt,omitempty"`
	Notes       string     `json:"notes"`
}

type GatewayPolicy struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenantId"`
	Name            string    `json:"name"`
	RoutePrefix     string    `json:"routePrefix"`
	TargetService   string    `json:"targetService"`
	Version         string    `json:"version"`
	RateLimitRPM    int       `json:"rateLimitRpm"`
	QuotaDaily      int       `json:"quotaDaily"`
	CanaryPercent   int       `json:"canaryPercent"`
	RequireScopes   []string  `json:"requireScopes"`
	CacheTTLSeconds int       `json:"cacheTtlSeconds"`
	Active          bool      `json:"active"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type WebhookEndpoint struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	AppID     uuid.UUID `json:"appId"`
	URL       string    `json:"url"`
	Secret    string    `json:"-"`
	Events    []string  `json:"events"`
	Version   string    `json:"version"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

type DeliveryStatus string

const (
	DeliveryPending DeliveryStatus = "pending"
	DeliverySuccess DeliveryStatus = "success"
	DeliveryFailed  DeliveryStatus = "failed"
	DeliveryDLQ     DeliveryStatus = "dlq"
)

type WebhookDelivery struct {
	ID           uuid.UUID      `json:"id"`
	TenantID     uuid.UUID      `json:"tenantId"`
	EndpointID   uuid.UUID      `json:"endpointId"`
	EventType    string         `json:"eventType"`
	Payload      map[string]any `json:"payload"`
	Status       DeliveryStatus `json:"status"`
	Attempts     int            `json:"attempts"`
	LastStatus   int            `json:"lastHttpStatus"`
	LastError    string         `json:"lastError,omitempty"`
	Signature    string         `json:"signature"`
	CreatedAt    time.Time      `json:"createdAt"`
	DeliveredAt  *time.Time     `json:"deliveredAt,omitempty"`
}

type SdkLanguage string

const (
	SdkGo      SdkLanguage = "go"
	SdkNode    SdkLanguage = "nodejs"
	SdkPython  SdkLanguage = "python"
	SdkFlutter SdkLanguage = "flutter"
	SdkJava    SdkLanguage = "java"
	SdkDotNet  SdkLanguage = "dotnet"
	SdkSwift   SdkLanguage = "swift"
	SdkKotlin  SdkLanguage = "kotlin"
	SdkWeb     SdkLanguage = "web"
)

type SdkPackage struct {
	ID        uuid.UUID   `json:"id"`
	TenantID  uuid.UUID   `json:"tenantId"`
	Language  SdkLanguage `json:"language"`
	Name      string      `json:"name"`
	Version   string      `json:"version"`
	RepoPath  string      `json:"repoPath"`
	Status    string      `json:"status"` // published|generating
	CreatedAt time.Time   `json:"createdAt"`
}

type IntegrationKind string

const (
	IntERP     IntegrationKind = "erp"
	IntCRM     IntegrationKind = "crm"
	IntSMS     IntegrationKind = "sms"
	IntEmail   IntegrationKind = "email"
	IntPush    IntegrationKind = "push"
	IntPayment IntegrationKind = "payment_provider"
	IntShipping IntegrationKind = "shipping"
	IntMaps    IntegrationKind = "maps"
	IntAI      IntegrationKind = "ai"
	IntIdP     IntegrationKind = "identity_provider"
	IntAccounting IntegrationKind = "accounting"
)

type IntegrationConnector struct {
	ID         uuid.UUID       `json:"id"`
	TenantID   uuid.UUID       `json:"tenantId"`
	AppID      uuid.UUID       `json:"appId"`
	Kind       IntegrationKind `json:"kind"`
	Provider   string          `json:"provider"`
	ConfigRef  string          `json:"configRef"` // vault/opaque
	Status     string          `json:"status"` // connected|pending|error
	CreatedAt  time.Time       `json:"createdAt"`
}

type UsageCounter struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	AppID     uuid.UUID `json:"appId"`
	Day       string    `json:"day"` // YYYY-MM-DD
	Requests  int64     `json:"requests"`
	Errors    int64     `json:"errors"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func ValidateApp(a DeveloperApp) error {
	if a.TenantID == uuid.Nil || strings.TrimSpace(a.Name) == "" || a.Kind == "" {
		return ErrInvalidArgument
	}
	return nil
}

func HashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func SignWebhook(secret, body string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(body))
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func VerifyWebhook(secret, body, signature string) bool {
	expected := SignWebhook(secret, body)
	return hmac.Equal([]byte(expected), []byte(signature))
}

func NextRetryDelay(attempts int) time.Duration {
	if attempts <= 0 {
		return time.Second
	}
	d := time.Second * time.Duration(1<<uint(min(attempts, 6)))
	return d
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DefaultPublicCatalog() []CatalogEntry {
	return []CatalogEntry{
		{Key: "auth", Title: "Authentication", Surface: SurfacePublic, BasePath: "/v1/auth", ServiceRef: "identity-service", OpenAPIPath: "services/identity-service/api/openapi/openapi.yaml"},
		{Key: "customers", Title: "Customers", Surface: SurfacePublic, BasePath: "/v1/customers", ServiceRef: "customer-profile-service", OpenAPIPath: "services/customer-profile-service/api/openapi/openapi.yaml"},
		{Key: "products", Title: "Products / Catalog", Surface: SurfacePublic, BasePath: "/v1/catalog", ServiceRef: "catalog-service", OpenAPIPath: "services/catalog-service/api/openapi/openapi.yaml"},
		{Key: "inventory", Title: "Inventory", Surface: SurfacePublic, BasePath: "/v1/inventory", ServiceRef: "inventory-service", OpenAPIPath: "services/inventory-service/api/openapi/openapi.yaml"},
		{Key: "orders", Title: "Orders", Surface: SurfacePublic, BasePath: "/v1/orders", ServiceRef: "order-service", OpenAPIPath: "services/order-service/api/openapi/openapi.yaml"},
		{Key: "checkout", Title: "Checkout", Surface: SurfacePublic, BasePath: "/v1/checkout", ServiceRef: "checkout-service", OpenAPIPath: "services/checkout-service/api/openapi/openapi.yaml"},
		{Key: "payments", Title: "Payments", Surface: SurfacePublic, BasePath: "/v1/payments", ServiceRef: "payment-service", OpenAPIPath: "services/payment-service/api/openapi/openapi.yaml"},
		{Key: "wallet", Title: "Wallet", Surface: SurfacePublic, BasePath: "/v1/wallet", ServiceRef: "wallet-service", OpenAPIPath: "services/wallet-service/api/openapi/openapi.yaml"},
		{Key: "loyalty", Title: "Loyalty", Surface: SurfacePublic, BasePath: "/v1/loyalty", ServiceRef: "loyalty-service", OpenAPIPath: "services/loyalty-service/api/openapi/openapi.yaml"},
		{Key: "coupons", Title: "Coupons", Surface: SurfacePublic, BasePath: "/v1/promotions", ServiceRef: "promotion-service", OpenAPIPath: "services/promotion-service/api/openapi/openapi.yaml"},
		{Key: "delivery", Title: "Delivery / Dispatch", Surface: SurfacePublic, BasePath: "/v1/dispatch", ServiceRef: "dispatch-service", OpenAPIPath: "services/dispatch-service/api/openapi/openapi.yaml"},
		{Key: "tracking", Title: "Tracking", Surface: SurfacePublic, BasePath: "/v1/tracking", ServiceRef: "tracking-service", OpenAPIPath: "services/tracking-service/api/openapi/openapi.yaml"},
		{Key: "notifications", Title: "Notifications", Surface: SurfacePublic, BasePath: "/v1/notifications", ServiceRef: "notification-service", OpenAPIPath: "services/notification-service/api/openapi/openapi.yaml"},
		{Key: "reviews", Title: "Reviews", Surface: SurfacePublic, BasePath: "/v1/reviews", ServiceRef: "review-service", OpenAPIPath: "services/review-service/api/openapi/openapi.yaml"},
		{Key: "search", Title: "Search", Surface: SurfacePublic, BasePath: "/v1/search", ServiceRef: "search-service", OpenAPIPath: "services/search-service/api/openapi/openapi.yaml"},
		{Key: "support", Title: "Support", Surface: SurfacePublic, BasePath: "/v1/crm", ServiceRef: "crm-service", OpenAPIPath: "services/crm-service/api/openapi/openapi.yaml"},
		{Key: "analytics", Title: "Analytics", Surface: SurfacePublic, BasePath: "/v1/data", ServiceRef: "data-platform-service", OpenAPIPath: "services/data-platform-service/api/openapi/openapi.yaml"},
	}
}

func DefaultPrivateCatalog() []CatalogEntry {
	return []CatalogEntry{
		{Key: "admin", Title: "Admin BFF", Surface: SurfacePrivate, BasePath: "/v1/admin", ServiceRef: "bff-admin", OpenAPIPath: "services/bff-admin/api/openapi/openapi.yaml"},
		{Key: "warehouse", Title: "Warehouse", Surface: SurfacePrivate, BasePath: "/v1/warehouse", ServiceRef: "warehouse-service", OpenAPIPath: "services/warehouse-service/api/openapi/openapi.yaml"},
		{Key: "courier", Title: "Courier BFF", Surface: SurfacePrivate, BasePath: "/v1/courier", ServiceRef: "bff-courier", OpenAPIPath: "services/bff-courier/api/openapi/openapi.yaml"},
		{Key: "erp", Title: "ERP", Surface: SurfacePrivate, BasePath: "/v1/erp", ServiceRef: "erp-service", OpenAPIPath: "services/erp-service/api/openapi/openapi.yaml"},
		{Key: "finance", Title: "Finance Ledger", Surface: SurfacePrivate, BasePath: "/v1/finance", ServiceRef: "finance-ledger-service", OpenAPIPath: "services/finance-ledger-service/api/openapi/openapi.yaml"},
		{Key: "ops", Title: "Platform Ops", Surface: SurfacePrivate, BasePath: "/v1/platform", ServiceRef: "platform-ops-service", OpenAPIPath: "services/platform-ops-service/api/openapi/openapi.yaml"},
		{Key: "ai", Title: "AI Platform", Surface: SurfacePrivate, BasePath: "/v1/ai", ServiceRef: "ai-platform-service", OpenAPIPath: "services/ai-platform-service/api/openapi/openapi.yaml"},
	}
}

func DefaultPartnerCatalog() []CatalogEntry {
	return []CatalogEntry{
		{Key: "supplier", Title: "Supplier", Surface: SurfacePartner, BasePath: "/v1/supplier", ServiceRef: "supplier-service", OpenAPIPath: "services/supplier-service/api/openapi/openapi.yaml"},
		{Key: "marketplace", Title: "Marketplace Sellers", Surface: SurfacePartner, BasePath: "/v1/supplier/sellers", ServiceRef: "supplier-service", OpenAPIPath: "services/supplier-service/api/openapi/openapi.yaml"},
		{Key: "liveops", Title: "LiveOps Config", Surface: SurfacePartner, BasePath: "/v1/liveops", ServiceRef: "liveops-service", OpenAPIPath: "services/liveops-service/api/openapi/openapi.yaml"},
		{Key: "global", Title: "Globalization", Surface: SurfacePartner, BasePath: "/v1/global", ServiceRef: "global-service", OpenAPIPath: "services/global-service/api/openapi/openapi.yaml"},
	}
}
