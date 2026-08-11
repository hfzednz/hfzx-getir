package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/platform-ops-service/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	Deployments []domain.Deployment
	Scaling     []domain.ScalingEvent
	Backups     []domain.BackupJob
	Recoveries  []domain.RecoveryJob
	Alerts      []domain.AlertEvent
	Costs       []domain.CostSnapshot
	SLOs        []domain.SLOReport
	Outbox      []domain.OutboxMessage
}

func NewStore() *Store { return &Store{} }

type Repos struct {
	Deployments *DeploymentRepo
	Scaling     *ScalingRepo
	Backups     *BackupRepo
	Recoveries  *RecoveryRepo
	Alerts      *AlertRepo
	Costs       *CostRepo
	SLOs        *SLORepo
	Outbox      *OutboxRepo
	GitOps      *MockGitOps
	Cluster     *MockCluster
	BackupTool  *MockBackup
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Deployments: &DeploymentRepo{s: s}, Scaling: &ScalingRepo{s: s}, Backups: &BackupRepo{s: s},
		Recoveries: &RecoveryRepo{s: s}, Alerts: &AlertRepo{s: s}, Costs: &CostRepo{s: s},
		SLOs: &SLORepo{s: s}, Outbox: &OutboxRepo{s: s},
		GitOps: &MockGitOps{}, Cluster: &MockCluster{}, BackupTool: &MockBackup{},
	}
}

type DeploymentRepo struct{ s *Store }

func (r *DeploymentRepo) Save(_ context.Context, d domain.Deployment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Deployments {
		if r.s.Deployments[i].ID == d.ID {
			r.s.Deployments[i] = d
			return nil
		}
	}
	r.s.Deployments = append(r.s.Deployments, d)
	return nil
}

func (r *DeploymentRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Deployment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, d := range r.s.Deployments {
		if d.TenantID == tenantID && d.ID == id {
			return d, nil
		}
	}
	return domain.Deployment{}, domain.ErrNotFound
}

func (r *DeploymentRepo) List(_ context.Context, tenantID uuid.UUID, env string) ([]domain.Deployment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Deployment{}
	for _, d := range r.s.Deployments {
		if d.TenantID == tenantID && (env == "" || d.Environment == env) {
			out = append(out, d)
		}
	}
	return out, nil
}

type ScalingRepo struct{ s *Store }

func (r *ScalingRepo) Save(_ context.Context, s domain.ScalingEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Scaling = append(r.s.Scaling, s)
	return nil
}

func (r *ScalingRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ScalingEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ScalingEvent{}
	for _, s := range r.s.Scaling {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type BackupRepo struct{ s *Store }

func (r *BackupRepo) Save(_ context.Context, b domain.BackupJob) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Backups {
		if r.s.Backups[i].ID == b.ID {
			r.s.Backups[i] = b
			return nil
		}
	}
	r.s.Backups = append(r.s.Backups, b)
	return nil
}

func (r *BackupRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.BackupJob, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.BackupJob{}
	for _, b := range r.s.Backups {
		if b.TenantID == tenantID {
			out = append(out, b)
		}
	}
	return out, nil
}

type RecoveryRepo struct{ s *Store }

func (r *RecoveryRepo) Save(_ context.Context, rec domain.RecoveryJob) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Recoveries {
		if r.s.Recoveries[i].ID == rec.ID {
			r.s.Recoveries[i] = rec
			return nil
		}
	}
	r.s.Recoveries = append(r.s.Recoveries, rec)
	return nil
}

func (r *RecoveryRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.RecoveryJob, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.RecoveryJob{}
	for _, rec := range r.s.Recoveries {
		if rec.TenantID == tenantID {
			out = append(out, rec)
		}
	}
	return out, nil
}

type AlertRepo struct{ s *Store }

func (r *AlertRepo) Save(_ context.Context, a domain.AlertEvent) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Alerts {
		if r.s.Alerts[i].ID == a.ID {
			r.s.Alerts[i] = a
			return nil
		}
	}
	r.s.Alerts = append(r.s.Alerts, a)
	return nil
}

func (r *AlertRepo) List(_ context.Context, tenantID uuid.UUID, status string) ([]domain.AlertEvent, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.AlertEvent{}
	for _, a := range r.s.Alerts {
		if a.TenantID == tenantID && (status == "" || a.Status == status) {
			out = append(out, a)
		}
	}
	return out, nil
}

type CostRepo struct{ s *Store }

func (r *CostRepo) Save(_ context.Context, c domain.CostSnapshot) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Costs = append(r.s.Costs, c)
	return nil
}

func (r *CostRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.CostSnapshot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.CostSnapshot{}
	for _, c := range r.s.Costs {
		if c.TenantID == tenantID {
			out = append(out, c)
		}
	}
	return out, nil
}

type SLORepo struct{ s *Store }

func (r *SLORepo) Save(_ context.Context, s domain.SLOReport) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.SLOs = append(r.s.SLOs, s)
	return nil
}

func (r *SLORepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.SLOReport, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SLOReport{}
	for _, s := range r.s.SLOs {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox = append(r.s.Outbox, m)
	return nil
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	for i := range r.s.Outbox {
		if r.s.Outbox[i].ID == m.ID {
			r.s.Outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

type MockGitOps struct{ Syncs, Rollbacks int }

func (m *MockGitOps) Sync(_ context.Context, _, _, _ string) error { m.Syncs++; return nil }
func (m *MockGitOps) Rollback(_ context.Context, _, _ string) error {
	m.Rollbacks++
	return nil
}

type MockCluster struct{ Last int }

func (m *MockCluster) Scale(_ context.Context, _, _ string, replicas int) error {
	m.Last = replicas
	return nil
}

type MockBackup struct{}

func (MockBackup) RunBackup(_ context.Context, kind, target string) (string, error) {
	return "s3://nexora-backups/" + kind + "/" + target + "/" + uuid.NewString()[:8], nil
}
