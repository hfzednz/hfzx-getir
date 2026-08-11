package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type SuiteKind string

const (
	SuiteUnit        SuiteKind = "unit"
	SuiteIntegration SuiteKind = "integration"
	SuiteAPI         SuiteKind = "api"
	SuiteUI          SuiteKind = "ui"
	SuiteE2E         SuiteKind = "e2e"
	SuitePerf        SuiteKind = "perf"
	SuiteSecurity    SuiteKind = "security"
	SuiteA11y        SuiteKind = "accessibility"
	SuiteChaos       SuiteKind = "chaos"
	SuiteDB          SuiteKind = "database"
	SuiteInfra       SuiteKind = "infra"
	SuiteAI          SuiteKind = "ai"
	SuiteObs         SuiteKind = "observability"
)

type Suite struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Kind        SuiteKind `json:"kind"`
	Owner       string    `json:"owner"`
	Path        string    `json:"path"` // repo-relative suite path
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RunStatus string

const (
	RunQueued    RunStatus = "queued"
	RunRunning   RunStatus = "running"
	RunPassed    RunStatus = "passed"
	RunFailed    RunStatus = "failed"
	RunCancelled RunStatus = "cancelled"
)

type TestRun struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	SuiteKey    string    `json:"suiteKey"`
	Kind        SuiteKind `json:"kind"`
	Status      RunStatus `json:"status"`
	Trigger     string    `json:"trigger"` // ci|nightly|manual|local
	CommitSHA   string    `json:"commitSha"`
	Branch      string    `json:"branch"`
	Environment string    `json:"environment"`
	StartedAt   time.Time `json:"startedAt"`
	FinishedAt  *time.Time `json:"finishedAt,omitempty"`
	Summary     RunSummary `json:"summary"`
}

type RunSummary struct {
	Total    int     `json:"total"`
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Skipped  int     `json:"skipped"`
	Flaky    int     `json:"flaky"`
	DurationMs int64 `json:"durationMs"`
}

type TestCaseResult struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	RunID     uuid.UUID `json:"runId"`
	Name      string    `json:"name"`
	ClassName string    `json:"className"`
	Status    string    `json:"status"` // passed|failed|skipped|flaky
	DurationMs int64    `json:"durationMs"`
	Message   string    `json:"message,omitempty"`
}

type CoverageReport struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	RunID        uuid.UUID `json:"runId"`
	Service      string    `json:"service"`
	LinePct      float64   `json:"linePct"`
	BranchPct    float64   `json:"branchPct"`
	APIPct       float64   `json:"apiPct"`
	WorkflowPct  float64   `json:"workflowPct"`
	CreatedAt    time.Time `json:"createdAt"`
}

type GateKind string

const (
	GateCompile   GateKind = "compile"
	GateLint      GateKind = "lint"
	GateUnitCov   GateKind = "unit_coverage"
	GateIntegration GateKind = "integration"
	GatePerf      GateKind = "performance"
	GateSecurity  GateKind = "security"
	GateA11y      GateKind = "accessibility"
	GateChaos     GateKind = "chaos"
)

type GatePolicy struct {
	ID        uuid.UUID         `json:"id"`
	TenantID  uuid.UUID         `json:"tenantId"`
	Key       string            `json:"key"`
	Kind      GateKind          `json:"kind"`
	Thresholds map[string]float64 `json:"thresholds"`
	Required  bool              `json:"required"`
	CreatedAt time.Time         `json:"createdAt"`
}

type GateEvaluation struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	PolicyKey string    `json:"policyKey"`
	RunID     uuid.UUID `json:"runId"`
	Passed    bool      `json:"passed"`
	Score     float64   `json:"score"`
	Details   map[string]any `json:"details"`
	CreatedAt time.Time `json:"createdAt"`
}

type CertKind string

const (
	CertReleaseReadiness CertKind = "release_readiness"
	CertRegression       CertKind = "regression"
	CertPerformance      CertKind = "performance"
	CertSecurity         CertKind = "security"
	CertCompliance       CertKind = "compliance"
	CertProduction       CertKind = "production"
)

type Certification struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	Kind        CertKind  `json:"kind"`
	Version     string    `json:"version"`
	CommitSHA   string    `json:"commitSha"`
	Status      string    `json:"status"` // issued|revoked|pending
	GateResults []uuid.UUID `json:"gateResultIds"`
	IssuedAt    time.Time `json:"issuedAt"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Notes       string    `json:"notes"`
}

type FlakyRecord struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	TestName   string    `json:"testName"`
	SuiteKey   string    `json:"suiteKey"`
	FailCount  int       `json:"failCount"`
	PassCount  int       `json:"passCount"`
	LastStatus string    `json:"lastStatus"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type PerfMetric struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	RunID     uuid.UUID `json:"runId"`
	Scenario  string    `json:"scenario"`
	P50Ms     float64   `json:"p50Ms"`
	P95Ms     float64   `json:"p95Ms"`
	P99Ms     float64   `json:"p99Ms"`
	ErrorRate float64   `json:"errorRate"`
	RPS       float64   `json:"rps"`
	CreatedAt time.Time `json:"createdAt"`
}

type SecurityFinding struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	RunID     uuid.UUID `json:"runId"`
	Tool      string    `json:"tool"` // zap|sast|deps
	Severity  string    `json:"severity"`
	Title     string    `json:"title"`
	Target    string    `json:"target"`
	CreatedAt time.Time `json:"createdAt"`
}

func ValidateSuite(s Suite) error {
	if s.TenantID == uuid.Nil || strings.TrimSpace(s.Key) == "" || s.Kind == "" {
		return ErrInvalidArgument
	}
	return nil
}

func EvaluateUnitCoverage(linePct float64, minLine float64) (bool, float64) {
	if minLine <= 0 {
		minLine = 70
	}
	return linePct >= minLine, linePct
}

func EvaluatePerf(p95Ms, maxP95 float64, errorRate, maxErr float64) (bool, float64) {
	if maxP95 <= 0 {
		maxP95 = 500
	}
	if maxErr <= 0 {
		maxErr = 0.01
	}
	ok := p95Ms <= maxP95 && errorRate <= maxErr
	score := 1.0
	if p95Ms > maxP95 {
		score -= 0.5
	}
	if errorRate > maxErr {
		score -= 0.5
	}
	if score < 0 {
		score = 0
	}
	return ok, score
}

func EvaluateSecurity(critical, high int, allowHigh bool) (bool, float64) {
	if critical > 0 {
		return false, 0
	}
	if high > 0 && !allowHigh {
		return false, 0.5
	}
	return true, 1
}

func CanComplete(run TestRun) error {
	if run.Status != RunRunning && run.Status != RunQueued {
		return ErrIllegalTransition
	}
	return nil
}

func AllRequiredGatesPassed(evals []GateEvaluation, policies []GatePolicy) bool {
	required := map[string]bool{}
	for _, p := range policies {
		if p.Required {
			required[p.Key] = true
		}
	}
	if len(required) == 0 {
		return true
	}
	passed := map[string]bool{}
	for _, e := range evals {
		passed[e.PolicyKey] = e.Passed
	}
	return allTrue(required, passed)
}

func allTrue(required, passed map[string]bool) bool {
	for k := range required {
		if !passed[k] {
			return false
		}
	}
	return true
}
