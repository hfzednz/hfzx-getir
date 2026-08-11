package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type TRL int // Technology Readiness Level 1..9

type ModuleStatus string

const (
	ModuleDraft       ModuleStatus = "draft"
	ModuleIncubating  ModuleStatus = "incubating"
	ModuleExperimental ModuleStatus = "experimental"
	ModuleEnabled     ModuleStatus = "enabled"
	ModuleRetired     ModuleStatus = "retired"
)

type InnovationDomain string

const (
	DomainDigitalTwin   InnovationDomain = "digital_twin"
	DomainSimulation    InnovationDomain = "simulation"
	DomainEdge          InnovationDomain = "edge"
	DomainIoT           InnovationDomain = "iot"
	DomainRobotics      InnovationDomain = "robotics"
	DomainDrone         InnovationDomain = "drone"
	DomainAutonomous    InnovationDomain = "autonomous"
	DomainBlockchain    InnovationDomain = "blockchain"
	DomainXR            InnovationDomain = "xr"
	DomainMultimodal    InnovationDomain = "multimodal"
	DomainAdvancedAI    InnovationDomain = "advanced_ai"
	DomainQuantum       InnovationDomain = "quantum"
	DomainGreen         InnovationDomain = "green"
	DomainLab           InnovationDomain = "lab"
)

