package policy

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/domain"
)

// Attributes are the ABAC request context evaluated against a decision.
type Attributes struct {
	TenantID    *uuid.UUID
	CityID      *uuid.UUID
	WarehouseID *uuid.UUID
	CountryCode string
	RiskScore   float64
	MFALevel    int // 0 = none, 1 = single factor, 2+ = step-up / multi
	TrustedDevice bool
}

// Requirement describes the minimum attributes needed for an action.
type Requirement struct {
	// RequiredTenant, when set, must equal attrs.TenantID.
	RequiredTenant *uuid.UUID
	// RequiredCity, when set, must equal attrs.CityID.
	RequiredCity *uuid.UUID
	// RequiredWarehouse, when set, must equal attrs.WarehouseID.
	RequiredWarehouse *uuid.UUID
	// MaxRisk is the inclusive ceiling; scores above deny.
	MaxRisk float64
	// MinMFALevel is the minimum authentication assurance.
	MinMFALevel int
	// RequireTrustedDevice forces a trusted device when true.
	RequireTrustedDevice bool
}

// Decision is the ABAC evaluation result.
type Decision struct {
	Allow   bool
	Reasons []string
}

// Evaluate applies attribute-based checks for tenant/city/warehouse/mfa/risk.
func Evaluate(attrs Attributes, req Requirement) Decision {
	var reasons []string

	if attrs.RiskScore < 0 || attrs.RiskScore > 100 {
		reasons = append(reasons, "risk_score out of range")
	}
	if req.MaxRisk < 0 || req.MaxRisk > 100 {
		reasons = append(reasons, "requirement max_risk out of range")
	}

	if req.RequiredTenant != nil {
		if attrs.TenantID == nil || *attrs.TenantID != *req.RequiredTenant {
			reasons = append(reasons, "tenant mismatch")
		}
	}
	if req.RequiredCity != nil {
		if attrs.CityID == nil || *attrs.CityID != *req.RequiredCity {
			reasons = append(reasons, "city mismatch")
		}
	}
	if req.RequiredWarehouse != nil {
		if attrs.WarehouseID == nil || *attrs.WarehouseID != *req.RequiredWarehouse {
			reasons = append(reasons, "warehouse mismatch")
		}
	}

	if attrs.RiskScore > req.MaxRisk {
		reasons = append(reasons, fmt.Sprintf("risk_score %.1f exceeds max %.1f", attrs.RiskScore, req.MaxRisk))
	}
	if attrs.MFALevel < req.MinMFALevel {
		reasons = append(reasons, fmt.Sprintf("mfa_level %d below required %d", attrs.MFALevel, req.MinMFALevel))
	}
	if req.RequireTrustedDevice && !attrs.TrustedDevice {
		reasons = append(reasons, "trusted device required")
	}

	return Decision{
		Allow:   len(reasons) == 0,
		Reasons: reasons,
	}
}

// EvaluateScope checks whether grant scope covers the request attributes.
func EvaluateScope(scope domain.Scope, attrs Attributes) Decision {
	if !scope.Matches(attrs.TenantID, attrs.CityID, attrs.WarehouseID) {
		return Decision{Allow: false, Reasons: []string{"scope mismatch"}}
	}
	return Decision{Allow: true}
}

// MFARequiredForRisk returns whether step-up MFA is needed given policy thresholds.
func MFARequiredForRisk(attrs Attributes, policy domain.SecurityPolicy) bool {
	if policy.MFARequired {
		return attrs.MFALevel < 1
	}
	if attrs.RiskScore >= policy.MFARequiredAboveRisk {
		return attrs.MFALevel < 2
	}
	return false
}

// RiskBlocked reports whether the risk score exceeds the hard block threshold.
func RiskBlocked(attrs Attributes, policy domain.SecurityPolicy) bool {
	return attrs.RiskScore >= policy.BlockAboveRisk
}
