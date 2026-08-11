package postgres

import "database/sql"

type Repos struct {
	Suppliers   *SupplierRepo
	Documents   *DocumentRepo
	Certs       *CertRepo
	Contracts   *ContractRepo
	RFQs        *RFQRepo
	Quotes      *QuotationRepo
	POs         *PORepo
	Shipments   *ShipmentRepo
	Invoices    *InvoiceMatchRepo
	Sellers     *SellerRepo
	Listings    *ListingRepo
	Submissions *SubmissionRepo
	EDI         *EDIRepo
	Scorecards  *ScorecardRepo
	Messages    *MessageRepo
	Changes     *ChangeRepo
	Outbox      *OutboxRepo
}

func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Suppliers: &SupplierRepo{DB: db}, Documents: &DocumentRepo{DB: db}, Certs: &CertRepo{DB: db},
		Contracts: &ContractRepo{DB: db}, RFQs: &RFQRepo{DB: db}, Quotes: &QuotationRepo{DB: db},
		POs: &PORepo{DB: db}, Shipments: &ShipmentRepo{DB: db}, Invoices: &InvoiceMatchRepo{DB: db},
		Sellers: &SellerRepo{DB: db}, Listings: &ListingRepo{DB: db}, Submissions: &SubmissionRepo{DB: db},
		EDI: &EDIRepo{DB: db}, Scorecards: &ScorecardRepo{DB: db}, Messages: &MessageRepo{DB: db},
		Changes: &ChangeRepo{DB: db}, Outbox: &OutboxRepo{DB: db},
	}
}
