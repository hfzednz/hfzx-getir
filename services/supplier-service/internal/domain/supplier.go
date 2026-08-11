package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Money is int64 minor units (constitution).
type Money struct {
	Currency   string `json:"currency"`
	AmountMinor int64  `json:"amountMinor"`
}

type SupplierStatus string

const (
	SupplierDraft      SupplierStatus = "draft"
	SupplierPending    SupplierStatus = "pending_verification"
	SupplierApproved   SupplierStatus = "approved"
	SupplierSuspended  SupplierStatus = "suspended"
	SupplierOffboarded SupplierStatus = "offboarded"
)

type PartnerKind string

const (
	PartnerLogistics    PartnerKind = "logistics"
	PartnerManufacturer PartnerKind = "manufacturer"
	PartnerWholesaler   PartnerKind = "wholesaler"
	PartnerDistributor  PartnerKind = "distributor"
	PartnerRetail       PartnerKind = "retail"
	PartnerTechnology   PartnerKind = "technology"
	PartnerService      PartnerKind = "service"
	PartnerAffiliate    PartnerKind = "affiliate"
	PartnerSupplier     PartnerKind = "supplier"
)

// Supplier is the ecosystem master (richer than ERP supplier stub; ERP refs via ErpSupplierID).
type Supplier struct {
	ID              uuid.UUID      `json:"id"`
	TenantID        uuid.UUID      `json:"tenantId"`
	CompanyID       uuid.UUID      `json:"companyId"`
	Code            string         `json:"code"`
	LegalName       string         `json:"legalName"`
	DisplayName     string         `json:"displayName"`
	Country         string         `json:"country"`
	TaxID           string         `json:"taxId"`
	Status          SupplierStatus `json:"status"`
	PartnerKinds    []PartnerKind  `json:"partnerKinds"`
	Contacts        []Contact      `json:"contacts"`
	BankingRef      string         `json:"bankingRef"` // opaque vault/token ref — not PAN/IBAN plaintext SoT
	ErpSupplierID   string         `json:"erpSupplierId,omitempty"`
	RiskScore       float64        `json:"riskScore"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
	ApprovedAt      *time.Time     `json:"approvedAt,omitempty"`
}

type Contact struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
	Role  string `json:"role"`
}

type DocumentKind string

const (
	DocTradeLicense DocumentKind = "trade_license"
	DocTaxCert      DocumentKind = "tax_certificate"
	DocInsurance    DocumentKind = "insurance"
	DocOther        DocumentKind = "other"
)

type SupplierDocument struct {
	ID         uuid.UUID    `json:"id"`
	TenantID   uuid.UUID    `json:"tenantId"`
	SupplierID uuid.UUID    `json:"supplierId"`
	Kind       DocumentKind `json:"kind"`
	Name       string       `json:"name"`
	URI        string       `json:"uri"` // encrypted object ref
	ExpiresAt  *time.Time   `json:"expiresAt,omitempty"`
	Verified   bool         `json:"verified"`
	CreatedAt  time.Time    `json:"createdAt"`
}

type Certification struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	SupplierID uuid.UUID  `json:"supplierId"`
	Name       string     `json:"name"`
	Issuer     string     `json:"issuer"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type ContractStatus string

const (
	ContractDraft    ContractStatus = "draft"
	ContractPending  ContractStatus = "pending_approval"
	ContractActive   ContractStatus = "active"
	ContractExpired  ContractStatus = "expired"
	ContractRenewed  ContractStatus = "renewed"
)

type Contract struct {
	ID         uuid.UUID      `json:"id"`
	TenantID   uuid.UUID      `json:"tenantId"`
	SupplierID uuid.UUID      `json:"supplierId"`
	Title      string         `json:"title"`
	Version    int            `json:"version"`
	Status     ContractStatus `json:"status"`
	StartsAt   time.Time      `json:"startsAt"`
	EndsAt     time.Time      `json:"endsAt"`
	TermsURI   string         `json:"termsUri"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
}

// --- Procurement collaboration (not ERP AP 3-way SoT) ---

type RFQStatus string

const (
	RFQDraft    RFQStatus = "draft"
	RFQOpen     RFQStatus = "open"
	RFQClosed   RFQStatus = "closed"
	RFQAwarded  RFQStatus = "awarded"
)

type RFQ struct {
	ID          uuid.UUID `json:"id"`
	TenantID    uuid.UUID `json:"tenantId"`
	CompanyID   uuid.UUID `json:"companyId"`
	Number      string    `json:"number"`
	Title       string    `json:"title"`
	Status      RFQStatus `json:"status"`
	Lines       []RFQLine `json:"lines"`
	DueAt       time.Time `json:"dueAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type RFQLine struct {
	SKU         string `json:"sku"`
	Description string `json:"description"`
	Qty         int64  `json:"qty"`
	UOM         string `json:"uom"`
}

type Quotation struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	RFQID        uuid.UUID `json:"rfqId"`
	SupplierID   uuid.UUID `json:"supplierId"`
	Currency     string    `json:"currency"`
	TotalMinor   int64     `json:"totalMinor"`
	LeadTimeDays int       `json:"leadTimeDays"`
	Status       string    `json:"status"` // submitted|selected|rejected
	CreatedAt    time.Time `json:"createdAt"`
}

