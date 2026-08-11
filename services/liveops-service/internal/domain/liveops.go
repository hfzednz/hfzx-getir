package domain

import (
	"hash/fnv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EvalContext is targeting input for flags/config/experiments.
type EvalContext struct {
	SubjectID   string
	Country     string
	City        string
	WarehouseID string
	Segment     string
	Membership  string
	Language    string
	DeviceID    string
	OS          string
	AppVersion  string
	Role        string
	Attributes  map[string]string
}

// FeatureFlag definition.
type FeatureFlag struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Key         string
	Description string
	Enabled     bool
	Percentage  int // 0..100 sticky by subject
	Rules       []TargetRule
	DependsOn   []string // other flag keys that must be on
	Version     int
	EmergencyOff bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TargetRule struct {
	Kind     string // country|city|warehouse|segment|role|os|app_version|device
	Op       string // in|not_in|gte|eq
	Values   []string
	Percent  int // optional nested percent
}

// FlagEvaluation result.
type FlagEvaluation struct {
	Key     string
	Enabled bool
	Reason  string
	Variant string
}

// ConfigDocument remote config.
type ConfigDocument struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Key       string
	Namespace string // app|ui|pricing_params|limits|checkout|search|home
	Payload   map[string]any
	Version   int
	Status    string // draft|published|archived
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Experiment A/B/MVT.
type Experiment struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	Key         string
	Name        string
	Status      string // draft|running|paused|completed|rolled_back
	Kind        string // ab|aa|mvt|canary
	Hypothesis  string
	Variants    []Variant
	PrimaryMetric string
	StartedAt   *time.Time
	EndedAt     *time.Time
	Winner      string
	CreatedAt   time.Time
}

type Variant struct {
	Key        string
	Weight     int // relative weight
	Payload    map[string]any
}

type Assignment struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	ExperimentID uuid.UUID
	SubjectID    string
	VariantKey   string
	AssignedAt   time.Time
}

// LiveOpsEvent calendar.
type LiveOpsEvent struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	Key       string
	Kind      string // flash|weekend|holiday|seasonal|regional|warehouse|emergency|custom
	Title     string
	Status    string // scheduled|active|ended|cancelled
	StartsAt  time.Time
	EndsAt    time.Time
	Scopes    []TargetRule
	ConfigPatch map[string]any
	CreatedAt time.Time
}

// ChangeRequest approval for dangerous mutations.
type ChangeRequest struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Kind       string // flag|config|experiment|rollback
	SubjectKey string
	Payload    map[string]any
	Status     string // pending|approved|rejected
	CreatedAt  time.Time
	DecidedAt  *time.Time
}

// RollbackRecord.
type RollbackRecord struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	Kind       string // flag|config|experiment|emergency
	SubjectKey string
	FromVersion int
	ToVersion   int
	Reason     string
	CreatedAt  time.Time
}

func NormalizeKey(k string) string {
	return strings.ToLower(strings.TrimSpace(k))
}

func ValidExperimentKind(k string) bool {
	switch k {
	case "ab", "aa", "mvt", "canary":
		return true
	default:
		return false
	}
}

// StickyBucket 0..99 from subject+salt.
func StickyBucket(subject, salt string) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(subject + "|" + salt))
	return int(h.Sum32() % 100)
}

// MatchRules returns true if all rules match (empty rules = match all).
func MatchRules(ctx EvalContext, rules []TargetRule) bool {
	for _, r := range rules {
		if !matchOne(ctx, r) {
			return false
		}
	}
	return true
}

