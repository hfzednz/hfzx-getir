package domain

// Validation issue codes (stable wire contract).
const (
	IssueCustomerInactive     = "customer_inactive"
	IssueCustomerMissing      = "customer_missing"
	IssueAddressMissing       = "address_missing"
	IssueAddressIncomplete    = "address_incomplete"
	IssueZoneOutOfRange       = "zone_out_of_range"
	IssueZoneUnavailable      = "zone_unavailable"
	IssueInventoryInsufficient = "inventory_insufficient"
	IssueInventoryUnavailable = "inventory_unavailable"
	IssuePriceStale           = "price_stale"
	IssuePriceFailed          = "price_failed"
	IssueCouponInvalid        = "coupon_invalid"
	IssueCouponIneligible     = "coupon_ineligible"
	IssueAgeRestricted        = "age_restricted"
	IssueRegionRestricted     = "region_restricted"
	IssueFraudHighRisk        = "fraud_high_risk"
	IssuePaymentIneligible    = "payment_ineligible"
	IssueDuplicateOrder       = "duplicate_order"
	IssueMinOrderNotMet       = "min_order_not_met"
	IssueSlotRequired         = "slot_required"
	IssueSlotUnavailable      = "slot_unavailable"
)

// ValidationIssue is a single pipeline failure/warning.
type ValidationIssue struct {
	Code     string         `json:"code"`
	Message  string         `json:"message"`
	Field    string         `json:"field,omitempty"`
	Severity string         `json:"severity,omitempty"` // error | warning
	Meta     map[string]any `json:"meta,omitempty"`
}

// Blocking reports whether the issue blocks completion.
func (i ValidationIssue) Blocking() bool {
	return i.Severity == "" || i.Severity == "error"
}
