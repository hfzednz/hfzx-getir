package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/supplier-service/internal/app/ports"
	"github.com/nexora/supplier-service/internal/domain"
)

func (d *Deps) OnboardSupplier(ctx context.Context, s domain.Supplier) (domain.Supplier, error) {
	if err := domain.ValidateSupplier(s); err != nil {
		return domain.Supplier{}, err
	}
	now := d.now()
	if s.ID == uuid.Nil {
		s.ID = d.newID()
	}
	if s.Status == "" {
		s.Status = domain.SupplierPending
	}
	if len(s.PartnerKinds) == 0 {
		s.PartnerKinds = []domain.PartnerKind{domain.PartnerSupplier}
	}
	s.CreatedAt = now
	s.UpdatedAt = now
	if d.AI != nil {
		if risk, err := d.AI.PredictRisk(ctx, s); err == nil {
			s.RiskScore = risk
		}
	}
	if err := d.Suppliers.Save(ctx, s); err != nil {
		return domain.Supplier{}, err
	}
	d.emit(ctx, s.TenantID, s.ID, domain.EventSupplierCreated, map[string]any{"code": s.Code})
	if d.Metrics != nil {
		_ = d.Metrics.Record(ctx, "supplier.onboarded", map[string]string{"country": s.Country}, 1)
	}
	return s, nil
}

func (d *Deps) ApproveSupplier(ctx context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	s, err := d.Suppliers.Get(ctx, tenantID, id)
	if err != nil {
		return domain.Supplier{}, err
	}
	if err := domain.CanApprove(s); err != nil {
		return domain.Supplier{}, err
	}
	now := d.now()
	s.Status = domain.SupplierApproved
	s.ApprovedAt = &now
	s.UpdatedAt = now
	if d.ERP != nil {
		if erpID, err := d.ERP.SyncSupplier(ctx, tenantID, s); err == nil && erpID != "" {
			s.ErpSupplierID = erpID
		}
	}
	if err := d.Suppliers.Save(ctx, s); err != nil {
		return domain.Supplier{}, err
	}
	d.emit(ctx, tenantID, s.ID, domain.EventSupplierApproved, map[string]any{"code": s.Code})
	return s, nil
}

func (d *Deps) ListSuppliers(ctx context.Context, tenantID uuid.UUID) ([]domain.Supplier, error) {
	return d.Suppliers.List(ctx, tenantID)
}

func (d *Deps) AddDocument(ctx context.Context, doc domain.SupplierDocument) (domain.SupplierDocument, error) {
	if doc.TenantID == uuid.Nil || doc.SupplierID == uuid.Nil || strings.TrimSpace(doc.URI) == "" {
		return domain.SupplierDocument{}, domain.ErrInvalidArgument
	}
	if _, err := d.Suppliers.Get(ctx, doc.TenantID, doc.SupplierID); err != nil {
		return domain.SupplierDocument{}, err
	}
	if doc.ID == uuid.Nil {
		doc.ID = d.newID()
	}
	doc.CreatedAt = d.now()
	if err := d.Documents.Save(ctx, doc); err != nil {
		return domain.SupplierDocument{}, err
	}
	return doc, nil
}

func (d *Deps) AddCertification(ctx context.Context, c domain.Certification) (domain.Certification, error) {
	if c.TenantID == uuid.Nil || c.SupplierID == uuid.Nil || c.Name == "" {
		return domain.Certification{}, domain.ErrInvalidArgument
	}
	if c.ID == uuid.Nil {
		c.ID = d.newID()
	}
	c.CreatedAt = d.now()
	if err := d.Certs.Save(ctx, c); err != nil {
		return domain.Certification{}, err
	}
	return c, nil
}

