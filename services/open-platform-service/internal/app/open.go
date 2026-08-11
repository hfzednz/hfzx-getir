package app

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/open-platform-service/internal/domain"
)

func (d *Deps) CreateApp(ctx context.Context, a domain.DeveloperApp) (domain.DeveloperApp, error) {
	if err := domain.ValidateApp(a); err != nil {
		return domain.DeveloperApp{}, err
	}
	a.ID = d.newID()
	a.Active = true
	a.CreatedAt = d.now()
	if len(a.Scopes) == 0 {
		a.Scopes = []string{"read"}
	}
	if d.Identity != nil {
		if cid, err := d.Identity.RegisterOAuthClient(ctx, a.TenantID, a.Name, a.Scopes); err == nil {
			a.OAuthClientID = cid
		}
	}
	if err := d.Apps.Save(ctx, a); err != nil {
		return domain.DeveloperApp{}, err
	}
	return a, nil
}

func (d *Deps) CreateApiKey(ctx context.Context, tenantID, appID uuid.UUID, name string, scopes []string) (domain.ApiKey, error) {
	if _, err := d.Apps.Get(ctx, tenantID, appID); err != nil {
		return domain.ApiKey{}, err
	}
	raw := make([]byte, 24)
	_, _ = rand.Read(raw)
	secret := "nx_" + hex.EncodeToString(raw)
	prefix := secret[:10]
	k := domain.ApiKey{
		ID: d.newID(), TenantID: tenantID, AppID: appID, Name: name, Prefix: prefix,
		SecretHash: domain.HashSecret(secret), SecretOnce: secret, Scopes: scopes,
		CreatedAt: d.now(),
	}
	if len(k.Scopes) == 0 {
		k.Scopes = []string{"read"}
	}
	if err := d.Keys.Save(ctx, k); err != nil {
		return domain.ApiKey{}, err
	}
	d.emit(ctx, tenantID, k.ID, domain.EventApiKeyCreated, map[string]any{"appId": appID.String(), "prefix": prefix})
	return k, nil
}

func (d *Deps) SeedCatalog(ctx context.Context, tenantID uuid.UUID) error {
	all := append(domain.DefaultPublicCatalog(), domain.DefaultPrivateCatalog()...)
	all = append(all, domain.DefaultPartnerCatalog()...)
	for _, e := range all {
		e.ID = d.newID()
		e.TenantID = tenantID
		e.Active = true
		e.CreatedAt = d.now()
		if err := d.Catalog.Save(ctx, e); err != nil {
			return err
		}
		v := domain.ApiVersion{
			ID: d.newID(), TenantID: tenantID, CatalogKey: e.Key, Version: "v1",
			Status: "current", ReleasedAt: d.now(), Notes: "initial",
		}
		_ = d.Versions.Save(ctx, v)
		d.emit(ctx, tenantID, v.ID, domain.EventApiVersionReleased, map[string]any{"catalogKey": e.Key, "version": "v1"})
	}
	return nil
}

func (d *Deps) UpsertPolicy(ctx context.Context, p domain.GatewayPolicy) (domain.GatewayPolicy, error) {
	if p.TenantID == uuid.Nil || p.Name == "" || p.RoutePrefix == "" || p.TargetService == "" {
		return domain.GatewayPolicy{}, domain.ErrInvalidArgument
	}
	if p.ID == uuid.Nil {
		p.ID = d.newID()
	}
	if p.RateLimitRPM <= 0 {
		p.RateLimitRPM = 600
	}
	if p.Version == "" {
		p.Version = "v1"
	}
	p.Active = true
	p.UpdatedAt = d.now()
	if err := d.Policies.Save(ctx, p); err != nil {
		return domain.GatewayPolicy{}, err
	}
	return p, nil
}

func (d *Deps) ExportGatewayPolicies(ctx context.Context, tenantID uuid.UUID) ([]map[string]any, error) {
	list, err := d.Policies.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(list))
	for _, p := range list {
		if !p.Active {
			continue
		}
		out = append(out, map[string]any{
			"name": p.Name, "prefix": p.RoutePrefix, "cluster": p.TargetService,
			"version": p.Version, "rateLimitRpm": p.RateLimitRPM, "quotaDaily": p.QuotaDaily,
			"canaryPercent": p.CanaryPercent, "scopes": p.RequireScopes, "cacheTtl": p.CacheTTLSeconds,
		})
	}
	return out, nil
}

