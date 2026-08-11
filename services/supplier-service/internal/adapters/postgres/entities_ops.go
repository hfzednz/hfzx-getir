package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/nexora/supplier-service/internal/app/ports"
	"github.com/nexora/supplier-service/internal/domain"
)

type PORepo struct{ DB *sql.DB }

func (r *PORepo) Save(ctx context.Context, po domain.SourcingPurchaseOrder) error {
	lines, err := marshalJSON(po.Lines)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `
		INSERT INTO supplier_sourcing_pos (
			id, tenant_id, company_id, supplier_id, number, status, currency, total_minor,
			lines_json, quotation_id, erp_po_id, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (tenant_id, number) DO UPDATE SET
			id=EXCLUDED.id, status=EXCLUDED.status, currency=EXCLUDED.currency, total_minor=EXCLUDED.total_minor,
			lines_json=EXCLUDED.lines_json, quotation_id=EXCLUDED.quotation_id, erp_po_id=EXCLUDED.erp_po_id,
			updated_at=EXCLUDED.updated_at`,
		po.ID, po.TenantID, po.CompanyID, po.SupplierID, po.Number, string(po.Status), po.Currency, po.TotalMinor,
		lines, nullUUID(po.QuotationID), po.ErpPOID, po.CreatedAt.UTC(), po.UpdatedAt.UTC())
	return err
}

func (r *PORepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.SourcingPurchaseOrder, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, number, status, currency, total_minor,
			lines_json, quotation_id, erp_po_id, created_at, updated_at
		FROM supplier_sourcing_pos WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	return scanPO(row)
}

func (r *PORepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.SourcingPurchaseOrder, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, number, status, currency, total_minor,
			lines_json, quotation_id, erp_po_id, created_at, updated_at
		FROM supplier_sourcing_pos WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPOs(rows)
}

func (r *PORepo) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.SourcingPurchaseOrder, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, company_id, supplier_id, number, status, currency, total_minor,
			lines_json, quotation_id, erp_po_id, created_at, updated_at
		FROM supplier_sourcing_pos WHERE tenant_id=$1 AND supplier_id=$2 ORDER BY created_at DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPOs(rows)
}

func scanPO(row scannable) (domain.SourcingPurchaseOrder, error) {
	var po domain.SourcingPurchaseOrder
	var status string
	var raw []byte
	var qid uuid.NullUUID
	err := row.Scan(&po.ID, &po.TenantID, &po.CompanyID, &po.SupplierID, &po.Number, &status, &po.Currency, &po.TotalMinor,
		&raw, &qid, &po.ErpPOID, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.SourcingPurchaseOrder{}, domain.ErrNotFound
		}
		return domain.SourcingPurchaseOrder{}, err
	}
	po.Status = domain.SourcingPOStatus(status)
	_ = unmarshalJSON(raw, &po.Lines)
	po.QuotationID = scanNullUUID(qid)
	po.CreatedAt = po.CreatedAt.UTC()
	po.UpdatedAt = po.UpdatedAt.UTC()
	return po, nil
}

func scanPOs(rows *sql.Rows) ([]domain.SourcingPurchaseOrder, error) {
	out := []domain.SourcingPurchaseOrder{}
	for rows.Next() {
		po, err := scanPO(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, po)
	}
	return out, rows.Err()
}

type ShipmentRepo struct{ DB *sql.DB }

