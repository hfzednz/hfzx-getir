package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/supplier-service/internal/app"
	"github.com/nexora/supplier-service/internal/app/memory"
	"github.com/nexora/supplier-service/internal/domain"
)

func testDeps() *app.Deps {
	s := memory.NewStore()
	r := memory.NewRepos(s)
	return &app.Deps{
		Suppliers: r.Suppliers, Documents: r.Documents, Certs: r.Certs, Contracts: r.Contracts,
		RFQs: r.RFQs, Quotes: r.Quotes, POs: r.POs, Shipments: r.Shipments, Invoices: r.Invoices,
		Sellers: r.Sellers, Listings: r.Listings, Submissions: r.Submissions, EDI: r.EDI,
		Scorecards: r.Scorecards, Messages: r.Messages, Changes: r.Changes, Outbox: r.Outbox,
		ERP: r.ERP, Catalog: r.Catalog, Inventory: r.Inventory, Settlement: r.Settlement,
		AI: r.AI, Metrics: r.Metrics, Clock: app.SystemClock{}, IDs: app.UUIDGen{},
	}
}

func TestSupplierEcosystemFlows(t *testing.T) {
	ctx := context.Background()
	d := testDeps()
	tid := uuid.New()
	co := uuid.New()

	s, err := d.OnboardSupplier(ctx, domain.Supplier{
		TenantID: tid, CompanyID: co, Code: "SUP-1", LegalName: "Acme Foods AS", Country: "TR",
		PartnerKinds: []domain.PartnerKind{domain.PartnerManufacturer, domain.PartnerWholesaler},
	})
	if err != nil {
		t.Fatal(err)
	}
	s, err = d.ApproveSupplier(ctx, tid, s.ID)
	if err != nil || s.Status != domain.SupplierApproved || s.ErpSupplierID == "" {
		t.Fatalf("%+v %v", s, err)
	}

	_, err = d.AddDocument(ctx, domain.SupplierDocument{
		TenantID: tid, SupplierID: s.ID, Kind: domain.DocTradeLicense, Name: "license", URI: "enc://doc/1",
	})
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	ct, err := d.UpsertContract(ctx, domain.Contract{
		TenantID: tid, SupplierID: s.ID, Title: "MSA 2026", StartsAt: now, EndsAt: now.AddDate(1, 0, 0), TermsURI: "enc://c/1",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.ActivateContract(ctx, tid, ct.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.RenewContract(ctx, tid, ct.ID, now.AddDate(2, 0, 0))
	if err != nil {
		t.Fatal(err)
	}

	rfq, err := d.CreateRFQ(ctx, domain.RFQ{
		TenantID: tid, CompanyID: co, Title: "Milk RFQ",
		Lines: []domain.RFQLine{{SKU: "MILK-1L", Qty: 100, UOM: "EA"}},
		DueAt: now.Add(48 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	q, err := d.SubmitQuotation(ctx, domain.Quotation{
		TenantID: tid, RFQID: rfq.ID, SupplierID: s.ID, Currency: "TRY", TotalMinor: 500000, LeadTimeDays: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	po, err := d.AwardQuotation(ctx, tid, q.ID)
	if err != nil || po.ErpPOID == "" || po.TotalMinor <= 0 {
		t.Fatalf("%+v %v", po, err)
	}

	ship, err := d.AnnounceShipment(ctx, domain.InboundShipment{
		TenantID: tid, SupplierID: s.ID, POID: po.ID, WarehouseID: "WH-1", TrackingRef: "TRK-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	ship, err = d.ReceiveShipment(ctx, tid, ship.ID, true)
	if err != nil || ship.Status != domain.ShipReceived {
		t.Fatal(err)
	}

	m, err := d.SignalInvoiceMatch(ctx, domain.InvoiceMatchSignal{
		TenantID: tid, POID: po.ID, InvoiceRef: "INV-1", AmountMinor: po.TotalMinor, Currency: "TRY",
	})
	if err != nil || !m.Matched {
		t.Fatalf("%+v %v", m, err)
	}

	seller, err := d.OnboardSeller(ctx, domain.MarketplaceSeller{
		TenantID: tid, SupplierID: s.ID, StoreName: "Acme Market",
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = d.UpsertListing(ctx, domain.ListingRef{
		TenantID: tid, SellerID: seller.ID, ExternalSKU: "EXT-1", PriceMinor: 1990, Currency: "TRY", StockHint: 50,
	})
	if err != nil {
		t.Fatal(err)
	}

	sub, err := d.SubmitCatalog(ctx, domain.CatalogSubmission{
		TenantID: tid, SupplierID: s.ID, SKU: "NEW-SKU", Title: "Yogurt", Attributes: map[string]any{"fat": "3%"},
	})
	if err != nil {
		t.Fatal(err)
	}
	sub, err = d.DecideCatalogSubmission(ctx, tid, sub.ID, true)
	if err != nil || sub.Status != domain.SubApproved {
		t.Fatal(err)
	}

	_, err = d.IngestEDI(ctx, domain.EDIDocument{
		TenantID: tid, SupplierID: s.ID, DocType: domain.EDIOrder, Payload: "ISA*...*850*",
	})
	if err != nil {
		t.Fatal(err)
	}

	sc, err := d.RateSupplier(ctx, domain.Scorecard{
		TenantID: tid, SupplierID: s.ID, Period: "2026-08",
		DeliveryScore: 0.95, QualityScore: 0.9, PriceScore: 0.85, ComplianceScore: 0.9, RiskScore: 0.2, FillRate: 0.98,
	})
	if err != nil || sc.Overall <= 0 {
		t.Fatal(err)
	}

	st, err := d.AdminStats(ctx, tid)
	if err != nil || st["approved"].(int) < 1 {
		t.Fatalf("%v", st)
	}
}