func (d *Deps) RegisterWebhook(ctx context.Context, w domain.WebhookEndpoint) (domain.WebhookEndpoint, error) {
	if w.TenantID == uuid.Nil || w.AppID == uuid.Nil || (!strings.HasPrefix(w.URL, "https://") && !strings.HasPrefix(w.URL, "http://")) {
		return domain.WebhookEndpoint{}, domain.ErrInvalidArgument
	}
	if _, err := d.Apps.Get(ctx, w.TenantID, w.AppID); err != nil {
		return domain.WebhookEndpoint{}, err
	}
	w.ID = d.newID()
	if w.Secret == "" {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		w.Secret = hex.EncodeToString(b)
	}
	if w.Version == "" {
		w.Version = "v1"
	}
	w.Active = true
	w.CreatedAt = d.now()
	if err := d.Webhooks.Save(ctx, w); err != nil {
		return domain.WebhookEndpoint{}, err
	}
	d.emit(ctx, w.TenantID, w.ID, domain.EventWebhookRegistered, map[string]any{"url": w.URL, "events": w.Events})
	return w, nil
}

func (d *Deps) EnqueueWebhook(ctx context.Context, tenantID uuid.UUID, eventType string, payload map[string]any) (int, error) {
	ends, err := d.Webhooks.ListActiveForEvent(ctx, tenantID, eventType)
	if err != nil {
		return 0, err
	}
	n := 0
	body, _ := json.Marshal(payload)
	for _, e := range ends {
		sig := domain.SignWebhook(e.Secret, string(body))
		del := domain.WebhookDelivery{
			ID: d.newID(), TenantID: tenantID, EndpointID: e.ID, EventType: eventType,
			Payload: payload, Status: domain.DeliveryPending, Signature: sig, CreatedAt: d.now(),
		}
		if err := d.Deliveries.Save(ctx, del); err != nil {
			continue
		}
		n++
	}
	return n, nil
}

func (d *Deps) ProcessDeliveries(ctx context.Context, limit int) (int, error) {
	if limit <= 0 {
		limit = 50
	}
	pending, err := d.Deliveries.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	okCount := 0
	for _, del := range pending {
		ep, err := d.Webhooks.Get(ctx, del.TenantID, del.EndpointID)
		if err != nil {
			del.Status = domain.DeliveryDLQ
			del.LastError = err.Error()
			_ = d.Deliveries.Save(ctx, del)
			continue
		}
		body, _ := json.Marshal(del.Payload)
		headers := map[string]string{
			"Content-Type":           "application/json",
			"X-Nexora-Signature":     del.Signature,
			"X-Nexora-Event":         del.EventType,
			"X-Nexora-Delivery-Id":   del.ID.String(),
			"X-Nexora-Webhook-Version": ep.Version,
		}
		del.Attempts++
		status := 0
		var postErr error
		if d.HTTP != nil {
			status, postErr = d.HTTP.Post(ctx, ep.URL, headers, body)
		} else {
			status = 200
		}
		del.LastStatus = status
		now := d.now()
		if postErr != nil || status < 200 || status >= 300 {
			del.LastError = fmt.Sprintf("status=%d err=%v", status, postErr)
			if del.Attempts >= 5 {
				del.Status = domain.DeliveryDLQ
			} else {
				del.Status = domain.DeliveryPending
			}
			_ = d.Deliveries.Save(ctx, del)
			continue
		}
		del.Status = domain.DeliverySuccess
		del.DeliveredAt = &now
		_ = d.Deliveries.Save(ctx, del)
		d.emit(ctx, del.TenantID, del.ID, domain.EventWebhookDelivered, map[string]any{
			"endpointId": ep.ID.String(), "eventType": del.EventType,
		})
		okCount++
	}
	return okCount, nil
}

