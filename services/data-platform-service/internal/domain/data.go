package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Lake layers.
const (
	LayerBronze = "bronze"
	LayerSilver = "silver"
	LayerGold   = "gold"
)

// Compatibility modes.
const (
	CompatBackward = "backward"
	CompatForward  = "forward"
	CompatFull     = "full"
	CompatNone     = "none"
)

// Known event families (taxonomy).
const (
	FamilyUser         = "user"
	FamilyProduct      = "product"
	FamilyOrder        = "order"
	FamilyPayment      = "payment"
	FamilyDelivery     = "delivery"
	FamilyWarehouse    = "warehouse"
	FamilySupport      = "support"
	FamilyAI           = "ai"
	FamilyNotification = "notification"
	FamilySystem       = "system"
	FamilySearch       = "search"
	FamilyCampaign     = "campaign"
)

// KPI keys.
const (
	KPIRevenue       = "revenue"
	KPIOrders        = "orders"
	KPIConversion    = "conversion"
	KPIRetention     = "retention"
	KPILTV           = "ltv"
	KPICAC           = "cac"
	KPIChurn         = "churn"
	KPINPS           = "nps"
	KPICSAT          = "csat"
	KPIFulfillment   = "fulfillment_rate"
	KPIDeliverySLA   = "delivery_sla"
	KPIStockTurnover = "stock_turnover"
	KPICouponUsage   = "coupon_usage"
)

// EventSchema is a versioned analytics contract.
type EventSchema struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	Name          string // e.g. order.placed
	Family        string
	Version       int
	Compatibility string
	JSONSchema    map[string]any
	Active        bool
	CreatedAt     time.Time
}

// AnalyticsEvent is an ingested event (bronze).
type AnalyticsEvent struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	Family         string
	SchemaVersion  int
	IdempotencyKey string
	OccurredAt     time.Time
	IngestedAt     time.Time
	UserID         *uuid.UUID
	SessionID      *uuid.UUID
	CityID         *uuid.UUID
	Payload        map[string]any
	PayloadHash    string
	Layer          string
	Valid          bool
	Error          string
}

// StreamJob configures windowed aggregation.
type StreamJob struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	EventName   string
	WindowSec   int
	MetricField string // payload field to sum/count
	Agg         string // count|sum|avg
	Enabled     bool
	UpdatedAt   time.Time
}

// AggregateWindow is a real-time rollup.
type AggregateWindow struct {
	TenantID  uuid.UUID
	JobID     uuid.UUID
	WindowStart time.Time
	WindowEnd   time.Time
	Value     float64
	Count     int64
	UpdatedAt time.Time
}

// LakeDataset metadata for bronze/silver/gold.
type LakeDataset struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Layer       string
	Format      string // parquet|json
	Location    string // s3://...
	PartitionBy []string
	RetentionDays int
	UpdatedAt   time.Time
}

// FactSnapshot is a warehouse fact row snapshot (control-plane mirror; CH is SoT OLAP).
type FactSnapshot struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	FactTable  string
	GrainKey   string
	Measures   map[string]float64
	Dims       map[string]string
	AsOf       time.Time
	CreatedAt  time.Time
}

// KPIValue is a computed KPI point.
type KPIValue struct {
	TenantID  uuid.UUID
	Key       string
	Value     float64
	Unit      string
	Dims      map[string]string
	AsOf      time.Time
}

// RealtimeMetric is a live counter gauge.
type RealtimeMetric struct {
	TenantID  uuid.UUID
	Key       string
	Value     float64
	UpdatedAt time.Time
}

// Experiment definition.
type Experiment struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Key        string
	Name       string
	Status     string // draft|running|decided|archived
	Variants   []ExperimentVariant
	PrimaryKPI string
	StartedAt  *time.Time
	EndedAt    *time.Time
	Winner     string
	UpdatedAt  time.Time
}

type ExperimentVariant struct {
	Name   string
	Weight int // percent
}

// ExperimentAssignment assigns a subject to a variant.
type ExperimentAssignment struct {
	TenantID     uuid.UUID
	ExperimentID uuid.UUID
	SubjectID    uuid.UUID
	Variant      string
	AssignedAt   time.Time
}

// ReportDef is a scheduled/interactive report.
type ReportDef struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Kind      string // executive|ops|finance|marketing|warehouse|courier|support|ai
	QuerySpec map[string]any
	Schedule  string // cron or empty
	Format    string // pdf|xlsx|csv|json
	Active    bool
	UpdatedAt time.Time
}

