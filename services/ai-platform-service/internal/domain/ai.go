package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Model stages in registry.
const (
	StageDev     = "dev"
	StageStaging = "staging"
	StageProd    = "prod"
	StageArchived = "archived"
)

// Deployment strategies.
const (
	DeployStable = "stable"
	DeployCanary = "canary"
	DeployShadow = "shadow"
	DeployAB     = "ab"
)

// Agent kinds.
const (
	AgentShopping   = "shopping"
	AgentOperations = "operations"
	AgentWarehouse  = "warehouse"
	AgentPricing    = "pricing"
	AgentSupport    = "support"
	AgentCampaign   = "campaign"
	AgentForecast   = "forecast"
	AgentAnalytics  = "analytics"
	AgentDeveloper  = "developer"
	AgentAdmin      = "admin"
)

// FeatureRecord is an online/offline feature vector row.
type FeatureRecord struct {
	TenantID   uuid.UUID
	EntityType string // user|product|order|courier|warehouse
	EntityID   uuid.UUID
	Name       string
	Version    int
	Values     map[string]float64
	Tags       map[string]string
	Lineage    string
	UpdatedAt  time.Time
}

// ModelCard is registry metadata.
type ModelCard struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Key          string // e.g. demand_forecast, fraud_score, embed_text
	Name         string
	Framework    string // pytorch|tensorflow|onnx|heuristic|llm
	Version      string
	Stage        string
	ArtifactURI  string
	Metrics      map[string]float64
	ApprovedBy   *uuid.UUID
	ApprovedAt   *time.Time
	DeployStrat  string
	CanaryPct    int
	Shadow       bool
	FallbackKey  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InferenceRequest routes to a model.
type InferenceRequest struct {
	TenantID   uuid.UUID
	ModelKey   string
	Version    string // optional pin
	Features   map[string]float64
	Inputs     map[string]any
	EntityType string
	EntityID   *uuid.UUID
}

// InferenceResult is model output.
type InferenceResult struct {
	ID         uuid.UUID
	ModelKey   string
	Version    string
	Stage      string
	Predictions map[string]float64
	Outputs    map[string]any
	LatencyMs  int64
	Shadow     bool
	Explain    map[string]float64 // simple feature attributions
	CreatedAt  time.Time
}

// PromptTemplate is a versioned prompt.
type PromptTemplate struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Key       string
	Locale    string
	System    string
	UserTpl   string
	Version   int
	Active    bool
	UpdatedAt time.Time
}

// ConversationMemory stores short-term LLM turns.
type ConversationMemory struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	SessionID uuid.UUID
	Role      string // system|user|assistant|tool
	Content   string
	CreatedAt time.Time
}

// LLMRequest orchestrates a completion.
type LLMRequest struct {
	TenantID   uuid.UUID
	Provider   string
	PromptKey  string
	Locale     string
	SessionID  *uuid.UUID
	UserPrompt string
	Variables  map[string]string
	Tools      []string
	RAGQuery   string
	MaxTokens  int
}

// LLMResponse is the completion result.
type LLMResponse struct {
	ID           uuid.UUID
	Provider     string
	Content      string
	ToolCalls    []ToolCall
	Blocked      bool
	BlockReason  string
	TokensIn     int
	TokensOut    int
	LatencyMs    int64
	Citations    []string
}

// ToolCall is a function/tool invocation request from the model.
type ToolCall struct {
	Name      string
	Arguments map[string]any
}

// AgentRun is an agent execution record.
type AgentRun struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Kind      string
	Input     string
	Output    string
	Steps     []AgentStep
	Status    string // succeeded|failed|blocked
	CreatedAt time.Time
}

// AgentStep is one thought/tool/observe cycle.
type AgentStep struct {
	Type    string // thought|tool|observe|answer
	Name    string
	Content string
}

// AutomationRule is a decision-engine rule.
type AutomationRule struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Enabled    bool
	Priority   int
	Conditions []RuleCondition
	Actions    []RuleAction
	RequireApproval bool
	UpdatedAt  time.Time
}

