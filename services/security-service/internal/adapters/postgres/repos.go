package postgres

import "database/sql"

type Repos struct {
	Policies   *PolicyRepo
	Audits     *AuditRepo
	Secrets    *SecretRepo
	Threats    *ThreatRepo
	Vulns      *VulnRepo
	Incidents  *IncidentRepo
	Compliance *ComplianceRepo
	DataGov    *DataGovRepo
	Risks      *RiskRepo
	Access     *AccessRepo
	Devices    *DeviceRepo
	AISec      *AISecRepo
	FraudSigs  *FraudRepo
	Outbox     *OutboxRepo
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Policies: &PolicyRepo{DB: db}, Audits: &AuditRepo{DB: db}, Secrets: &SecretRepo{DB: db},
		Threats: &ThreatRepo{DB: db}, Vulns: &VulnRepo{DB: db}, Incidents: &IncidentRepo{DB: db},
		Compliance: &ComplianceRepo{DB: db}, DataGov: &DataGovRepo{DB: db}, Risks: &RiskRepo{DB: db},
		Access: &AccessRepo{DB: db}, Devices: &DeviceRepo{DB: db}, AISec: &AISecRepo{DB: db},
		FraudSigs: &FraudRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