func (d *Deps) UpsertContract(ctx context.Context, c domain.Contract) (domain.Contract, error) {
	if c.TenantID == uuid.Nil || c.SupplierID == uuid.Nil || c.Title == "" {
		return domain.Contract{}, domain.ErrInvalidArgument
	}
	now := d.now()
	if c.ID == uuid.Nil {
		c.ID = d.newID()
		c.Version = 1
		c.CreatedAt = now
		if c.Status == "" {
			c.Status = domain.ContractDraft
		}
	} else {
		prev, err := d.Contracts.Get(ctx, c.TenantID, c.ID)
		if err == nil {
			c.Version = prev.Version + 1
			c.CreatedAt = prev.CreatedAt
		}
	}
	c.UpdatedAt = now
	if err := d.Contracts.Save(ctx, c); err != nil {
		return domain.Contract{}, err
	}
	return c, nil
}

func (d *Deps) ActivateContract(ctx context.Context, tenantID, id uuid.UUID) (domain.Contract, error) {
	c, err := d.Contracts.Get(ctx, tenantID, id)
	if err != nil {
		return domain.Contract{}, err
	}
	c.Status = domain.ContractActive
	c.UpdatedAt = d.now()
	if err := d.Contracts.Save(ctx, c); err != nil {
		return domain.Contract{}, err
	}
	return c, nil
}

func (d *Deps) RenewContract(ctx context.Context, tenantID, id uuid.UUID, endsAt time.Time) (domain.Contract, error) {
	c, err := d.Contracts.Get(ctx, tenantID, id)
	if err != nil {
		return domain.Contract{}, err
	}
	c.Status = domain.ContractRenewed
	c.EndsAt = endsAt
	c.Version++
	c.UpdatedAt = d.now()
	if err := d.Contracts.Save(ctx, c); err != nil {
		return domain.Contract{}, err
	}
	d.emit(ctx, tenantID, c.ID, domain.EventContractRenewed, map[string]any{"supplierId": c.SupplierID.String()})
	return c, nil
}

func (d *Deps) CreateRFQ(ctx context.Context, r domain.RFQ) (domain.RFQ, error) {
	if r.TenantID == uuid.Nil || r.CompanyID == uuid.Nil || r.Title == "" || len(r.Lines) == 0 {
		return domain.RFQ{}, domain.ErrInvalidArgument
	}
	r.ID = d.newID()
	r.Number = fmt.Sprintf("RFQ-%s", r.ID.String()[:8])
	r.Status = domain.RFQOpen
	r.CreatedAt = d.now()
	if err := d.RFQs.Save(ctx, r); err != nil {
		return domain.RFQ{}, err
	}
	return r, nil
}

func (d *Deps) SubmitQuotation(ctx context.Context, q domain.Quotation) (domain.Quotation, error) {
	if q.TenantID == uuid.Nil || q.RFQID == uuid.Nil || q.SupplierID == uuid.Nil {
		return domain.Quotation{}, domain.ErrInvalidArgument
	}
	s, err := d.Suppliers.Get(ctx, q.TenantID, q.SupplierID)
	if err != nil {
		return domain.Quotation{}, err
	}
	if s.Status != domain.SupplierApproved {
		return domain.Quotation{}, domain.ErrNotVerified
	}
	rfq, err := d.RFQs.Get(ctx, q.TenantID, q.RFQID)
	if err != nil {
		return domain.Quotation{}, err
	}
	if rfq.Status != domain.RFQOpen {
		return domain.Quotation{}, domain.ErrIllegalTransition
	}
	q.ID = d.newID()
	q.Status = "submitted"
	q.CreatedAt = d.now()
	if q.Currency == "" {
		q.Currency = "TRY"
	}
	if err := d.Quotes.Save(ctx, q); err != nil {
		return domain.Quotation{}, err
	}
	return q, nil
}

