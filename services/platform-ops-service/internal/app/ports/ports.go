package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/domain"
)

type Clock interface{ Now() time.Time }
type IDGen interface{ New() uuid.UUID }

type DeploymentRepo interface {
	Save(ctx context.Context, d domain.Deployment) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Deployment, error)
	List(ctx context.Context, tenantID uuid.UUID, env string) ([]domain.Deployment, error)
}

type ScalingRepo interface {
	Save(ctx context.Context, s domain.ScalingEvent) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ScalingEvent, error)
}

type BackupRepo interface {
	Save(ctx context.Context, b domain.BackupJob) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.BackupJob, error)
}

type RecoveryRepo interface {
	Save(ctx context.Context, r domain.RecoveryJob) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.RecoveryJob, error)
}

type AlertRepo interface {
	Save(ctx context.Context, a domain.AlertEvent) error
	List(ctx context.Context, tenantID uuid.UUID, status string) ([]domain.AlertEvent, error)
}

type CostRepo interface {
	Save(ctx context.Context, c domain.CostSnapshot) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.CostSnapshot, error)
}

type SLORepo interface {
	Save(ctx context.Context, s domain.SLOReport) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.SLOReport, error)
}

type OutboxRepository interface {
	Enqueue(ctx context.Context, m domain.OutboxMessage) error
	ListPending(ctx context.Context, limit int) ([]domain.OutboxMessage, error)
	Update(ctx context.Context, m domain.OutboxMessage) error
}

type EventPublisher interface {
	Publish(ctx context.Context, topic, key string, payload map[string]any) error
}

// GitOpsClient syncs/rollbacks overlays (Argo).
type GitOpsClient interface {
	Sync(ctx context.Context, environment, service, imageTag string) error
	Rollback(ctx context.Context, environment, service string) error
}

// ClusterClient scales workloads.
type ClusterClient interface {
	Scale(ctx context.Context, environment, service string, replicas int) error
}

// BackupClient triggers backup tools (Velero/pg).
type BackupClient interface {
	RunBackup(ctx context.Context, kind, target string) (location string, err error)
}

// Registry persists the super-admin tenant/company directory.
type Registry interface {
	ListTenants(ctx context.Context) ([]domain.PlatformTenant, []domain.DualControlProposal, error)
	GetTenant(ctx context.Context, id string) (domain.PlatformTenant, error)
	SaveTenant(ctx context.Context, t domain.PlatformTenant) error
	ListCompanies(ctx context.Context) ([]domain.PlatformCompany, error)
	GetCompany(ctx context.Context, id string) (domain.PlatformCompany, error)
	SaveCompany(ctx context.Context, c domain.PlatformCompany) error
	DeleteCompany(ctx context.Context, id string) error
	SaveProposal(ctx context.Context, p domain.DualControlProposal) error
	GetProposal(ctx context.Context, id string) (domain.DualControlProposal, error)
	AppendAudit(ctx context.Context, e domain.PlatformAuditEntry) error
	ListAudit(ctx context.Context, q string) ([]domain.PlatformAuditEntry, error)
	ListPeople(ctx context.Context) ([]domain.PlatformPerson, error)
	SavePerson(ctx context.Context, p domain.PlatformPerson) error
}
