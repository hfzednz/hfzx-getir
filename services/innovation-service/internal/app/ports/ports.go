package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/innovation-service/internal/domain"
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

type ModuleRepo interface {
	Save(ctx context.Context, m domain.InnovationModule) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.InnovationModule, error)
	GetByKey(ctx context.Context, tenantID uuid.UUID, key string) (domain.InnovationModule, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.InnovationModule, error)
}

type ExperimentRepo interface {
	Save(ctx context.Context, e domain.ResearchExperiment) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.ResearchExperiment, error)
}

type SimulationRepo interface {
	Save(ctx context.Context, s domain.SimulationRun) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.SimulationRun, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.SimulationRun, error)
}

type TwinRepo interface {
	Save(ctx context.Context, t domain.DigitalTwin) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.DigitalTwin, error)
}

type EdgeRepo interface {
	Save(ctx context.Context, n domain.EdgeNode) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.EdgeNode, error)
}

type IoTRepo interface {
	Save(ctx context.Context, d domain.IoTDevice) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.IoTDevice, error)
}

type RobotRepo interface {
	Save(ctx context.Context, r domain.Robot) error
	Get(ctx context.Context, tenantID, id uuid.UUID) (domain.Robot, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.Robot, error)
}

type AssignmentRepo interface {
	Save(ctx context.Context, a domain.RobotAssignment) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.RobotAssignment, error)
}

type DroneRepo interface {
	Save(ctx context.Context, m domain.DroneMission) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.DroneMission, error)
}

type BlockchainRepo interface {
	Save(ctx context.Context, h domain.BlockchainHook) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.BlockchainHook, error)
}

type XRRepo interface {
	Save(ctx context.Context, x domain.XRExperience) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.XRExperience, error)
}

type MultimodalRepo interface {
	Save(ctx context.Context, s domain.MultimodalSession) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.MultimodalSession, error)
}

type GreenRepo interface {
	Save(ctx context.Context, g domain.GreenMetric) error
	GetByPeriod(ctx context.Context, tenantID uuid.UUID, period string) (domain.GreenMetric, error)
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.GreenMetric, error)
}

type QuantumRepo interface {
	Save(ctx context.Context, q domain.QuantumHook) error
	List(ctx context.Context, tenantID uuid.UUID) ([]domain.QuantumHook, error)
}

// LiveOpsClient gates experimental enablement without owning flag SoT.
type LiveOpsClient interface {
	InnovationAllowed(ctx context.Context, tenantID uuid.UUID, moduleKey string) (bool, error)
}

// AIClient assists research scoring / synthetic insights.
type AIClient interface {
	ScoreInnovation(ctx context.Context, tenantID uuid.UUID, key string) (float64, error)
}

type MetricsClient interface {
	Record(ctx context.Context, name string, tags map[string]string, value float64) error
}