func (r *ShipmentRepo) Save(ctx context.Context, s domain.InboundShipment) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_shipments (
			id, tenant_id, supplier_id, po_id, asn_number, status, tracking_ref, warehouse_id, qc_passed, created_at, received_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, tracking_ref=EXCLUDED.tracking_ref,
			qc_passed=EXCLUDED.qc_passed, received_at=EXCLUDED.received_at`,
		s.ID, s.TenantID, s.SupplierID, s.POID, s.ASNNumber, string(s.Status), s.TrackingRef, s.WarehouseID,
		nullBool(s.QCPassed), s.CreatedAt.UTC(), nullTime(s.ReceivedAt))
	return err
}

func (r *ShipmentRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.InboundShipment, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, supplier_id, po_id, asn_number, status, tracking_ref, warehouse_id, qc_passed, created_at, received_at
		FROM supplier_shipments WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var s domain.InboundShipment
	var status string
	var qc sql.NullBool
	var received sql.NullTime
	err := row.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.POID, &s.ASNNumber, &status, &s.TrackingRef, &s.WarehouseID, &qc, &s.CreatedAt, &received)
	if err != nil {
		if isNoRows(err) {
			return domain.InboundShipment{}, domain.ErrNotFound
		}
		return domain.InboundShipment{}, err
	}
	s.Status = domain.ShipmentStatus(status)
	s.QCPassed = scanNullBool(qc)
	s.ReceivedAt = scanNullTime(received)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

func (r *ShipmentRepo) ListByPO(ctx context.Context, tenantID, poID uuid.UUID) ([]domain.InboundShipment, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, po_id, asn_number, status, tracking_ref, warehouse_id, qc_passed, created_at, received_at
		FROM supplier_shipments WHERE tenant_id=$1 AND po_id=$2 ORDER BY created_at ASC`, tenantID, poID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InboundShipment{}
	for rows.Next() {
		var s domain.InboundShipment
		var status string
		var qc sql.NullBool
		var received sql.NullTime
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.POID, &s.ASNNumber, &status, &s.TrackingRef, &s.WarehouseID, &qc, &s.CreatedAt, &received); err != nil {
			return nil, err
		}
		s.Status = domain.ShipmentStatus(status)
		s.QCPassed = scanNullBool(qc)
		s.ReceivedAt = scanNullTime(received)
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type InvoiceMatchRepo struct{ DB *sql.DB }

func (r *InvoiceMatchRepo) Save(ctx context.Context, m domain.InvoiceMatchSignal) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_invoice_matches (
			id, tenant_id, supplier_id, po_id, invoice_ref, amount_minor, currency, matched, match_score, erp_invoice_id, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET matched=EXCLUDED.matched, match_score=EXCLUDED.match_score, erp_invoice_id=EXCLUDED.erp_invoice_id`,
		m.ID, m.TenantID, m.SupplierID, m.POID, m.InvoiceRef, m.AmountMinor, m.Currency, m.Matched, m.MatchScore, m.ErpInvoiceID, m.CreatedAt.UTC())
	return err
}

func (r *InvoiceMatchRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.InvoiceMatchSignal, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, po_id, invoice_ref, amount_minor, currency, matched, match_score, erp_invoice_id, created_at
		FROM supplier_invoice_matches WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.InvoiceMatchSignal{}
	for rows.Next() {
		var m domain.InvoiceMatchSignal
		if err := rows.Scan(&m.ID, &m.TenantID, &m.SupplierID, &m.POID, &m.InvoiceRef, &m.AmountMinor, &m.Currency, &m.Matched, &m.MatchScore, &m.ErpInvoiceID, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

type SellerRepo struct{ DB *sql.DB }

func (r *SellerRepo) Save(ctx context.Context, s domain.MarketplaceSeller) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO marketplace_sellers (
			id, tenant_id, supplier_id, store_name, status, rating_avg, rating_count, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET store_name=EXCLUDED.store_name, status=EXCLUDED.status,
			rating_avg=EXCLUDED.rating_avg, rating_count=EXCLUDED.rating_count`,
		s.ID, s.TenantID, s.SupplierID, s.StoreName, string(s.Status), s.RatingAvg, s.RatingCount, s.CreatedAt.UTC())
	return err
}

func (r *SellerRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.MarketplaceSeller, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, supplier_id, store_name, status, rating_avg, rating_count, created_at
		FROM marketplace_sellers WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var s domain.MarketplaceSeller
	var status string
	err := row.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.StoreName, &status, &s.RatingAvg, &s.RatingCount, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.MarketplaceSeller{}, domain.ErrNotFound
		}
		return domain.MarketplaceSeller{}, err
	}
	s.Status = domain.SellerStatus(status)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