// InnovationModule is an optional, isolated future capability.
type InnovationModule struct {
	ID          uuid.UUID        `json:"id"`
	TenantID    uuid.UUID        `json:"tenantId"`
	Key         string           `json:"key"`
	Name        string           `json:"name"`
	Domain      InnovationDomain `json:"domain"`
	Status      ModuleStatus     `json:"status"`
	TRL         TRL              `json:"trl"`
	Score       float64          `json:"score"` // innovation scoring 0..100
	SandboxOnly bool             `json:"sandboxOnly"`
	Description string           `json:"description"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

type ResearchExperiment struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	ModuleID    uuid.UUID `json:"moduleId"`
	Name        string    `json:"name"`
	Hypothesis  string    `json:"hypothesis"`
	Status      string    `json:"status"` // draft|running|completed|failed
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type SimulationKind string

const (
	SimDemand    SimulationKind = "demand"
	SimCustomer  SimulationKind = "customer"
	SimCourier   SimulationKind = "courier"
	SimWarehouse SimulationKind = "warehouse"
	SimPricing   SimulationKind = "pricing"
	SimDisaster  SimulationKind = "disaster"
	SimTraffic   SimulationKind = "traffic"
	SimInventory SimulationKind = "inventory"
	SimOps       SimulationKind = "operations"
)

type SimulationStatus string

const (
	SimPending   SimulationStatus = "pending"
	SimRunning   SimulationStatus = "running"
	SimCompleted SimulationStatus = "completed"
	SimFailed    SimulationStatus = "failed"
)

type SimulationRun struct {
	ID          uuid.UUID         `json:"id"`
	TenantID    uuid.UUID         `json:"tenantId"`
	Kind        SimulationKind    `json:"kind"`
	Name        string            `json:"name"`
	Status      SimulationStatus  `json:"status"`
	Params      map[string]any    `json:"params"`
	Accuracy    float64           `json:"accuracy"` // 0..1 estimated
	ResultSummary string          `json:"resultSummary"`
	StartedAt   *time.Time        `json:"startedAt,omitempty"`
	CompletedAt *time.Time        `json:"completedAt,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

type TwinKind string

const (
	TwinWarehouse TwinKind = "warehouse"
	TwinFleet     TwinKind = "fleet"
	TwinCity      TwinKind = "city"
)

type DigitalTwin struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Kind      TwinKind  `json:"kind"`
	RefKey    string    `json:"refKey"` // opaque warehouse/fleet/city ref
	Name      string    `json:"name"`
	ModelURI  string    `json:"modelUri"`
	Version   string    `json:"version"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type EdgeNode struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Key       string    `json:"key"`
	Region    string    `json:"region"`
	Caps      []string  `json:"capabilities"` // inference|cache|sync|analytics
	Status    string    `json:"status"`       // online|offline|draining
	CreatedAt time.Time `json:"createdAt"`
}

type IoTDeviceKind string

const (
	IoTTemp     IoTDeviceKind = "temperature"
	IoTHumidity IoTDeviceKind = "humidity"
	IoTGPS      IoTDeviceKind = "gps"
	IoTVehicle  IoTDeviceKind = "vehicle_telemetry"
	IoTShelf    IoTDeviceKind = "smart_shelf"
	IoTEnergy   IoTDeviceKind = "energy"
	IoTGeneric  IoTDeviceKind = "warehouse_sensor"
)

type IoTDevice struct {
	ID         uuid.UUID     `json:"id"`
	TenantID   uuid.UUID     `json:"tenantId"`
	DeviceKey  string        `json:"deviceKey"`
	Kind       IoTDeviceKind `json:"kind"`
	Location   string        `json:"location"`
	Connected  bool          `json:"connected"`
	LastSeenAt *time.Time    `json:"lastSeenAt,omitempty"`
	CreatedAt  time.Time     `json:"createdAt"`
}

type RobotKind string

const (
	RobotPicker  RobotKind = "picking"
	RobotSorter  RobotKind = "sorting"
	RobotWarehouse RobotKind = "warehouse"
	RobotDelivery RobotKind = "delivery"
	RobotAutonomous RobotKind = "autonomous"
)

type Robot struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Key       string    `json:"key"`
	Kind      RobotKind `json:"kind"`
	Status    string    `json:"status"` // idle|assigned|busy|offline
	CreatedAt time.Time `json:"createdAt"`
}

type RobotAssignment struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	RobotID    uuid.UUID `json:"robotId"`
	TaskRef    string    `json:"taskRef"` // opaque WMS/dispatch ref
	Status     string    `json:"status"`
	AssignedAt time.Time `json:"assignedAt"`
}

type DroneMission struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	DroneKey    string    `json:"droneKey"`
	OrderRef    string    `json:"orderRef"` // opaque
	LandingZone string    `json:"landingZone"`
	Status      string    `json:"status"` // planned|inflight|completed|aborted
	Compliance  []string  `json:"compliance"`
	CreatedAt   time.Time `json:"createdAt"`
}

type BlockchainHook struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Purpose   string    `json:"purpose"` // traceability|certificate|smart_contract
	ChainRef  string    `json:"chainRef"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

type XRExperience struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Kind      string    `json:"kind"` // ar_shopping|vr_warehouse|mr_training|product_3d
	AssetURI  string    `json:"assetUri"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

type MultimodalSession struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	SubjectID string    `json:"subjectId"`
	Modes     []string  `json:"modes"` // vision|voice|gesture|camera
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

type GreenMetric struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenantId"`
	Period         string    `json:"period"` // YYYY-MM
	CarbonGrams    int64     `json:"carbonGrams"`
	EnergyWh       int64     `json:"energyWh"`
	SavingsPercent float64   `json:"savingsPercent"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type QuantumHook struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Kind      string    `json:"kind"` // crypto_pqc|optimization|research
	Adapter   string    `json:"adapter"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"createdAt"`
}

func ValidateModule(m InnovationModule) error {
	if m.TenantID == uuid.Nil || strings.TrimSpace(m.Key) == "" || m.Name == "" || m.Domain == "" {
		return ErrInvalidArgument
	}
	if m.TRL < 1 || m.TRL > 9 {
		return ErrInvalidArgument
	}
	return nil
}

func CanEnable(m InnovationModule) error {
	if m.SandboxOnly && m.TRL < 6 {
		return ErrNotReady
	}
	if m.TRL < 4 {
		return ErrNotReady
	}
	return nil
}

func DefaultInnovationCatalog() []InnovationModule {
	defs := []struct {
		key, name string
		domain    InnovationDomain
		trl       TRL
	}{
		{"twin.warehouse", "Warehouse Digital Twin", DomainDigitalTwin, 6},
		{"twin.fleet", "Fleet Digital Twin", DomainDigitalTwin, 5},
		{"twin.city", "City Digital Twin", DomainDigitalTwin, 4},
		{"sim.demand", "Demand Simulation", DomainSimulation, 7},
		{"sim.disaster", "Disaster Simulation", DomainSimulation, 5},
		{"edge.inference", "Edge Inference", DomainEdge, 6},
		{"iot.sensors", "Warehouse IoT Sensors", DomainIoT, 7},
		{"robot.picking", "Picking Robots", DomainRobotics, 5},
		{"drone.delivery", "Drone Delivery", DomainDrone, 4},
		{"auto.delivery", "Autonomous Delivery", DomainAutonomous, 4},
		{"chain.trace", "Supply Traceability Ledger", DomainBlockchain, 3},
		{"xr.ar_shop", "AR Shopping", DomainXR, 5},
		{"mm.voice", "Multimodal Voice Assistant", DomainMultimodal, 6},
		{"ai.multiagent", "Multi-Agent Planning", DomainAdvancedAI, 5},
		{"quantum.pqc", "Quantum-Safe Crypto Hooks", DomainQuantum, 3},
		{"green.carbon", "Carbon Tracking", DomainGreen, 7},
	}
	out := make([]InnovationModule, 0, len(defs))
	for _, d := range defs {
		out = append(out, InnovationModule{
			Key: d.key, Name: d.name, Domain: d.domain, TRL: d.trl,
			Status: ModuleIncubating, SandboxOnly: d.trl < 6, Score: float64(d.trl) * 10,
		})
	}
	return out
}

func EstimateSimAccuracy(kind SimulationKind, params map[string]any) float64 {
	base := 0.72
	switch kind {
	case SimDemand, SimPricing:
		base = 0.78
	case SimDisaster:
		base = 0.55
	case SimWarehouse, SimInventory:
		base = 0.80
	}
	if params != nil {
		if n, ok := params["iterations"].(float64); ok && n > 100 {
			base += 0.05
		}
	}
	if base > 0.95 {
		return 0.95
	}
	return base
}