func (d *Deps) ReplayDelivery(ctx context.Context, tenantID, id uuid.UUID) (domain.WebhookDelivery, error) {
	del, err := d.Deliveries.Get(ctx, tenantID, id)
	if err != nil {
		return domain.WebhookDelivery{}, err
	}
	del.Status = domain.DeliveryPending
	del.Attempts = 0
	del.LastError = ""
	if err := d.Deliveries.Save(ctx, del); err != nil {
		return domain.WebhookDelivery{}, err
	}
	_, _ = d.ProcessDeliveries(ctx, 1)
	return d.Deliveries.Get(ctx, tenantID, id)
}

func (d *Deps) PublishSdk(ctx context.Context, s domain.SdkPackage) (domain.SdkPackage, error) {
	if s.TenantID == uuid.Nil || s.Language == "" || s.Name == "" {
		return domain.SdkPackage{}, domain.ErrInvalidArgument
	}
	s.ID = d.newID()
	if s.Version == "" {
		s.Version = "0.1.0"
	}
	if s.RepoPath == "" {
		s.RepoPath = fmt.Sprintf("packages/sdk/%s", s.Language)
	}
	s.Status = "published"
	s.CreatedAt = d.now()
	if err := d.SDKs.Save(ctx, s); err != nil {
		return domain.SdkPackage{}, err
	}
	d.emit(ctx, s.TenantID, s.ID, domain.EventSdkGenerated, map[string]any{
		"language": string(s.Language), "version": s.Version, "path": s.RepoPath,
	})
	return s, nil
}

func (d *Deps) ConnectIntegration(ctx context.Context, i domain.IntegrationConnector) (domain.IntegrationConnector, error) {
	if i.TenantID == uuid.Nil || i.AppID == uuid.Nil || i.Kind == "" || i.Provider == "" {
		return domain.IntegrationConnector{}, domain.ErrInvalidArgument
	}
	if _, err := d.Apps.Get(ctx, i.TenantID, i.AppID); err != nil {
		return domain.IntegrationConnector{}, err
	}
	i.ID = d.newID()
	i.Status = "connected"
	i.CreatedAt = d.now()
	if i.ConfigRef == "" {
		i.ConfigRef = "vault://integrations/" + string(i.Kind) + "/" + i.Provider
	}
	if err := d.Integrations.Save(ctx, i); err != nil {
		return domain.IntegrationConnector{}, err
	}
	d.emit(ctx, i.TenantID, i.ID, domain.EventPartnerIntegrated, map[string]any{
		"kind": string(i.Kind), "provider": i.Provider, "appId": i.AppID.String(),
	})
	return i, nil
}

func (d *Deps) RecordUsage(ctx context.Context, tenantID, appID uuid.UUID, ok bool) error {
	day := d.now().Format("2006-01-02")
	u, err := d.Usage.Get(ctx, tenantID, appID, day)
	if err != nil {
		u = domain.UsageCounter{ID: d.newID(), TenantID: tenantID, AppID: appID, Day: day}
	}
	u.Requests++
	if !ok {
		u.Errors++
	}
	u.UpdatedAt = d.now()
	return d.Usage.Save(ctx, u)
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	apps, _ := d.Apps.List(ctx, tenantID)
	cat, _ := d.Catalog.List(ctx, tenantID)
	sdks, _ := d.SDKs.List(ctx, tenantID)
	ints, _ := d.Integrations.List(ctx, tenantID)
	pols, _ := d.Policies.List(ctx, tenantID)
	return map[string]any{
		"apps": len(apps), "catalogEntries": len(cat), "sdks": len(sdks),
		"integrations": len(ints), "gatewayPolicies": len(pols),
	}, nil
}

func (d *Deps) ListCatalog(ctx context.Context, tenantID uuid.UUID, surface string) ([]domain.CatalogEntry, error) {
	all, err := d.Catalog.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if surface == "" {
		return all, nil
	}
	out := []domain.CatalogEntry{}
	for _, e := range all {
		if string(e.Surface) == surface {
			out = append(out, e)
		}
	}
	return out, nil
}
