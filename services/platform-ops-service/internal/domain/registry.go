package domain

import "time"

type PlatformTenant struct {
	ID            string    `json:"id"`
	Slug          string    `json:"slug"`
	Name          string    `json:"name"`
	CompanyID     string    `json:"companyId"`
	CompanyName   string    `json:"companyName"`
	IsolationMode string    `json:"isolationMode"`
	Status        string    `json:"status"`
	Region        string    `json:"region"`
	CreatedAt     time.Time `json:"createdAt"`
	Description   string    `json:"description"`
}

type DualControlProposal struct {
	ID             string    `json:"id"`
	Action         string    `json:"action"`
	TenantID       string    `json:"tenantId"`
	TenantName     string    `json:"tenantName"`
	RequesterID    string    `json:"requesterId"`
	RequesterEmail string    `json:"requesterEmail"`
	Status         string    `json:"status"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"createdAt"`
}

type PlatformCompany struct {
	ID               string    `json:"id"`
	LegalName        string    `json:"legalName"`
	TradeName        string    `json:"tradeName"`
	CountryCode      string    `json:"countryCode"`
	Status           string    `json:"status"`
	TenantCount      int       `json:"tenantCount"`
	PrimaryCurrency  string    `json:"primaryCurrency"`
	CreatedAt        time.Time `json:"createdAt"`
	Industry         string    `json:"industry,omitempty"`
	TaxID            string    `json:"taxId,omitempty"`
	VATNumber        string    `json:"vatNumber,omitempty"`
	BillingEmail     string    `json:"billingEmail,omitempty"`
	RegisteredAddr   string    `json:"registeredAddress,omitempty"`
	DefaultLocale    string    `json:"defaultLocale,omitempty"`
	TimeZone         string    `json:"timeZone,omitempty"`
	PrimaryColor     string    `json:"primaryColor,omitempty"`
}

type PlatformAuditEntry struct {
	ID           string    `json:"id"`
	ActorID      string    `json:"actorId"`
	ActorEmail   string    `json:"actorEmail"`
	Action       string    `json:"action"`
	Resource     string    `json:"resource"`
	ResourceID   string    `json:"resourceId"`
	When         time.Time `json:"when"`
	Where        string    `json:"where"`
	Device       string    `json:"device"`
	IP           string    `json:"ip"`
	SessionID    string    `json:"sessionId"`
	OldValue     *string   `json:"oldValue"`
	NewValue     *string   `json:"newValue"`
	Severity     string    `json:"severity"`
	Sealed       bool      `json:"sealed"`
}

type PlatformPerson struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Kind        string `json:"kind"`
	OrgUnitID   string `json:"orgUnitId"`
	OrgUnitName string `json:"orgUnitName"`
	Status      string `json:"status"`
}
