package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/innovation-service/internal/domain"
)

type Store struct {
	mu           sync.RWMutex
	Modules      map[uuid.UUID]domain.InnovationModule
	ModuleByKey  map[string]uuid.UUID
	Experiments  map[uuid.UUID]domain.ResearchExperiment
	Simulations  map[uuid.UUID]domain.SimulationRun
	Twins        map[uuid.UUID]domain.DigitalTwin
	Edge         map[uuid.UUID]domain.EdgeNode
	IoT          map[uuid.UUID]domain.IoTDevice
	Robots       map[uuid.UUID]domain.Robot
	Assignments  map[uuid.UUID]domain.RobotAssignment
	Drones       map[uuid.UUID]domain.DroneMission
	Blockchain   map[uuid.UUID]domain.BlockchainHook
	XR           map[uuid.UUID]domain.XRExperience
	Multimodal   map[uuid.UUID]domain.MultimodalSession
	Green        map[string]domain.GreenMetric
	Quantum      map[uuid.UUID]domain.QuantumHook
	Outbox       map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Modules: map[uuid.UUID]domain.InnovationModule{}, ModuleByKey: map[string]uuid.UUID{},
		Experiments: map[uuid.UUID]domain.ResearchExperiment{}, Simulations: map[uuid.UUID]domain.SimulationRun{},
		Twins: map[uuid.UUID]domain.DigitalTwin{}, Edge: map[uuid.UUID]domain.EdgeNode{},
		IoT: map[uuid.UUID]domain.IoTDevice{}, Robots: map[uuid.UUID]domain.Robot{},
		Assignments: map[uuid.UUID]domain.RobotAssignment{}, Drones: map[uuid.UUID]domain.DroneMission{},
		Blockchain: map[uuid.UUID]domain.BlockchainHook{}, XR: map[uuid.UUID]domain.XRExperience{},
		Multimodal: map[uuid.UUID]domain.MultimodalSession{}, Green: map[string]domain.GreenMetric{},
		Quantum: map[uuid.UUID]domain.QuantumHook{}, Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

func mk(tenantID uuid.UUID, key string) string { return tenantID.String() + ":" + key }
func greenKey(tenantID uuid.UUID, period string) string {
	return tenantID.String() + ":" + period
}

type Repos struct {
	Modules     *ModuleRepo
	Experiments *ExperimentRepo
	Simulations *SimulationRepo
	Twins       *TwinRepo
	Edge        *EdgeRepo
	IoT         *IoTRepo
	Robots      *RobotRepo
	Assignments *AssignmentRepo
	Drones      *DroneRepo
	Blockchain  *BlockchainRepo
	XR          *XRRepo
	Multimodal  *MultimodalRepo
	Green       *GreenRepo
	Quantum     *QuantumRepo
	Outbox      *OutboxRepo
	LiveOps     *MockLiveOps
	AI          *MockAI
	Metrics     *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Modules: &ModuleRepo{s: s}, Experiments: &ExperimentRepo{s: s}, Simulations: &SimulationRepo{s: s},
		Twins: &TwinRepo{s: s}, Edge: &EdgeRepo{s: s}, IoT: &IoTRepo{s: s}, Robots: &RobotRepo{s: s},
		Assignments: &AssignmentRepo{s: s}, Drones: &DroneRepo{s: s}, Blockchain: &BlockchainRepo{s: s},
		XR: &XRRepo{s: s}, Multimodal: &MultimodalRepo{s: s}, Green: &GreenRepo{s: s}, Quantum: &QuantumRepo{s: s},
		Outbox: &OutboxRepo{s: s}, LiveOps: &MockLiveOps{allowed: true}, AI: &MockAI{}, Metrics: &MockMetrics{},
	}
}

type ModuleRepo struct{ s *Store }

