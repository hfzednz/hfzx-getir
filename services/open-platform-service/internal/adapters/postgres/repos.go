package postgres

import "database/sql"

type Repos struct {
	Apps         *AppRepo
	Keys         *ApiKeyRepo
	Catalog      *CatalogRepo
	Versions     *VersionRepo
	Policies     *PolicyRepo
	Webhooks     *WebhookRepo
	Deliveries   *DeliveryRepo
	SDKs         *SdkRepo
	Integrations *IntegrationRepo
	Usage        *UsageRepo
	Outbox       *OutboxRepo
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Apps: &AppRepo{DB: db}, Keys: &ApiKeyRepo{DB: db}, Catalog: &CatalogRepo{DB: db},
		Versions: &VersionRepo{DB: db}, Policies: &PolicyRepo{DB: db}, Webhooks: &WebhookRepo{DB: db},
		Deliveries: &DeliveryRepo{DB: db}, SDKs: &SdkRepo{DB: db}, Integrations: &IntegrationRepo{DB: db},
		Usage: &UsageRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