type SourcingPOStatus string

const (
	PODraft     SourcingPOStatus = "draft"
	POSent      SourcingPOStatus = "sent"
	POConfirmed SourcingPOStatus = "confirmed"
	POPartial   SourcingPOStatus = "partial"
	POReceived  SourcingPOStatus = "received"
	POClosed    SourcingPOStatus = "closed"
	POCancelled SourcingPOStatus = "cancelled"
)

// SourcingPurchaseOrder is supplier-ecosystem PO; may reference ERP PO via ErpPOID.
type SourcingPurchaseOrder struct {
	ID           uuid.UUID        `json:"id"`
	TenantID     uuid.UUID        `json:"tenantId"`
	CompanyID    uuid.UUID        `json:"companyId"`
	SupplierID   uuid.UUID        `json:"supplierId"`
	Number       string           `json:"number"`
	Status       SourcingPOStatus `json:"status"`
	Currency     string           `json:"currency"`
	TotalMinor   int64            `json:"totalMinor"`
	Lines        []POLine         `json:"lines"`
	QuotationID  *uuid.UUID       `json:"quotationId,omitempty"`
	ErpPOID      string           `json:"erpPoId,omitempty"`
	CreatedAt    time.Time        `json:"createdAt"`
	UpdatedAt    time.Time        `json:"updatedAt"`
}

type POLine struct {
	SKU       string `json:"sku"`
	Qty       int64  `json:"qty"`
	UnitMinor int64  `json:"unitMinor"`
}

type ShipmentStatus string

const (
	ShipAnnounced ShipmentStatus = "announced"
	ShipInTransit ShipmentStatus = "in_transit"
	ShipReceived  ShipmentStatus = "received"
	ShipException ShipmentStatus = "exception"
)

type InboundShipment struct {
	ID            uuid.UUID      `json:"id"`
	TenantID      uuid.UUID      `json:"tenantId"`
	SupplierID    uuid.UUID      `json:"supplierId"`
	POID          uuid.UUID      `json:"poId"`
	ASNNumber     string         `json:"asnNumber"`
	Status        ShipmentStatus `json:"status"`
	TrackingRef   string         `json:"trackingRef"`
	WarehouseID   string         `json:"warehouseId"`
	QCPassed      *bool          `json:"qcPassed,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	ReceivedAt    *time.Time     `json:"receivedAt,omitempty"`
}

type InvoiceMatchSignal struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	SupplierID   uuid.UUID `json:"supplierId"`
	POID         uuid.UUID `json:"poId"`
	InvoiceRef   string    `json:"invoiceRef"`
	AmountMinor  int64     `json:"amountMinor"`
	Currency     string    `json:"currency"`
	Matched      bool      `json:"matched"`
	MatchScore   float64   `json:"matchScore"`
	ErpInvoiceID string    `json:"erpInvoiceId,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

// --- Marketplace ---

type SellerStatus string

const (
	SellerPending  SellerStatus = "pending"
	SellerActive   SellerStatus = "active"
	SellerSuspended SellerStatus = "suspended"
)

type MarketplaceSeller struct {
	ID         uuid.UUID    `json:"id"`
	TenantID   uuid.UUID    `json:"tenantId"`
	SupplierID uuid.UUID    `json:"supplierId"`
	StoreName  string       `json:"storeName"`
	Status     SellerStatus `json:"status"`
	RatingAvg  float64      `json:"ratingAvg"`
	RatingCount int         `json:"ratingCount"`
	CreatedAt  time.Time    `json:"createdAt"`
}

type ListingRef struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenantId"`
	SellerID     uuid.UUID `json:"sellerId"`
	ExternalSKU  string    `json:"externalSku"`
	CatalogSKU   string    `json:"catalogSku,omitempty"`
	PriceMinor   int64     `json:"priceMinor"`
	Currency     string    `json:"currency"`
	StockHint    int64     `json:"stockHint"` // advisory; inventory SoT elsewhere
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
}

