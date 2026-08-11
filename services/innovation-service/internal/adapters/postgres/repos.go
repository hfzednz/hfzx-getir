package postgres

import "database/sql"

// Repos groups innovation persistence adapters.
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
}

// NewRepos constructs postgres-backed repositories.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Modules: &ModuleRepo{DB: db}, Experiments: &ExperimentRepo{DB: db}, Simulations: &SimulationRepo{DB: db},
		Twins: &TwinRepo{DB: db}, Edge: &EdgeRepo{DB: db}, IoT: &IoTRepo{DB: db}, Robots: &RobotRepo{DB: db},
		Assignments: &AssignmentRepo{DB: db}, Drones: &DroneRepo{DB: db}, Blockchain: &BlockchainRepo{DB: db},
		XR: &XRRepo{DB: db}, Multimodal: &MultimodalRepo{DB: db}, Green: &GreenRepo{DB: db}, Quantum: &QuantumRepo{DB: db},
		Outbox: &OutboxRepo{DB: db},
	}
}