func (d *Deps) AwardQuotation(ctx context.Context, tenantID, quoteID uuid.UUID) (domain.SourcingPurchaseOrder, error) {
	q, err := d.Quotes.Get(ctx, tenantID, quoteID)
	if err != nil {
		return domain.SourcingPurchaseOrder{}, err
	}
	rfq, err := d.RFQs.Get(ctx, tenantID, q.RFQID)
	if err != nil {
		return domain.SourcingPurchaseOrder{}, err
	}
	q.Status = "selected"
	_ = d.Quotes.Save(ctx, q)
	rfq.Status = domain.RFQAwarded
	_ = d.RFQs.Save(ctx, rfq)

	lines := make([]domain.POLine, 0, len(rfq.Lines))
	var total int64
	unit := int64(0)
	if len(rfq.Lines) > 0 && q.TotalMinor > 0 {
		unit = q.TotalMinor / rfq.Lines[0].Qty
		if unit <= 0 {
			unit = q.TotalMinor
		}
	}
	for _, ln := range rfq.Lines {
		lines = append(lines, domain.POLine{SKU: ln.SKU, Qty: ln.Qty, UnitMinor: unit})
		total += unit * ln.Qty
	}
	if total == 0 {
		total = q.TotalMinor
	}
	qid := q.ID
	po := domain.SourcingPurchaseOrder{
		ID: d.newID(), TenantID: tenantID, CompanyID: rfq.CompanyID, SupplierID: q.SupplierID,
		Number: fmt.Sprintf("PO-%s", d.newID().String()[:8]), Status: domain.POSent,
		Currency: q.Currency, TotalMinor: total, Lines: lines, QuotationID: &qid, CreatedAt: d.now(), UpdatedAt: d.now(),
	}
	if d.ERP != nil {
		if erpID, err := d.ERP.EnsurePO(ctx, tenantID, po); err == nil {
			po.ErpPOID = erpID
		}
	}
	if err := d.POs.Save(ctx, po); err != nil {
		return domain.SourcingPurchaseOrder{}, err
	}
	d.emit(ctx, tenantID, po.ID, domain.EventPurchaseOrderCreated, map[string]any{
		"number": po.Number, "supplierId": po.SupplierID.String(), "totalMinor": po.TotalMinor,
	})
	return po, nil
}

func (d *Deps) CreatePO(ctx context.Context, po domain.SourcingPurchaseOrder) (domain.SourcingPurchaseOrder, error) {
	if po.TenantID == uuid.Nil || po.SupplierID == uuid.Nil || po.CompanyID == uuid.Nil || len(po.Lines) == 0 {
		return domain.SourcingPurchaseOrder{}, domain.ErrInvalidArgument
	}
	s, err := d.Suppliers.Get(ctx, po.TenantID, po.SupplierID)
	if err != nil {
		return domain.SourcingPurchaseOrder{}, err
	}
	if s.Status != domain.SupplierApproved {
		return domain.SourcingPurchaseOrder{}, domain.ErrNotVerified
	}
	po.ID = d.newID()
	po.Number = fmt.Sprintf("PO-%s", po.ID.String()[:8])
	po.Status = domain.POSent
	var total int64
	for _, ln := range po.Lines {
		total += ln.UnitMinor * ln.Qty
	}
	po.TotalMinor = total
	if po.Currency == "" {
		po.Currency = "TRY"
	}
	now := d.now()
	po.CreatedAt = now
	po.UpdatedAt = now
	if d.ERP != nil {
		if erpID, err := d.ERP.EnsurePO(ctx, po.TenantID, po); err == nil {
			po.ErpPOID = erpID
		}
	}
	if err := d.POs.Save(ctx, po); err != nil {
		return domain.SourcingPurchaseOrder{}, err
	}
	d.emit(ctx, po.TenantID, po.ID, domain.EventPurchaseOrderCreated, map[string]any{"number": po.Number})
	return po, nil
}

