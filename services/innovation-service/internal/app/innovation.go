package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/innovation-service/internal/domain"
)

func (d *Deps) SeedCatalog(ctx context.Context, tenantID uuid.UUID) error {
	for _, m := range domain.DefaultInnovationCatalog() {
		m.TenantID = tenantID
		if _, err := d.UpsertModule(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deps) UpsertModule(ctx context.Context, m domain.InnovationModule) (domain.InnovationModule, error) {
	if err := domain.ValidateModule(m); err != nil {
		return domain.InnovationModule{}, err
	}
	now := d.now()
	if m.ID == uuid.Nil {
		if existing, err := d.Modules.GetByKey(ctx, m.TenantID, m.Key); err == nil {
			m.ID = existing.ID
			m.CreatedAt = existing.CreatedAt
			if m.Status == "" {
				m.Status = existing.Status
			}
		} else {
			m.ID = d.newID()
			m.CreatedAt = now
		}
	}
	if m.Status == "" {
		m.Status = domain.ModuleIncubating
	}
	if d.AI != nil {
		if score, err := d.AI.ScoreInnovation(ctx, m.TenantID, m.Key); err == nil && score > 0 {
			m.Score = score
		}
	}
	m.UpdatedAt = now
	if err := d.Modules.Save(ctx, m); err != nil {
		return domain.InnovationModule{}, err
	}
	return m, nil
}

func (d *Deps) EnableModule(ctx context.Context, tenantID uuid.UUID, key string) (domain.InnovationModule, error) {
	m, err := d.Modules.GetByKey(ctx, tenantID, key)
	if err != nil {
		return domain.InnovationModule{}, err
	}
	if err := domain.CanEnable(m); err != nil {
		return domain.InnovationModule{}, err
	}
	if d.LiveOps != nil {
		if ok, err := d.LiveOps.InnovationAllowed(ctx, tenantID, key); err == nil && !ok {
			return domain.InnovationModule{}, domain.ErrForbidden
		}
	}
	m.Status = domain.ModuleEnabled
	m.UpdatedAt = d.now()
	if err := d.Modules.Save(ctx, m); err != nil {
		return domain.InnovationModule{}, err
	}
	d.emit(ctx, tenantID, m.ID, domain.EventInnovationEnabled, map[string]any{
		"key": m.Key, "domain": string(m.Domain), "trl": int(m.TRL),
	})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "innovation.enabled", map[string]string{"key": m.Key}, 1)
	}
	return m, nil
}

func (d *Deps) CreateExperiment(ctx context.Context, e domain.ResearchExperiment) (domain.ResearchExperiment, error) {
	if e.TenantID == uuid.Nil || e.ModuleID == uuid.Nil || e.Name == "" {
		return domain.ResearchExperiment{}, domain.ErrInvalidArgument
	}
	if _, err := d.Modules.Get(ctx, e.TenantID, e.ModuleID); err != nil {
		return domain.ResearchExperiment{}, err
	}
	e.ID = d.newID()
	e.Status = "running"
	e.CreatedAt = d.now()
	if err := d.Experiments.Save(ctx, e); err != nil {
		return domain.ResearchExperiment{}, err
	}
	d.emit(ctx, e.TenantID, e.ID, domain.EventResearchExperimentCreated, map[string]any{
		"moduleId": e.ModuleID.String(), "name": e.Name,
	})
	return e, nil
}

func (d *Deps) StartSimulation(ctx context.Context, s domain.SimulationRun) (domain.SimulationRun, error) {
	if s.TenantID == uuid.Nil || s.Kind == "" || s.Name == "" {
		return domain.SimulationRun{}, domain.ErrInvalidArgument
	}
	now := d.now()
	s.ID = d.newID()
	s.Status = domain.SimRunning
	s.StartedAt = &now
	s.CreatedAt = now
	if s.Params == nil {
		s.Params = map[string]any{}
	}
	if err := d.Simulations.Save(ctx, s); err != nil {
		return domain.SimulationRun{}, err
	}
	d.emit(ctx, s.TenantID, s.ID, domain.EventSimulationStarted, map[string]any{
		"kind": string(s.Kind), "name": s.Name,
	})
	return s, nil
}

func (d *Deps) CompleteSimulation(ctx context.Context, tenantID, id uuid.UUID) (domain.SimulationRun, error) {
	s, err := d.Simulations.Get(ctx, tenantID, id)
	if err != nil {
		return domain.SimulationRun{}, err
	}
	if s.Status != domain.SimRunning {
		return domain.SimulationRun{}, domain.ErrIllegalTransition
	}
	now := d.now()
	s.Status = domain.SimCompleted
	s.CompletedAt = &now
	s.Accuracy = domain.EstimateSimAccuracy(s.Kind, s.Params)
	s.ResultSummary = fmt.Sprintf("%s simulation completed accuracy=%.2f", s.Kind, s.Accuracy)
	if err := d.Simulations.Save(ctx, s); err != nil {
		return domain.SimulationRun{}, err
	}
	d.emit(ctx, tenantID, s.ID, domain.EventSimulationCompleted, map[string]any{
		"kind": string(s.Kind), "accuracy": s.Accuracy,
	})
	return s, nil
}

func (d *Deps) UpsertTwin(ctx context.Context, t domain.DigitalTwin) (domain.DigitalTwin, error) {
	if t.TenantID == uuid.Nil || t.Kind == "" || t.RefKey == "" || t.Name == "" {
		return domain.DigitalTwin{}, domain.ErrInvalidArgument
	}
	now := d.now()
	if t.ID == uuid.Nil {
		t.ID = d.newID()
		t.CreatedAt = now
	}
	if t.Version == "" {
		t.Version = "1.0.0"
	}
	t.Active = true
	t.UpdatedAt = now
	if err := d.Twins.Save(ctx, t); err != nil {
		return domain.DigitalTwin{}, err
	}
	return t, nil
}

func (d *Deps) RegisterEdge(ctx context.Context, n domain.EdgeNode) (domain.EdgeNode, error) {
	if n.TenantID == uuid.Nil || n.Key == "" {
		return domain.EdgeNode{}, domain.ErrInvalidArgument
	}
	n.ID = d.newID()
	n.Status = "online"
	n.CreatedAt = d.now()
	if len(n.Caps) == 0 {
		n.Caps = []string{"cache", "sync"}
	}
	if err := d.Edge.Save(ctx, n); err != nil {
		return domain.EdgeNode{}, err
	}
	d.emit(ctx, n.TenantID, n.ID, domain.EventEdgeNodeRegistered, map[string]any{
		"key": n.Key, "region": n.Region,
	})
	return n, nil
}

func (d *Deps) ConnectIoT(ctx context.Context, device domain.IoTDevice) (domain.IoTDevice, error) {
	if device.TenantID == uuid.Nil || device.DeviceKey == "" || device.Kind == "" {
		return domain.IoTDevice{}, domain.ErrInvalidArgument
	}
	now := d.now()
	device.ID = d.newID()
	device.Connected = true
	device.LastSeenAt = &now
	device.CreatedAt = now
	if err := d.IoT.Save(ctx, device); err != nil {
		return domain.IoTDevice{}, err
	}
	d.emit(ctx, device.TenantID, device.ID, domain.EventIoTDeviceConnected, map[string]any{
		"deviceKey": device.DeviceKey, "kind": string(device.Kind),
	})
	return device, nil
}

func (d *Deps) RegisterRobot(ctx context.Context, r domain.Robot) (domain.Robot, error) {
	if r.TenantID == uuid.Nil || r.Key == "" || r.Kind == "" {
		return domain.Robot{}, domain.ErrInvalidArgument
	}
	r.ID = d.newID()
	r.Status = "idle"
	r.CreatedAt = d.now()
	if err := d.Robots.Save(ctx, r); err != nil {
		return domain.Robot{}, err
	}
	return r, nil
}

func (d *Deps) AssignRobot(ctx context.Context, tenantID, robotID uuid.UUID, taskRef string) (domain.RobotAssignment, error) {
	if taskRef == "" {
		return domain.RobotAssignment{}, domain.ErrInvalidArgument
	}
	robot, err := d.Robots.Get(ctx, tenantID, robotID)
	if err != nil {
		return domain.RobotAssignment{}, err
	}
	if robot.Status == "offline" {
		return domain.RobotAssignment{}, domain.ErrForbidden
	}
	a := domain.RobotAssignment{
		ID: d.newID(), TenantID: tenantID, RobotID: robotID, TaskRef: taskRef,
		Status: "assigned", AssignedAt: d.now(),
	}
	robot.Status = "assigned"
	_ = d.Robots.Save(ctx, robot)
	if err := d.Assignments.Save(ctx, a); err != nil {
		return domain.RobotAssignment{}, err
	}
	d.emit(ctx, tenantID, a.ID, domain.EventRobotAssigned, map[string]any{
		"robotId": robotID.String(), "taskRef": taskRef,
	})
	return a, nil
}

func (d *Deps) CreateDroneMission(ctx context.Context, m domain.DroneMission) (domain.DroneMission, error) {
	if m.TenantID == uuid.Nil || m.DroneKey == "" || m.LandingZone == "" {
		return domain.DroneMission{}, domain.ErrInvalidArgument
	}
	m.ID = d.newID()
	m.Status = "planned"
	m.CreatedAt = d.now()
	if len(m.Compliance) == 0 {
		m.Compliance = []string{"airspace_check", "no_fly_zone"}
	}
	if err := d.Drones.Save(ctx, m); err != nil {
		return domain.DroneMission{}, err
	}
	d.emit(ctx, m.TenantID, m.ID, domain.EventDroneMissionCreated, map[string]any{
		"droneKey": m.DroneKey, "landingZone": m.LandingZone,
	})
	return m, nil
}

func (d *Deps) RegisterBlockchainHook(ctx context.Context, h domain.BlockchainHook) (domain.BlockchainHook, error) {
	if h.TenantID == uuid.Nil || h.Purpose == "" {
		return domain.BlockchainHook{}, domain.ErrInvalidArgument
	}
	h.ID = d.newID()
	h.Active = true
	h.CreatedAt = d.now()
	if err := d.Blockchain.Save(ctx, h); err != nil {
		return domain.BlockchainHook{}, err
	}
	return h, nil
}

func (d *Deps) RegisterXR(ctx context.Context, x domain.XRExperience) (domain.XRExperience, error) {
	if x.TenantID == uuid.Nil || x.Kind == "" || x.AssetURI == "" {
		return domain.XRExperience{}, domain.ErrInvalidArgument
	}
	x.ID = d.newID()
	x.Active = true
	x.CreatedAt = d.now()
	if err := d.XR.Save(ctx, x); err != nil {
		return domain.XRExperience{}, err
	}
	return x, nil
}

func (d *Deps) StartMultimodal(ctx context.Context, s domain.MultimodalSession) (domain.MultimodalSession, error) {
	if s.TenantID == uuid.Nil || s.SubjectID == "" || len(s.Modes) == 0 {
		return domain.MultimodalSession{}, domain.ErrInvalidArgument
	}
	s.ID = d.newID()
	s.Status = "active"
	s.CreatedAt = d.now()
	if err := d.Multimodal.Save(ctx, s); err != nil {
		return domain.MultimodalSession{}, err
	}
	return s, nil
}

func (d *Deps) UpsertGreen(ctx context.Context, g domain.GreenMetric) (domain.GreenMetric, error) {
	if g.TenantID == uuid.Nil || g.Period == "" {
		return domain.GreenMetric{}, domain.ErrInvalidArgument
	}
	if existing, err := d.Green.GetByPeriod(ctx, g.TenantID, g.Period); err == nil {
		g.ID = existing.ID
	} else {
		g.ID = d.newID()
	}
	g.UpdatedAt = d.now()
	if err := d.Green.Save(ctx, g); err != nil {
		return domain.GreenMetric{}, err
	}
	return g, nil
}

func (d *Deps) RegisterQuantum(ctx context.Context, q domain.QuantumHook) (domain.QuantumHook, error) {
	if q.TenantID == uuid.Nil || q.Kind == "" || q.Adapter == "" {
		return domain.QuantumHook{}, domain.ErrInvalidArgument
	}
	q.ID = d.newID()
	q.Active = true
	q.CreatedAt = d.now()
	if err := d.Quantum.Save(ctx, q); err != nil {
		return domain.QuantumHook{}, err
	}
	return q, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	mods, _ := d.Modules.List(ctx, tenantID)
	sims, _ := d.Simulations.List(ctx, tenantID)
	twins, _ := d.Twins.List(ctx, tenantID)
	edge, _ := d.Edge.List(ctx, tenantID)
	iot, _ := d.IoT.List(ctx, tenantID)
	robots, _ := d.Robots.List(ctx, tenantID)
	drones, _ := d.Drones.List(ctx, tenantID)
	exps, _ := d.Experiments.List(ctx, tenantID)
	enabled := 0
	for _, m := range mods {
		if m.Status == domain.ModuleEnabled {
			enabled++
		}
	}
	return map[string]any{
		"modules": len(mods), "enabled": enabled, "simulations": len(sims),
		"twins": len(twins), "edgeNodes": len(edge), "iotDevices": len(iot),
		"robots": len(robots), "droneMissions": len(drones), "experiments": len(exps),
	}, nil
}

func (d *Deps) FutureReadiness(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	mods, err := d.Modules.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if len(mods) == 0 {
		return map[string]any{"score": 0.0, "modules": 0}, nil
	}
	sum := 0.0
	for _, m := range mods {
		sum += float64(m.TRL)*10 + m.Score/2
	}
	score := sum / float64(len(mods)*2)
	if score > 100 {
		score = 100
	}
	return map[string]any{"score": score, "modules": len(mods)}, nil
}