func (r *SellerRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.MarketplaceSeller, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, store_name, status, rating_avg, rating_count, created_at
		FROM marketplace_sellers WHERE tenant_id=$1 ORDER BY store_name ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MarketplaceSeller{}
	for rows.Next() {
		var s domain.MarketplaceSeller
		var status string
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.StoreName, &status, &s.RatingAvg, &s.RatingCount, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Status = domain.SellerStatus(status)
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type ListingRepo struct{ DB *sql.DB }

func (r *ListingRepo) Save(ctx context.Context, l domain.ListingRef) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO marketplace_listings (
			id, tenant_id, seller_id, external_sku, catalog_sku, price_minor, currency, stock_hint, active, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET catalog_sku=EXCLUDED.catalog_sku, price_minor=EXCLUDED.price_minor,
			stock_hint=EXCLUDED.stock_hint, active=EXCLUDED.active`,
		l.ID, l.TenantID, l.SellerID, l.ExternalSKU, l.CatalogSKU, l.PriceMinor, l.Currency, l.StockHint, l.Active, l.CreatedAt.UTC())
	return err
}

func (r *ListingRepo) ListBySeller(ctx context.Context, tenantID, sellerID uuid.UUID) ([]domain.ListingRef, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, seller_id, external_sku, catalog_sku, price_minor, currency, stock_hint, active, created_at
		FROM marketplace_listings WHERE tenant_id=$1 AND seller_id=$2 ORDER BY created_at DESC`, tenantID, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ListingRef{}
	for rows.Next() {
		var l domain.ListingRef
		if err := rows.Scan(&l.ID, &l.TenantID, &l.SellerID, &l.ExternalSKU, &l.CatalogSKU, &l.PriceMinor, &l.Currency, &l.StockHint, &l.Active, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.CreatedAt = l.CreatedAt.UTC()
		out = append(out, l)
	}
	return out, rows.Err()
}

type SubmissionRepo struct{ DB *sql.DB }

func (r *SubmissionRepo) Save(ctx context.Context, s domain.CatalogSubmission) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO catalog_submissions (
			id, tenant_id, supplier_id, sku, title, attributes, media_uris, version, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, attributes=EXCLUDED.attributes,
			media_uris=EXCLUDED.media_uris, version=EXCLUDED.version, status=EXCLUDED.status`,
		s.ID, s.TenantID, s.SupplierID, s.SKU, s.Title, JSONMap(s.Attributes), textArray(s.MediaURIs), s.Version, string(s.Status), s.CreatedAt.UTC())
	return err
}

func (r *SubmissionRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.CatalogSubmission, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, supplier_id, sku, title, attributes, media_uris, version, status, created_at
		FROM catalog_submissions WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var s domain.CatalogSubmission
	var attrs JSONMap
	var media []string
	var status string
	err := row.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.SKU, &s.Title, &attrs, pq.Array(&media), &s.Version, &status, &s.CreatedAt)
	if err != nil {
		if isNoRows(err) {
			return domain.CatalogSubmission{}, domain.ErrNotFound
		}
		return domain.CatalogSubmission{}, err
	}
	s.Attributes = map[string]any(attrs)
	s.MediaURIs = media
	s.Status = domain.SubmissionStatus(status)
	s.CreatedAt = s.CreatedAt.UTC()
	return s, nil
}

func (r *SubmissionRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.CatalogSubmission, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, sku, title, attributes, media_uris, version, status, created_at
		FROM catalog_submissions WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.CatalogSubmission{}
	for rows.Next() {
		var s domain.CatalogSubmission
		var attrs JSONMap
		var media []string
		var status string
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.SKU, &s.Title, &attrs, pq.Array(&media), &s.Version, &status, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.Attributes = map[string]any(attrs)
		s.MediaURIs = media
		s.Status = domain.SubmissionStatus(status)
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type EDIRepo struct{ DB *sql.DB }

func (r *EDIRepo) Save(ctx context.Context, d domain.EDIDocument) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO edi_documents (
			id, tenant_id, supplier_id, doc_type, direction, payload, mapped_ref, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, mapped_ref=EXCLUDED.mapped_ref, payload=EXCLUDED.payload`,
		d.ID, d.TenantID, d.SupplierID, string(d.DocType), d.Direction, d.Payload, d.MappedRef, d.Status, d.CreatedAt.UTC())
	return err
}

func (r *EDIRepo) List(ctx context.Context, tenantID uuid.UUID) ([]domain.EDIDocument, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, doc_type, direction, payload, mapped_ref, status, created_at
		FROM edi_documents WHERE tenant_id=$1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.EDIDocument{}
	for rows.Next() {
		var d domain.EDIDocument
		var docType string
		if err := rows.Scan(&d.ID, &d.TenantID, &d.SupplierID, &docType, &d.Direction, &d.Payload, &d.MappedRef, &d.Status, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.DocType = domain.EDIDocType(docType)
		d.CreatedAt = d.CreatedAt.UTC()
		out = append(out, d)
	}
	return out, rows.Err()
}

type ScorecardRepo struct{ DB *sql.DB }

func (r *ScorecardRepo) Save(ctx context.Context, s domain.Scorecard) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_scorecards (
			id, tenant_id, supplier_id, period, delivery_score, quality_score, price_score,
			lead_time_days_avg, fill_rate, compliance_score, risk_score, overall, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET delivery_score=EXCLUDED.delivery_score, quality_score=EXCLUDED.quality_score,
			price_score=EXCLUDED.price_score, overall=EXCLUDED.overall, risk_score=EXCLUDED.risk_score`,
		s.ID, s.TenantID, s.SupplierID, s.Period, s.DeliveryScore, s.QualityScore, s.PriceScore,
		s.LeadTimeDaysAvg, s.FillRate, s.ComplianceScore, s.RiskScore, s.Overall, s.CreatedAt.UTC())
	return err
}

func (r *ScorecardRepo) ListBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.Scorecard, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, period, delivery_score, quality_score, price_score,
			lead_time_days_avg, fill_rate, compliance_score, risk_score, overall, created_at
		FROM supplier_scorecards WHERE tenant_id=$1 AND supplier_id=$2 ORDER BY period DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Scorecard{}
	for rows.Next() {
		var s domain.Scorecard
		if err := rows.Scan(&s.ID, &s.TenantID, &s.SupplierID, &s.Period, &s.DeliveryScore, &s.QualityScore, &s.PriceScore,
			&s.LeadTimeDaysAvg, &s.FillRate, &s.ComplianceScore, &s.RiskScore, &s.Overall, &s.CreatedAt); err != nil {
			return nil, err
		}
		s.CreatedAt = s.CreatedAt.UTC()
		out = append(out, s)
	}
	return out, rows.Err()
}

type MessageRepo struct{ DB *sql.DB }

func (r *MessageRepo) SaveThread(ctx context.Context, t domain.MessageThread) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_threads (id, tenant_id, supplier_id, subject, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (id) DO UPDATE SET subject=EXCLUDED.subject`,
		t.ID, t.TenantID, t.SupplierID, t.Subject, t.CreatedAt.UTC())
	return err
}

func (r *MessageRepo) SaveMessage(ctx context.Context, m domain.Message) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_messages (id, tenant_id, thread_id, sender, body, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (id) DO UPDATE SET body=EXCLUDED.body`,
		m.ID, m.TenantID, m.ThreadID, m.Sender, m.Body, m.CreatedAt.UTC())
	return err
}

func (r *MessageRepo) ListThreads(ctx context.Context, tenantID, supplierID uuid.UUID) ([]domain.MessageThread, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, supplier_id, subject, created_at
		FROM supplier_threads WHERE tenant_id=$1 AND supplier_id=$2 ORDER BY created_at DESC`, tenantID, supplierID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.MessageThread{}
	for rows.Next() {
		var t domain.MessageThread
		if err := rows.Scan(&t.ID, &t.TenantID, &t.SupplierID, &t.Subject, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.CreatedAt = t.CreatedAt.UTC()
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *MessageRepo) ListMessages(ctx context.Context, tenantID, threadID uuid.UUID) ([]domain.Message, error) {
	rows, err := r.DB.QueryContext(ctx, `
		SELECT id, tenant_id, thread_id, sender, body, created_at
		FROM supplier_messages WHERE tenant_id=$1 AND thread_id=$2 ORDER BY created_at ASC`, tenantID, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Message{}
	for rows.Next() {
		var m domain.Message
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ThreadID, &m.Sender, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.CreatedAt = m.CreatedAt.UTC()
		out = append(out, m)
	}
	return out, rows.Err()
}

type ChangeRepo struct{ DB *sql.DB }

func (r *ChangeRepo) Save(ctx context.Context, c domain.ChangeRequest) error {
	_, err := r.DB.ExecContext(ctx, `
		INSERT INTO supplier_changes (
			id, tenant_id, kind, subject_key, payload, status, created_at, decided_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET status=EXCLUDED.status, payload=EXCLUDED.payload, decided_at=EXCLUDED.decided_at`,
		c.ID, c.TenantID, c.Kind, c.SubjectKey, JSONMap(c.Payload), c.Status, c.CreatedAt.UTC(), nullTime(c.DecidedAt))
	return err
}

func (r *ChangeRepo) Get(ctx context.Context, tenantID, id uuid.UUID) (domain.ChangeRequest, error) {
	row := r.DB.QueryRowContext(ctx, `
		SELECT id, tenant_id, kind, subject_key, payload, status, created_at, decided_at
		FROM supplier_changes WHERE id=$1 AND tenant_id=$2`, id, tenantID)
	var c domain.ChangeRequest
	var payload JSONMap
	var decided sql.NullTime
	err := row.Scan(&c.ID, &c.TenantID, &c.Kind, &c.SubjectKey, &payload, &c.Status, &c.CreatedAt, &decided)
	if err != nil {
		if isNoRows(err) {
			return domain.ChangeRequest{}, domain.ErrNotFound
		}
		return domain.ChangeRequest{}, err
	}
	c.Payload = map[string]any(payload)
	c.DecidedAt = scanNullTime(decided)
	c.CreatedAt = c.CreatedAt.UTC()
	return c, nil
}

var (
	_ ports.PORepo           = (*PORepo)(nil)
	_ ports.ShipmentRepo     = (*ShipmentRepo)(nil)
	_ ports.InvoiceMatchRepo = (*InvoiceMatchRepo)(nil)
	_ ports.SellerRepo       = (*SellerRepo)(nil)
	_ ports.ListingRepo      = (*ListingRepo)(nil)
	_ ports.SubmissionRepo   = (*SubmissionRepo)(nil)
	_ ports.EDIRepo          = (*EDIRepo)(nil)
	_ ports.ScorecardRepo    = (*ScorecardRepo)(nil)
	_ ports.MessageRepo      = (*MessageRepo)(nil)
	_ ports.ChangeRepo       = (*ChangeRepo)(nil)
)