func (d *Deps) AnnounceShipment(ctx context.Context, ship domain.InboundShipment) (domain.InboundShipment, error) {
	if ship.TenantID == uuid.Nil || ship.POID == uuid.Nil || ship.SupplierID == uuid.Nil {
		return domain.InboundShipment{}, domain.ErrInvalidArgument
	}
	if _, err := d.POs.Get(ctx, ship.TenantID, ship.POID); err != nil {
		return domain.InboundShipment{}, err
	}
	ship.ID = d.newID()
	if ship.ASNNumber == "" {
		ship.ASNNumber = fmt.Sprintf("ASN-%s", ship.ID.String()[:8])
	}
	ship.Status = domain.ShipAnnounced
	ship.CreatedAt = d.now()
	if err := d.Shipments.Save(ctx, ship); err != nil {
		return domain.InboundShipment{}, err
	}
	if d.Inventory != nil {
		_ = d.Inventory.AnnounceASN(ctx, ship.TenantID, ship)
	}
	return ship, nil
}

func (d *Deps) ReceiveShipment(ctx context.Context, tenantID, id uuid.UUID, qcPassed bool) (domain.InboundShipment, error) {
	ship, err := d.Shipments.Get(ctx, tenantID, id)
	if err != nil {
		return domain.InboundShipment{}, err
	}
	now := d.now()
	ship.Status = domain.ShipReceived
	ship.QCPassed = &qcPassed
	ship.ReceivedAt = &now
	if err := d.Shipments.Save(ctx, ship); err != nil {
		return domain.InboundShipment{}, err
	}
	po, err := d.POs.Get(ctx, tenantID, ship.POID)
	if err == nil {
		po.Status = domain.POReceived
		po.UpdatedAt = now
		_ = d.POs.Save(ctx, po)
		if qcPassed && d.Inventory != nil {
			wh := warehouseID(tenantID, ship.WarehouseID)
			for i, line := range po.Lines {
				if strings.TrimSpace(line.SKU) == "" || line.Qty <= 0 {
					continue
				}
				if err := d.Inventory.ReceiveStock(ctx, ports.ReceiveStockRequest{
					TenantID:       tenantID,
					WarehouseID:    wh,
					SKUCode:        line.SKU,
					Qty:            line.Qty,
					IdempotencyKey: fmt.Sprintf("po-recv:%s:%s:%d", ship.ID.String(), line.SKU, i),
					Reason:         "supplier_receive",
				}); err != nil {
					return domain.InboundShipment{}, err
				}
			}
		}
	}
	d.emit(ctx, tenantID, ship.ID, domain.EventShipmentReceived, map[string]any{
		"asn": ship.ASNNumber, "qcPassed": qcPassed,
	})
	return ship, nil
}

func (d *Deps) SignalInvoiceMatch(ctx context.Context, m domain.InvoiceMatchSignal) (domain.InvoiceMatchSignal, error) {
	if m.TenantID == uuid.Nil || m.POID == uuid.Nil {
		return domain.InvoiceMatchSignal{}, domain.ErrInvalidArgument
	}
	po, err := d.POs.Get(ctx, m.TenantID, m.POID)
	if err != nil {
		return domain.InvoiceMatchSignal{}, err
	}
	matched, score := domain.MatchInvoiceHint(po.TotalMinor, m.AmountMinor)
	m.ID = d.newID()
	m.Matched = matched
	m.MatchScore = score
	m.SupplierID = po.SupplierID
	if m.Currency == "" {
		m.Currency = po.Currency
	}
	m.CreatedAt = d.now()
	if err := d.Invoices.Save(ctx, m); err != nil {
		return domain.InvoiceMatchSignal{}, err
	}
	d.emit(ctx, m.TenantID, m.ID, domain.EventInvoiceMatched, map[string]any{
		"matched": matched, "score": score, "poId": po.ID.String(),
	})
	return m, nil
}

func (d *Deps) OnboardSeller(ctx context.Context, seller domain.MarketplaceSeller) (domain.MarketplaceSeller, error) {
	if seller.TenantID == uuid.Nil || seller.SupplierID == uuid.Nil || seller.StoreName == "" {
		return domain.MarketplaceSeller{}, domain.ErrInvalidArgument
	}
	s, err := d.Suppliers.Get(ctx, seller.TenantID, seller.SupplierID)
	if err != nil {
		return domain.MarketplaceSeller{}, err
	}
	if s.Status != domain.SupplierApproved {
		return domain.MarketplaceSeller{}, domain.ErrNotVerified
	}
	seller.ID = d.newID()
	seller.Status = domain.SellerActive
	seller.CreatedAt = d.now()
	if err := d.Sellers.Save(ctx, seller); err != nil {
		return domain.MarketplaceSeller{}, err
	}
	d.emit(ctx, seller.TenantID, seller.ID, domain.EventSellerOnboarded, map[string]any{"storeName": seller.StoreName})
	return seller, nil
}

func (d *Deps) ListSellers(ctx context.Context, tenantID uuid.UUID) ([]domain.MarketplaceSeller, error) {
	if tenantID == uuid.Nil || d.Sellers == nil {
		return nil, domain.ErrInvalidArgument
	}
	return d.Sellers.List(ctx, tenantID)
}

