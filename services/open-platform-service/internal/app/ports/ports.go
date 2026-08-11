package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/open-platform-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

type AppRepo interface {
	Save(ctx context.Context, a domain.DeveloperApp) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.DeveloperApp, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.DeveloperApp, error)
}

type ApiKeyRepo interface {
	Save(ctx context.Context, k domain.ApiKey) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ApiKey, error)
	ListByApp(ctx context.Context, tenantID, appID uuid.UUID) ([]domain.ApiKey, error)
}

type CatalogRepo interface {
	Save(ctx context.Context, e domain.CatalogEntry) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.CatalogEntry, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.CatalogEntry, error)
}

type VersionRepo interface {
	Save(ctx context.Context, v domain.ApiVersion) error
	ListByCatalog(ctx context.Context, tenantID uuid.UUID, catalogKey string) ([]domain.ApiVersion, error)
}

type PolicyRepo interface {
	Save(ctx context.Context, p domain.GatewayPolicy) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.GatewayPolicy, error)
}

type WebhookRepo interface {
	Save(ctx context.Context, w domain.WebhookEndpoint) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.WebhookEndpoint, error)
	ListByApp(ctx context.Context, tenantID, appID uuid.UUID) ([]domain.WebhookEndpoint, error)
	ListActiveForEvent(ctx context.Context, tenantID uuid.UUID, eventType string) ([]domain.WebhookEndpoint, error)
}

type DeliveryRepo interface {
	Save(ctx context.Context, d domain.WebhookDelivery) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.WebhookDelivery, error)
	ListPending(ctx context.Context, limit int) ([]domain.WebhookDelivery, error)
	ListByEndpoint(ctx context.Context, tenantID, endpointID uuid.UUID) ([]domain.WebhookDelivery, error)
}

type SdkRepo interface {
	Save(ctx context.Context, s domain.SdkPackage) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.SdkPackage, error)
}

type IntegrationRepo interface {
	Save(ctx context.Context, i domain.IntegrationConnector) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.IntegrationConnector, error)
}

type UsageRepo interface {
	Save(ctx context.Context, u domain.UsageCounter) error
	Get(ctx context.Context, tenantID, appID uuid.UUID, day string) (domain.UsageCounter, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.UsageCounter, error)
}

type WebhookHTTPClient interface {
	Post(ctx context.Context, url string, headers map[string]string, body []byte) (status int, err error)
}

type IdentityClient interface {
	RegisterOAuthClient(ctx context.Context, tenantID uuid.UUID, name string, scopes []string) (clientID string, err error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
