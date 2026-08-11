package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuditDomain string

const (
	AuditArchitecture AuditDomain = "architecture"
	AuditPerformance  AuditDomain = "performance"
	AuditSecurity     AuditDomain = "security"
	AuditInfrastructure AuditDomain = "infrastructure"
	AuditDatabase     AuditDomain = "database"
	AuditAPI          AuditDomain = "api"
	AuditAI           AuditDomain = "ai"
	AuditOperational  AuditDomain = "operational"
	AuditDependency   AuditDomain = "dependency"
)

type Audit struct {
	ID        uuid.UUID   `json:"id"`
	TenantID  uuid.UUID   `json:"tenantId"`
	Domain    AuditDomain `json:"domain"`
	Title     string      `json:"title"`
	Status    string      `json:"status"` // open|completed
	CreatedAt time.Time   `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type FindingSeverity string

const (
	SeverityCritical FindingSeverity = "critical"
	SeverityHigh     FindingSeverity = "high"
	SeverityMedium   FindingSeverity = "medium"
	SeverityLow      FindingSeverity = "low"
)

type Finding struct {
	ID         uuid.UUID       `json:"id"`
	TenantID   uuid.UUID       `json:"tenantId"`
	AuditID    uuid.UUID       `json:"auditId"`
	Code       string          `json:"code"`
	Title      string          `json:"title"`
	Severity   FindingSeverity `json:"severity"`
	Status     string          `json:"status"` // open|resolved
	Resolution string          `json:"resolution"`
	CreatedAt  time.Time       `json:"createdAt"`
	ResolvedAt *time.Time      `json:"resolvedAt,omitempty"`
}

type BenchmarkKind string

const (
	BenchOrders        BenchmarkKind = "orders_per_sec"
	BenchSearch        BenchmarkKind = "search_per_sec"
	BenchPayments      BenchmarkKind = "payments_per_sec"
	BenchAI            BenchmarkKind = "ai_per_sec"
	BenchNotifications BenchmarkKind = "notifications_per_sec"
	BenchTracking      BenchmarkKind = "delivery_updates_per_sec"
	BenchDBTPS         BenchmarkKind = "database_tps"
	BenchAPILatency    BenchmarkKind = "api_latency_p99_ms"
)

type BenchmarkRun struct {
	ID         uuid.UUID     `json:"id"`
	TenantID   uuid.UUID     `json:"tenantId"`
	Kind       BenchmarkKind `json:"kind"`
	Value      float64       `json:"value"`
	Target     float64       `json:"target"`
	Passed     bool          `json:"passed"`
	Scenario   string        `json:"scenario"`
	CreatedAt  time.Time     `json:"createdAt"`
}

type CapacityScenario struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	PeakRPS     int       `json:"peakRps"`
	Regions     []string  `json:"regions"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"createdAt"`
}

type ChaosKind string

const (
	ChaosNode     ChaosKind = "node"
	ChaosCluster  ChaosKind = "cluster"
	ChaosRegion   ChaosKind = "region"
	ChaosDatabase ChaosKind = "database"
	ChaosKafka    ChaosKind = "kafka"
	ChaosRedis    ChaosKind = "redis"
	ChaosPayment  ChaosKind = "payment"
	ChaosAI       ChaosKind = "ai"
)

type ChaosExperiment struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Kind        ChaosKind `json:"kind"`
	Name        string    `json:"name"`
	Status      string    `json:"status"` // planned|running|passed|failed
	RecoverySec int       `json:"recoverySec"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type TuningProfile struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	Key       string    `json:"key"`
	Layer     string    `json:"layer"` // postgres|redis|kafka|k8s|envoy|ai
	URI       string    `json:"uri"`
	Applied   bool      `json:"applied"`
	CreatedAt time.Time `json:"createdAt"`
	AppliedAt *time.Time `json:"appliedAt,omitempty"`
}

type Certificate struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Version     string    `json:"version"`
	Status      string    `json:"status"` // pending|issued|revoked
	Gates       map[string]bool `json:"gates"`
	IssuedAt    *time.Time `json:"issuedAt,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// DefaultTargets are hyperscale certification targets.
func DefaultTargets() map[BenchmarkKind]float64 {
	return map[BenchmarkKind]float64{
		BenchOrders: 5000, BenchSearch: 20000, BenchPayments: 3000,
		BenchAI: 500, BenchNotifications: 10000, BenchTracking: 15000,
		BenchDBTPS: 25000, BenchAPILatency: 150, // p99 ms — lower is better
	}
}

func BenchmarkPasses(kind BenchmarkKind, value, target float64) bool {
	if kind == BenchAPILatency {
		return value > 0 && value <= target
	}
	return value >= target
}

func ValidateFinding(f Finding) error {
	if f.TenantID == uuid.Nil || f.AuditID == uuid.Nil || strings.TrimSpace(f.Code) == "" || f.Title == "" {
		return ErrInvalidArgument
	}
	return nil
}

func CertGatesRequired() []string {
	return []string{
		"performance", "security", "scalability", "reliability",
		"observability", "disaster_recovery", "zero_critical_findings",
	}
}

func DefaultCapacityCatalog() []CapacityScenario {
	return []CapacityScenario{
		{Key: "black_friday", Name: "Black Friday Peak", PeakRPS: 50000, Regions: []string{"tr", "eu"}, Notes: "3x normal peak"},
		{Key: "holiday", Name: "Holiday Traffic", PeakRPS: 35000, Regions: []string{"tr"}, Notes: "sustained 48h"},
		{Key: "marketing_spike", Name: "Campaign Spike", PeakRPS: 25000, Regions: []string{"tr", "eu"}, Notes: "15m burst"},
		{Key: "regional_spike", Name: "Regional Spike", PeakRPS: 20000, Regions: []string{"tr-ist"}, Notes: "city launch"},
		{Key: "disaster", Name: "DR Failover Load", PeakRPS: 15000, Regions: []string{"eu-west"}, Notes: "failover absorption"},
	}
}