// --- Catalog collaboration ---

type SubmissionStatus string

const (
	SubPending  SubmissionStatus = "pending"
	SubApproved SubmissionStatus = "approved"
	SubRejected SubmissionStatus = "rejected"
)

type CatalogSubmission struct {
	ID         uuid.UUID        `json:"id"`
	TenantID   uuid.UUID        `json:"tenantId"`
	SupplierID uuid.UUID        `json:"supplierId"`
	SKU        string           `json:"sku"`
	Title      string           `json:"title"`
	Attributes map[string]any   `json:"attributes"`
	MediaURIs  []string         `json:"mediaUris"`
	Version    int              `json:"version"`
	Status     SubmissionStatus `json:"status"`
	CreatedAt  time.Time        `json:"createdAt"`
}

// --- EDI ---

type EDIDocType string

const (
	EDIOrder      EDIDocType = "850"
	EDIAck        EDIDocType = "855"
	EDIASN        EDIDocType = "856"
	EDIInvoice    EDIDocType = "810"
	EDIInventory  EDIDocType = "846"
)

type EDIDocument struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	SupplierID uuid.UUID  `json:"supplierId"`
	DocType    EDIDocType `json:"docType"`
	Direction  string     `json:"direction"` // in|out
	Payload    string     `json:"payload"`
	MappedRef  string     `json:"mappedRef,omitempty"`
	Status     string     `json:"status"` // received|mapped|failed
	CreatedAt  time.Time  `json:"createdAt"`
}

// --- Performance & collab ---

type Scorecard struct {
	ID               uuid.UUID `json:"id"`
	TenantID         uuid.UUID `json:"tenantId"`
	SupplierID       uuid.UUID `json:"supplierId"`
	Period           string    `json:"period"` // YYYY-MM
	DeliveryScore    float64   `json:"deliveryScore"`
	QualityScore     float64   `json:"qualityScore"`
	PriceScore       float64   `json:"priceScore"`
	LeadTimeDaysAvg  float64   `json:"leadTimeDaysAvg"`
	FillRate         float64   `json:"fillRate"`
	ComplianceScore  float64   `json:"complianceScore"`
	RiskScore        float64   `json:"riskScore"`
	Overall          float64   `json:"overall"`
	CreatedAt        time.Time `json:"createdAt"`
}

type MessageThread struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenantId"`
	SupplierID uuid.UUID `json:"supplierId"`
	Subject    string    `json:"subject"`
	CreatedAt  time.Time `json:"createdAt"`
}

type Message struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenantId"`
	ThreadID  uuid.UUID `json:"threadId"`
	Sender    string    `json:"sender"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}

type ChangeRequest struct {
	ID         uuid.UUID  `json:"id"`
	TenantID   uuid.UUID  `json:"tenantId"`
	Kind       string     `json:"kind"`
	SubjectKey string     `json:"subjectKey"`
	Payload    map[string]any `json:"payload"`
	Status     string     `json:"status"` // pending|approved|rejected
	CreatedAt  time.Time  `json:"createdAt"`
	DecidedAt  *time.Time `json:"decidedAt,omitempty"`
}

func ValidateSupplier(s Supplier) error {
	if s.TenantID == uuid.Nil || s.CompanyID == uuid.Nil {
		return ErrInvalidArgument
	}
	if strings.TrimSpace(s.Code) == "" || strings.TrimSpace(s.LegalName) == "" {
		return ErrInvalidArgument
	}
	if s.Country == "" {
		return ErrInvalidArgument
	}
	return nil
}

func CanApprove(s Supplier) error {
	if s.Status != SupplierPending && s.Status != SupplierDraft {
		return ErrIllegalTransition
	}
	return nil
}

func ComputeOverallScore(sc Scorecard) float64 {
	return (sc.DeliveryScore + sc.QualityScore + sc.PriceScore + sc.ComplianceScore + (1 - sc.RiskScore)) / 5
}

func MatchInvoiceHint(poTotal, invTotal int64) (bool, float64) {
	if poTotal <= 0 {
		return false, 0
	}
	diff := poTotal - invTotal
	if diff < 0 {
		diff = -diff
	}
	score := 1 - float64(diff)/float64(poTotal)
	if score < 0 {
		score = 0
	}
	return score >= 0.98, score
}