type RuleCondition struct {
	Feature string
	Op      string // gt|gte|lt|lte|eq|neq
	Value   float64
}

type RuleAction struct {
	Type   string // emit_event|set_flag|invoke_model|invoke_agent|notify_port
	Target string
	Params map[string]string
}

// AutomationRun records a rule evaluation.
type AutomationRun struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	RuleID    uuid.UUID
	Matched   bool
	Approved  bool
	Result    string
	CreatedAt time.Time
}

// DriftReport captures monitoring signal.
type DriftReport struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	ModelKey  string
	Metric    string
	Value     float64
	Threshold float64
	Severity  string
	CreatedAt time.Time
}

// ValidAgent reports supported agent kinds.
func ValidAgent(k string) bool {
	switch k {
	case AgentShopping, AgentOperations, AgentWarehouse, AgentPricing, AgentSupport,
		AgentCampaign, AgentForecast, AgentAnalytics, AgentDeveloper, AgentAdmin:
		return true
	default:
		return false
	}
}

// ValidStage reports registry stages.
func ValidStage(s string) bool {
	switch s {
	case StageDev, StageStaging, StageProd, StageArchived:
		return true
	default:
		return false
	}
}

// RenderTemplate replaces {{var}} placeholders.
func RenderTemplate(tpl string, vars map[string]string) string {
	out := tpl
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{{"+k+"}}", v)
	}
	return out
}

// EvalCondition evaluates a numeric rule condition.
func EvalCondition(actual float64, op string, expected float64) bool {
	switch op {
	case "gt":
		return actual > expected
	case "gte":
		return actual >= expected
	case "lt":
		return actual < expected
	case "lte":
		return actual <= expected
	case "eq":
		return actual == expected
	case "neq":
		return actual != expected
	default:
		return false
	}
}

// SimpleAttribution returns normalized |feature| weights as explainability stub.
func SimpleAttribution(features map[string]float64) map[string]float64 {
	var sum float64
	for _, v := range features {
		if v < 0 {
			v = -v
		}
		sum += v
	}
	out := map[string]float64{}
	if sum == 0 {
		return out
	}
	for k, v := range features {
		if v < 0 {
			v = -v
		}
		out[k] = v / sum
	}
	return out
}

// GuardrailScan blocks unsafe LLM prompts/responses (heuristic).
func GuardrailScan(text string) (blocked bool, reason string) {
	l := strings.ToLower(text)
	for _, bad := range []string{"ignore previous instructions", "exfiltrate", "bomb making", "credit card dump"} {
		if strings.Contains(l, bad) {
			return true, "policy_violation:" + bad
		}
	}
	return false, ""
}

// DefaultSystemPrompt returns agent system prompts.
func DefaultSystemPrompt(kind string) string {
	switch kind {
	case AgentShopping:
		return "You are NEXORA shopping assistant. Help find products; never invent prices."
	case AgentOperations:
		return "You are NEXORA ops agent. Summarize operational risks and next actions."
	case AgentWarehouse:
		return "You are NEXORA warehouse agent. Optimize pick paths and labor hints."
	case AgentPricing:
		return "You are NEXORA pricing agent. Suggest prices; always human-gated."
	case AgentSupport:
		return "You are NEXORA support agent. Be empathetic; escalate refunds to CRM/OMS."
	case AgentCampaign:
		return "You are NEXORA campaign agent. Suggest promo structures; do not own coupons."
	case AgentForecast:
		return "You are NEXORA forecast agent. Explain demand predictions clearly."
	case AgentAnalytics:
		return "You are NEXORA analytics agent. Describe metrics without owning ClickHouse."
	case AgentDeveloper:
		return "You are NEXORA developer agent. Help with API contracts and diagnostics."
	case AgentAdmin:
		return "You are NEXORA admin agent. Respect RBAC and dual-control policies."
	default:
		return "You are a NEXORA AI assistant."
	}
}