// ReportRun is a generated report artifact metadata.
type ReportRun struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	ReportID   uuid.UUID
	Status     string
	Location   string
	RowCount   int
	CreatedAt  time.Time
	CompletedAt *time.Time
}

// ObservabilitySignal spans/logs/metrics ingest.
type ObservabilitySignal struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Kind      string // span|log|metric
	Service   string
	Name      string
	Value     float64
	Status    string
	TraceID   string
	Attrs     map[string]string
	OccurredAt time.Time
}

// AlertRule threshold/anomaly rule.
type AlertRule struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Name       string
	MetricKey  string
	Op         string // gt|lt|gte|lte
	Threshold  float64
	Severity   string
	Enabled    bool
	UpdatedAt  time.Time
}

// AlertEvent fired alert.
type AlertEvent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	RuleID    uuid.UUID
	MetricKey string
	Value     float64
	Severity  string
	Message   string
	CreatedAt time.Time
}

// CatalogAsset data catalog entry.
type CatalogAsset struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Name        string
	Type        string // event|table|mart|metric|dashboard
	Owner       string
	Description string
	Tags        []string
	Classification string // public|internal|confidential|restricted
	UpdatedAt   time.Time
}

// LineageEdge from → to.
type LineageEdge struct {
	TenantID uuid.UUID
	FromName string
	ToName   string
	Kind     string // derives|ingests|aggregates
}

// QualityCheck result.
type QualityCheck struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	AssetName string
	CheckType string // freshness|completeness|uniqueness|validity
	Passed    bool
	Score     float64
	Details   string
	CreatedAt time.Time
}

// ValidFamily reports event family.
func ValidFamily(f string) bool {
	switch f {
	case FamilyUser, FamilyProduct, FamilyOrder, FamilyPayment, FamilyDelivery,
		FamilyWarehouse, FamilySupport, FamilyAI, FamilyNotification, FamilySystem,
		FamilySearch, FamilyCampaign:
		return true
	default:
		return false
	}
}

// ValidLayer reports lake layer.
func ValidLayer(l string) bool {
	switch l {
	case LayerBronze, LayerSilver, LayerGold:
		return true
	default:
		return false
	}
}

// ValidKPI reports known KPI.
func ValidKPI(k string) bool {
	switch k {
	case KPIRevenue, KPIOrders, KPIConversion, KPIRetention, KPILTV, KPICAC, KPIChurn,
		KPINPS, KPICSAT, KPIFulfillment, KPIDeliverySLA, KPIStockTurnover, KPICouponUsage:
		return true
	default:
		return false
	}
}

// HashPayload stable hash for dedupe.
func HashPayload(payload map[string]any) string {
	b, _ := json.Marshal(payload)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ValidateRequired checks required keys exist in payload.
func ValidateRequired(payload map[string]any, required []string) error {
	for _, k := range required {
		if _, ok := payload[k]; !ok {
			return ErrInvalidArgument
		}
	}
	return nil
}

// EvalThreshold evaluates alert condition.
func EvalThreshold(value float64, op string, threshold float64) bool {
	switch op {
	case "gt":
		return value > threshold
	case "gte":
		return value >= threshold
	case "lt":
		return value < threshold
	case "lte":
		return value <= threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}

// AssignVariant deterministic weighted assignment.
func AssignVariant(subjectID uuid.UUID, variants []ExperimentVariant) string {
	if len(variants) == 0 {
		return "control"
	}
	total := 0
	for _, v := range variants {
		total += v.Weight
	}
	if total <= 0 {
		return variants[0].Name
	}
	bucket := int(subjectID[15]) % total
	acc := 0
	for _, v := range variants {
		acc += v.Weight
		if bucket < acc {
			return v.Name
		}
	}
	return variants[len(variants)-1].Name
}

// DecideWinner picks variant with highest primary metric map.
func DecideWinner(scores map[string]float64) string {
	best := ""
	bestScore := 0.0
	first := true
	for k, v := range scores {
		if first || v > bestScore {
			best, bestScore, first = k, v, false
		}
	}
	return best
}

// RequiredFieldsFromSchema extracts required property names from a simple JSON schema map.
func RequiredFieldsFromSchema(schema map[string]any) []string {
	if schema == nil {
		return nil
	}
	raw, ok := schema["required"]
	if !ok {
		return nil
	}
	switch t := raw.(type) {
	case []any:
		out := make([]string, 0, len(t))
		for _, v := range t {
			if s, ok := v.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	default:
		return nil
	}
}

// NormalizeEventName lowercases and trims.
func NormalizeEventName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