func (d *Deps) ListListings(ctx context.Context, tenantID, sellerID uuid.UUID) ([]domain.ListingRef, error) {
	if tenantID == uuid.Nil || d.Listings == nil || d.Sellers == nil {
		return nil, domain.ErrInvalidArgument
	}
	if sellerID != uuid.Nil {
		if _, err := d.Sellers.Get(ctx, tenantID, sellerID); err != nil {
			return nil, err
		}
		return d.Listings.ListBySeller(ctx, tenantID, sellerID)
	}
	sellers, err := d.Sellers.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]domain.ListingRef, 0)
	for _, s := range sellers {
		items, err := d.Listings.ListBySeller(ctx, tenantID, s.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func warehouseID(tenantID uuid.UUID, raw string) uuid.UUID {
	raw = strings.TrimSpace(raw)
	if id, err := uuid.Parse(raw); err == nil {
		return id
	}
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(tenantID.String()+"|"+raw))
}

func (d *Deps) UpsertListing(ctx context.Context, l domain.ListingRef) (domain.ListingRef, error) {
	if l.TenantID == uuid.Nil || l.SellerID == uuid.Nil || l.ExternalSKU == "" {
		return domain.ListingRef{}, domain.ErrInvalidArgument
	}
	if _, err := d.Sellers.Get(ctx, l.TenantID, l.SellerID); err != nil {
		return domain.ListingRef{}, err
	}
	if l.ID == uuid.Nil {
		l.ID = d.newID()
		l.CreatedAt = d.now()
	}
	l.Active = true
	if l.Currency == "" {
		l.Currency = "TRY"
	}
	if err := d.Listings.Save(ctx, l); err != nil {
		return domain.ListingRef{}, err
	}
	return l, nil
}

func (d *Deps) SubmitCatalog(ctx context.Context, sub domain.CatalogSubmission) (domain.CatalogSubmission, error) {
	if sub.TenantID == uuid.Nil || sub.SupplierID == uuid.Nil || sub.SKU == "" || sub.Title == "" {
		return domain.CatalogSubmission{}, domain.ErrInvalidArgument
	}
	sub.ID = d.newID()
	sub.Version = 1
	sub.Status = domain.SubPending
	sub.CreatedAt = d.now()
	if sub.Attributes == nil {
		sub.Attributes = map[string]any{}
	}
	if err := d.Submissions.Save(ctx, sub); err != nil {
		return domain.CatalogSubmission{}, err
	}
	return sub, nil
}

func (d *Deps) DecideCatalogSubmission(ctx context.Context, tenantID, id uuid.UUID, approve bool) (domain.CatalogSubmission, error) {
	sub, err := d.Submissions.Get(ctx, tenantID, id)
	if err != nil {
		return domain.CatalogSubmission{}, err
	}
	if sub.Status != domain.SubPending {
		return domain.CatalogSubmission{}, domain.ErrIllegalTransition
	}
	if approve {
		sub.Status = domain.SubApproved
		if d.Catalog != nil {
			_ = d.Catalog.SubmitProduct(ctx, tenantID, sub)
		}
	} else {
		sub.Status = domain.SubRejected
	}
	if err := d.Submissions.Save(ctx, sub); err != nil {
		return domain.CatalogSubmission{}, err
	}
	return sub, nil
}

func (d *Deps) IngestEDI(ctx context.Context, doc domain.EDIDocument) (domain.EDIDocument, error) {
	if doc.TenantID == uuid.Nil || doc.SupplierID == uuid.Nil || doc.DocType == "" || doc.Payload == "" {
		return domain.EDIDocument{}, domain.ErrInvalidArgument
	}
	doc.ID = d.newID()
	doc.Status = "mapped"
	doc.MappedRef = fmt.Sprintf("%s-%s", doc.DocType, doc.ID.String()[:8])
	doc.CreatedAt = d.now()
	if doc.Direction == "" {
		doc.Direction = "in"
	}
	if err := d.EDI.Save(ctx, doc); err != nil {
		return domain.EDIDocument{}, err
	}
	// ASN→shipment mapping requires an explicit PO linkage via AnnounceShipment; EDI store only here.
	return doc, nil
}

func (d *Deps) RateSupplier(ctx context.Context, sc domain.Scorecard) (domain.Scorecard, error) {
	if sc.TenantID == uuid.Nil || sc.SupplierID == uuid.Nil || sc.Period == "" {
		return domain.Scorecard{}, domain.ErrInvalidArgument
	}
	sc.ID = d.newID()
	sc.Overall = domain.ComputeOverallScore(sc)
	sc.CreatedAt = d.now()
	if err := d.Scorecards.Save(ctx, sc); err != nil {
		return domain.Scorecard{}, err
	}
	s, err := d.Suppliers.Get(ctx, sc.TenantID, sc.SupplierID)
	if err == nil {
		s.RiskScore = sc.RiskScore
		s.UpdatedAt = d.now()
		_ = d.Suppliers.Save(ctx, s)
	}
	d.emit(ctx, sc.TenantID, sc.ID, domain.EventSupplierRated, map[string]any{
		"supplierId": sc.SupplierID.String(), "overall": sc.Overall,
	})
	return sc, nil
}

func (d *Deps) PostMessage(ctx context.Context, tenantID, supplierID uuid.UUID, subject, sender, body string) (domain.MessageThread, domain.Message, error) {
	if tenantID == uuid.Nil || supplierID == uuid.Nil || body == "" {
		return domain.MessageThread{}, domain.Message{}, domain.ErrInvalidArgument
	}
	th := domain.MessageThread{ID: d.newID(), TenantID: tenantID, SupplierID: supplierID, Subject: subject, CreatedAt: d.now()}
	if err := d.Messages.SaveThread(ctx, th); err != nil {
		return domain.MessageThread{}, domain.Message{}, err
	}
	msg := domain.Message{ID: d.newID(), TenantID: tenantID, ThreadID: th.ID, Sender: sender, Body: body, CreatedAt: d.now()}
	if err := d.Messages.SaveMessage(ctx, msg); err != nil {
		return domain.MessageThread{}, domain.Message{}, err
	}
	return th, msg, nil
}

func (d *Deps) RequestChange(ctx context.Context, c domain.ChangeRequest) (domain.ChangeRequest, error) {
	if c.TenantID == uuid.Nil || c.Kind == "" || c.SubjectKey == "" {
		return domain.ChangeRequest{}, domain.ErrInvalidArgument
	}
	c.ID = d.newID()
	c.Status = "pending"
	c.CreatedAt = d.now()
	if c.Payload == nil {
		c.Payload = map[string]any{}
	}
	if err := d.Changes.Save(ctx, c); err != nil {
		return domain.ChangeRequest{}, err
	}
	return c, nil
}

func (d *Deps) DecideChange(ctx context.Context, tenantID, id uuid.UUID, approve bool) (domain.ChangeRequest, error) {
	c, err := d.Changes.Get(ctx, tenantID, id)
	if err != nil {
		return domain.ChangeRequest{}, err
	}
	if c.Status != "pending" {
		return domain.ChangeRequest{}, domain.ErrIllegalTransition
	}
	now := d.now()
	c.DecidedAt = &now
	if approve {
		c.Status = "approved"
	} else {
		c.Status = "rejected"
	}
	if err := d.Changes.Save(ctx, c); err != nil {
		return domain.ChangeRequest{}, err
	}
	return c, nil
}

func (d *Deps) RecommendSuppliers(ctx context.Context, tenantID uuid.UUID, sku string, limit int) ([]uuid.UUID, error) {
	if d.AI == nil {
		all, err := d.Suppliers.List(ctx, tenantID)
		if err != nil {
			return nil, err
		}
		out := make([]uuid.UUID, 0)
		for _, s := range all {
			if s.Status == domain.SupplierApproved {
				out = append(out, s.ID)
				if limit > 0 && len(out) >= limit {
					break
				}
			}
		}
		return out, nil
	}
	return d.AI.RecommendSuppliers(ctx, tenantID, sku, limit)
}

func (d *Deps) PortalSnapshot(ctx context.Context, tenantID, supplierID uuid.UUID) (map[string]any, error) {
	if _, err := d.Suppliers.Get(ctx, tenantID, supplierID); err != nil {
		return nil, err
	}
	pos, _ := d.POs.ListBySupplier(ctx, tenantID, supplierID)
	docs, _ := d.Documents.ListBySupplier(ctx, tenantID, supplierID)
	contracts, _ := d.Contracts.ListBySupplier(ctx, tenantID, supplierID)
	scores, _ := d.Scorecards.ListBySupplier(ctx, tenantID, supplierID)
	return map[string]any{
		"orders": pos, "documents": docs, "contracts": contracts, "scorecards": scores,
	}, nil
}

func (d *Deps) AdminStats(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	sups, _ := d.Suppliers.List(ctx, tenantID)
	pos, _ := d.POs.List(ctx, tenantID)
	sellers, _ := d.Sellers.List(ctx, tenantID)
	edi, _ := d.EDI.List(ctx, tenantID)
	approved := 0
	for _, s := range sups {
		if s.Status == domain.SupplierApproved {
			approved++
		}
	}
	return map[string]any{
		"suppliers": len(sups), "approved": approved, "purchaseOrders": len(pos),
		"sellers": len(sellers), "ediDocuments": len(edi),
	}, nil
}

func (d *Deps) MerchantDashboard(ctx context.Context, tenantID uuid.UUID) (map[string]any, error) {
	if tenantID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	sellers, err := d.ListSellers(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	listings, err := d.ListListings(ctx, tenantID, uuid.Nil)
	if err != nil {
		listings = []domain.ListingRef{}
	}
	var pos []domain.SourcingPurchaseOrder
	if d.POs != nil {
		pos, _ = d.POs.List(ctx, tenantID)
	}
	if pos == nil {
		pos = []domain.SourcingPurchaseOrder{}
	}
	if sellers == nil {
		sellers = []domain.MarketplaceSeller{}
	}
	activeListings := 0
	stockUnits := int64(0)
	for _, l := range listings {
		if l.Active {
			activeListings++
		}
		stockUnits += l.StockHint
	}
	var profile any
	if len(sellers) > 0 {
		profile = sellers[0]
	}
	return map[string]any{
		"profile": profile,
		"status":  "ready",
		"sellers": sellers,
		"listings": listings,
		"orders": pos,
		"inventory": listings,
		"summary": map[string]any{
			"activeOrders": len(pos),
			"products":     activeListings,
			"inventory":    stockUnits,
			"sellers":      len(sellers),
		},
	}, nil
}