func (r *ModuleRepo) Save(_ context.Context, m domain.InnovationModule) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Modules[m.ID] = m
	r.s.ModuleByKey[mk(m.TenantID, m.Key)] = m.ID
	return nil
}
func (r *ModuleRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.InnovationModule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	m, ok := r.s.Modules[id]
	if !ok || m.TenantID != tenantID {
		return domain.InnovationModule{}, domain.ErrNotFound
	}
	return m, nil
}
func (r *ModuleRepo) GetByKey(_ context.Context, tenantID uuid.UUID, key string) (domain.InnovationModule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	id, ok := r.s.ModuleByKey[mk(tenantID, key)]
	if !ok {
		return domain.InnovationModule{}, domain.ErrNotFound
	}
	return r.s.Modules[id], nil
}
func (r *ModuleRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.InnovationModule, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.InnovationModule{}
	for _, m := range r.s.Modules {
		if m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

type ExperimentRepo struct{ s *Store }

func (r *ExperimentRepo) Save(_ context.Context, e domain.ResearchExperiment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Experiments[e.ID] = e
	return nil
}
func (r *ExperimentRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.ResearchExperiment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ResearchExperiment{}
	for _, e := range r.s.Experiments {
		if e.TenantID == tenantID {
			out = append(out, e)
		}
	}
	return out, nil
}

type SimulationRepo struct{ s *Store }

func (r *SimulationRepo) Save(_ context.Context, s domain.SimulationRun) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Simulations[s.ID] = s
	return nil
}
func (r *SimulationRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.SimulationRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Simulations[id]
	if !ok || s.TenantID != tenantID {
		return domain.SimulationRun{}, domain.ErrNotFound
	}
	return s, nil
}
func (r *SimulationRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.SimulationRun, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SimulationRun{}
	for _, s := range r.s.Simulations {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type TwinRepo struct{ s *Store }

func (r *TwinRepo) Save(_ context.Context, t domain.DigitalTwin) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Twins[t.ID] = t
	return nil
}
func (r *TwinRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.DigitalTwin, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.DigitalTwin{}
	for _, t := range r.s.Twins {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	return out, nil
}

type EdgeRepo struct{ s *Store }

func (r *EdgeRepo) Save(_ context.Context, n domain.EdgeNode) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Edge[n.ID] = n
	return nil
}
func (r *EdgeRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.EdgeNode, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.EdgeNode{}
	for _, n := range r.s.Edge {
		if n.TenantID == tenantID {
			out = append(out, n)
		}
	}
	return out, nil
}

type IoTRepo struct{ s *Store }

func (r *IoTRepo) Save(_ context.Context, d domain.IoTDevice) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.IoT[d.ID] = d
	return nil
}
func (r *IoTRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.IoTDevice, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.IoTDevice{}
	for _, d := range r.s.IoT {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out, nil
}

type RobotRepo struct{ s *Store }

func (r *RobotRepo) Save(_ context.Context, robot domain.Robot) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Robots[robot.ID] = robot
	return nil
}
func (r *RobotRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Robot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	robot, ok := r.s.Robots[id]
	if !ok || robot.TenantID != tenantID {
		return domain.Robot{}, domain.ErrNotFound
	}
	return robot, nil
}
func (r *RobotRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Robot, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Robot{}
	for _, robot := range r.s.Robots {
		if robot.TenantID == tenantID {
			out = append(out, robot)
		}
	}
	return out, nil
}

type AssignmentRepo struct{ s *Store }

func (r *AssignmentRepo) Save(_ context.Context, a domain.RobotAssignment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Assignments[a.ID] = a
	return nil
}
func (r *AssignmentRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.RobotAssignment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.RobotAssignment{}
	for _, a := range r.s.Assignments {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

type DroneRepo struct{ s *Store }

func (r *DroneRepo) Save(_ context.Context, m domain.DroneMission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Drones[m.ID] = m
	return nil
}
func (r *DroneRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.DroneMission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.DroneMission{}
	for _, m := range r.s.Drones {
		if m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

type BlockchainRepo struct{ s *Store }

func (r *BlockchainRepo) Save(_ context.Context, h domain.BlockchainHook) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Blockchain[h.ID] = h
	return nil
}
func (r *BlockchainRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.BlockchainHook, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.BlockchainHook{}
	for _, h := range r.s.Blockchain {
		if h.TenantID == tenantID {
			out = append(out, h)
		}
	}
	return out, nil
}

type XRRepo struct{ s *Store }

func (r *XRRepo) Save(_ context.Context, x domain.XRExperience) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.XR[x.ID] = x
	return nil
}
func (r *XRRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.XRExperience, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.XRExperience{}
	for _, x := range r.s.XR {
		if x.TenantID == tenantID {
			out = append(out, x)
		}
	}
	return out, nil
}

type MultimodalRepo struct{ s *Store }

func (r *MultimodalRepo) Save(_ context.Context, s domain.MultimodalSession) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Multimodal[s.ID] = s
	return nil
}
func (r *MultimodalRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.MultimodalSession, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.MultimodalSession{}
	for _, s := range r.s.Multimodal {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type GreenRepo struct{ s *Store }

func (r *GreenRepo) Save(_ context.Context, g domain.GreenMetric) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Green[greenKey(g.TenantID, g.Period)] = g
	return nil
}
func (r *GreenRepo) GetByPeriod(_ context.Context, tenantID uuid.UUID, period string) (domain.GreenMetric, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	g, ok := r.s.Green[greenKey(tenantID, period)]
	if !ok {
		return domain.GreenMetric{}, domain.ErrNotFound
	}
	return g, nil
}
func (r *GreenRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.GreenMetric, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.GreenMetric{}
	for _, g := range r.s.Green {
		if g.TenantID == tenantID {
			out = append(out, g)
		}
	}
	return out, nil
}

type QuantumRepo struct{ s *Store }

func (r *QuantumRepo) Save(_ context.Context, q domain.QuantumHook) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Quantum[q.ID] = q
	return nil
}
func (r *QuantumRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.QuantumHook, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.QuantumHook{}
	for _, q := range r.s.Quantum {
		if q.TenantID == tenantID {
			out = append(out, q)
		}
	}
	return out, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}
func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}

type MockLiveOps struct{ allowed bool }

func (m *MockLiveOps) InnovationAllowed(context.Context, uuid.UUID, string) (bool, error) {
	return m.allowed, nil
}

type MockAI struct{}

func (MockAI) ScoreInnovation(context.Context, uuid.UUID, string) (float64, error) { return 72.5, nil }

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }
