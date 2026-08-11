package app_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nexora/innovation-service/internal/app"
	"github.com/nexora/innovation-service/internal/app/memory"
	"github.com/nexora/innovation-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Modules: r.Modules, Experiments: r.Experiments, Simulations: r.Simulations,
		Twins: r.Twins, Edge: r.Edge, IoT: r.IoT, Robots: r.Robots,
		Assignments: r.Assignments, Drones: r.Drones, Blockchain: r.Blockchain,
		XR: r.XR, Multimodal: r.Multimodal, Green: r.Green, Quantum: r.Quantum,
		Outbox: r.Outbox, LiveOps: r.LiveOps, AI: r.AI, Metrics: r.Metrics,
		Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestInnovationFlows(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()

	if err := d.SeedCatalog(ctx, tid); err != nil {
		t.Fatal(err)
	}
	mods, _ := d.Modules.List(ctx, tid)
	if len(mods) < 15 {
		t.Fatal(len(mods))
	}

	// TRL 7 green.carbon should enable
	m, err := d.EnableModule(ctx, tid, "green.carbon")
	if err != nil || m.Status != domain.ModuleEnabled {
		t.Fatalf("%+v %v", m, err)
	}

	// low TRL blockchain should fail
	if _, err := d.EnableModule(ctx, tid, "chain.trace"); err != domain.ErrNotReady {
		t.Fatalf("want not ready got %v", err)
	}

	mod, _ := d.Modules.GetByKey(ctx, tid, "sim.demand")
	exp, err := d.CreateExperiment(ctx, domain.ResearchExperiment{
		TenantID: tid, ModuleID: mod.ID, Name: "Demand uplift", Hypothesis: "sim accuracy > 0.7",
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = exp

	sim, err := d.StartSimulation(ctx, domain.SimulationRun{
		TenantID: tid, Kind: domain.SimDemand, Name: "peak-hour",
		Params: map[string]any{"iterations": float64(200)},
	})
	if err != nil {
		t.Fatal(err)
	}
	done, err := d.CompleteSimulation(ctx, tid, sim.ID)
	if err != nil || done.Status != domain.SimCompleted || done.Accuracy < 0.7 {
		t.Fatalf("%+v %v", done, err)
	}

	_, err = d.UpsertTwin(ctx, domain.DigitalTwin{
		TenantID: tid, Kind: domain.TwinWarehouse, RefKey: "wh-opaque-1", Name: "WH Twin",
		ModelURI: "twin://warehouse/wh-opaque-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.RegisterEdge(ctx, domain.EdgeNode{TenantID: tid, Key: "edge-ist-1", Region: "tr-ist"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.ConnectIoT(ctx, domain.IoTDevice{
		TenantID: tid, DeviceKey: "temp-01", Kind: domain.IoTTemp, Location: "cold-aisle",
	})
	if err != nil {
		t.Fatal(err)
	}

	robot, err := d.RegisterRobot(ctx, domain.Robot{TenantID: tid, Key: "picker-1", Kind: domain.RobotPicker})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.AssignRobot(ctx, tid, robot.ID, "task-opaque-9")
	if err != nil {
		t.Fatal(err)
	}

	_, err = d.CreateDroneMission(ctx, domain.DroneMission{
		TenantID: tid, DroneKey: "drone-7", OrderRef: "ord-opaque", LandingZone: "lz-a1",
	})
	if err != nil {
		t.Fatal(err)
	}

	_, _ = d.RegisterBlockchainHook(ctx, domain.BlockchainHook{TenantID: tid, Purpose: "traceability", ChainRef: "optional"})
	_, _ = d.RegisterXR(ctx, domain.XRExperience{TenantID: tid, Kind: "ar_shopping", AssetURI: "xr://sku/demo"})
	_, _ = d.StartMultimodal(ctx, domain.MultimodalSession{TenantID: tid, SubjectID: "u1", Modes: []string{"voice", "vision"}})
	_, _ = d.UpsertGreen(ctx, domain.GreenMetric{TenantID: tid, Period: "2026-08", CarbonGrams: 12000, EnergyWh: 5000, SavingsPercent: 8.2})
	_, _ = d.RegisterQuantum(ctx, domain.QuantumHook{TenantID: tid, Kind: "crypto_pqc", Adapter: "research://pqc"})

	st, _ := d.AdminStats(ctx, tid)
	if st["enabled"].(int) < 1 || st["simulations"].(int) < 1 {
		t.Fatal(st)
	}
	ready, _ := d.FutureReadiness(ctx, tid)
	if ready["score"].(float64) <= 0 {
		t.Fatal(ready)
	}
}