func matchOne(ctx EvalContext, r TargetRule) bool {
	val := ""
	switch r.Kind {
	case "country":
		val = ctx.Country
	case "city":
		val = ctx.City
	case "warehouse":
		val = ctx.WarehouseID
	case "segment":
		val = ctx.Segment
	case "role":
		val = ctx.Role
	case "os":
		val = ctx.OS
	case "app_version":
		val = ctx.AppVersion
	case "device":
		val = ctx.DeviceID
	case "membership":
		val = ctx.Membership
	case "language":
		val = ctx.Language
	default:
		if ctx.Attributes != nil {
			val = ctx.Attributes[r.Kind]
		}
	}
	val = strings.ToLower(strings.TrimSpace(val))
	switch r.Op {
	case "", "in", "eq":
		if len(r.Values) == 0 {
			return true
		}
		for _, v := range r.Values {
			if strings.ToLower(strings.TrimSpace(v)) == val {
				return true
			}
		}
		return false
	case "not_in":
		for _, v := range r.Values {
			if strings.ToLower(strings.TrimSpace(v)) == val {
				return false
			}
		}
		return true
	case "gte":
		if len(r.Values) == 0 {
			return true
		}
		return compareVersion(val, r.Values[0]) >= 0
	default:
		return false
	}
}

func compareVersion(a, b string) int {
	as := strings.Split(a, ".")
	bs := strings.Split(b, ".")
	n := len(as)
	if len(bs) > n {
		n = len(bs)
	}
	for i := 0; i < n; i++ {
		ai, bi := 0, 0
		if i < len(as) {
			ai = atoi(as[i])
		}
		if i < len(bs) {
			bi = atoi(bs[i])
		}
		if ai < bi {
			return -1
		}
		if ai > bi {
			return 1
		}
	}
	return 0
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// EvaluateFlag applies emergency, deps, rules, percentage.
func EvaluateFlag(f FeatureFlag, ctx EvalContext, depEnabled map[string]bool) FlagEvaluation {
	if f.EmergencyOff || !f.Enabled {
		return FlagEvaluation{Key: f.Key, Enabled: false, Reason: "disabled"}
	}
	for _, dep := range f.DependsOn {
		if !depEnabled[NormalizeKey(dep)] {
			return FlagEvaluation{Key: f.Key, Enabled: false, Reason: "dependency_off:" + dep}
		}
	}
	if !MatchRules(ctx, f.Rules) {
		return FlagEvaluation{Key: f.Key, Enabled: false, Reason: "rules_mismatch"}
	}
	pct := f.Percentage
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	if pct < 100 {
		b := StickyBucket(ctx.SubjectID, f.Key)
		if b >= pct {
			return FlagEvaluation{Key: f.Key, Enabled: false, Reason: "percentage"}
		}
	}
	return FlagEvaluation{Key: f.Key, Enabled: true, Reason: "matched", Variant: "on"}
}

// AssignVariant sticky weighted assignment.
func AssignVariant(exp Experiment, subjectID string) (string, error) {
	if exp.Status != "running" {
		return "", ErrExperimentClosed
	}
	total := 0
	for _, v := range exp.Variants {
		if v.Weight > 0 {
			total += v.Weight
		}
	}
	if total <= 0 || len(exp.Variants) == 0 {
		return "", ErrInvalidArgument
	}
	bucket := StickyBucket(subjectID, exp.Key)
	// map 0..99 onto weights
	cursor := 0
	target := bucket * total / 100
	if target >= total {
		target = total - 1
	}
	for _, v := range exp.Variants {
		w := v.Weight
		if w <= 0 {
			continue
		}
		if target < cursor+w {
			return v.Key, nil
		}
		cursor += w
	}
	return exp.Variants[len(exp.Variants)-1].Key, nil
}

// PickWinner simple conversion-rate winner (rates 0..1).
func PickWinner(variants []string, rates map[string]float64) string {
	best := ""
	bestRate := -1.0
	for _, v := range variants {
		r := rates[v]
		if r > bestRate {
			bestRate = r
			best = v
		}
	}
	return best
}

// Significant if absolute lift >= minLift and sample hint ok (simplified).
func Significant(control, treatment float64, minLift float64) bool {
	if control <= 0 {
		return treatment >= minLift
	}
	lift := (treatment - control) / control
	if lift < 0 {
		lift = -lift
	}
	return lift >= minLift
}

// EventActive now within window and scopes match.
func EventActive(e LiveOpsEvent, now time.Time, ctx EvalContext) bool {
	if e.Status != "active" && e.Status != "scheduled" {
		return false
	}
	if now.Before(e.StartsAt) || now.After(e.EndsAt) {
		return false
	}
	return MatchRules(ctx, e.Scopes)
}
