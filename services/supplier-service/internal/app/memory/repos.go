package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/supplier-service/internal/app/ports"
	"github.com/nexora/supplier-service/internal/domain"
)

type Store struct {
	mu          sync.RWMutex
	Suppliers   map[uuid.UUID]domain.Supplier
	Documents   map[uuid.UUID]domain.SupplierDocument
	Certs       map[uuid.UUID]domain.Certification
	Contracts   map[uuid.UUID]domain.Contract
	RFQs        map[uuid.UUID]domain.RFQ
	Quotes      map[uuid.UUID]domain.Quotation
	POs         map[uuid.UUID]domain.SourcingPurchaseOrder
	Shipments   map[uuid.UUID]domain.InboundShipment
	Invoices    map[uuid.UUID]domain.InvoiceMatchSignal
	Sellers     map[uuid.UUID]domain.MarketplaceSeller
	Listings    map[uuid.UUID]domain.ListingRef
	Submissions map[uuid.UUID]domain.CatalogSubmission
	EDI         map[uuid.UUID]domain.EDIDocument
	Scorecards  map[uuid.UUID]domain.Scorecard
	Threads     map[uuid.UUID]domain.MessageThread
	Messages    map[uuid.UUID]domain.Message
	Changes     map[uuid.UUID]domain.ChangeRequest
	Outbox      map[uuid.UUID]domain.OutboxMessage
}

func NewStore() *Store {
	return &Store{
		Suppliers: map[uuid.UUID]domain.Supplier{}, Documents: map[uuid.UUID]domain.SupplierDocument{},
		Certs: map[uuid.UUID]domain.Certification{}, Contracts: map[uuid.UUID]domain.Contract{},
		RFQs: map[uuid.UUID]domain.RFQ{}, Quotes: map[uuid.UUID]domain.Quotation{},
		POs: map[uuid.UUID]domain.SourcingPurchaseOrder{}, Shipments: map[uuid.UUID]domain.InboundShipment{},
		Invoices: map[uuid.UUID]domain.InvoiceMatchSignal{}, Sellers: map[uuid.UUID]domain.MarketplaceSeller{},
		Listings: map[uuid.UUID]domain.ListingRef{}, Submissions: map[uuid.UUID]domain.CatalogSubmission{},
		EDI: map[uuid.UUID]domain.EDIDocument{}, Scorecards: map[uuid.UUID]domain.Scorecard{},
		Threads: map[uuid.UUID]domain.MessageThread{}, Messages: map[uuid.UUID]domain.Message{},
		Changes: map[uuid.UUID]domain.ChangeRequest{}, Outbox: map[uuid.UUID]domain.OutboxMessage{},
	}
}

type Repos struct {
	Suppliers   *SupplierRepo
	Documents   *DocumentRepo
	Certs       *CertRepo
	Contracts   *ContractRepo
	RFQs        *RFQRepo
	Quotes      *QuoteRepo
	POs         *PORepo
	Shipments   *ShipmentRepo
	Invoices    *InvoiceRepo
	Sellers     *SellerRepo
	Listings    *ListingRepo
	Submissions *SubmissionRepo
	EDI         *EDIRepo
	Scorecards  *ScorecardRepo
	Messages    *MessageRepo
	Changes     *ChangeRepo
	Outbox      *OutboxRepo
	ERP         *MockERP
	Catalog     *MockCatalog
	Inventory   *MockInventory
	Settlement  *MockSettlement
	AI          *MockAI
	Metrics     *MockMetrics
}

func NewRepos(s *Store) *Repos {
	return &Repos{
		Suppliers: &SupplierRepo{s: s}, Documents: &DocumentRepo{s: s}, Certs: &CertRepo{s: s},
		Contracts: &ContractRepo{s: s}, RFQs: &RFQRepo{s: s}, Quotes: &QuoteRepo{s: s},
		POs: &PORepo{s: s}, Shipments: &ShipmentRepo{s: s}, Invoices: &InvoiceRepo{s: s},
		Sellers: &SellerRepo{s: s}, Listings: &ListingRepo{s: s}, Submissions: &SubmissionRepo{s: s},
		EDI: &EDIRepo{s: s}, Scorecards: &ScorecardRepo{s: s}, Messages: &MessageRepo{s: s},
		Changes: &ChangeRepo{s: s}, Outbox: &OutboxRepo{s: s},
		ERP: &MockERP{}, Catalog: &MockCatalog{}, Inventory: &MockInventory{},
		Settlement: &MockSettlement{}, AI: &MockAI{}, Metrics: &MockMetrics{},
	}
}

type SupplierRepo struct{ s *Store }

func (r *SupplierRepo) Save(_ context.Context, s domain.Supplier) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Suppliers[s.ID] = s
	return nil
}
func (r *SupplierRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Suppliers[id]
	if !ok || s.TenantID != tenantID {
		return domain.Supplier{}, domain.ErrNotFound
	}
	return s, nil
}
func (r *SupplierRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Supplier, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	for _, s := range r.s.Suppliers {
		if s.TenantID == tenantID && s.Code == code {
			return s, nil
		}
	}
	return domain.Supplier{}, domain.ErrNotFound
}
func (r *SupplierRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.Supplier, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Supplier{}
	for _, s := range r.s.Suppliers {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type DocumentRepo struct{ s *Store }

func (r *DocumentRepo) Save(_ context.Context, d domain.SupplierDocument) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Documents[d.ID] = d
	return nil
}
func (r *DocumentRepo) ListBySupplier(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SupplierDocument{}
	for _, d := range r.s.Documents {
		if d.TenantID == tenantID && d.SupplierID == supplierID {
			out = append(out, d)
		}
	}
	return out, nil
}

type CertRepo struct{ s *Store }

func (r *CertRepo) Save(_ context.Context, c domain.Certification) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Certs[c.ID] = c
	return nil
}
func (r *CertRepo) ListBySupplier(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.Certification, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Certification{}
	for _, c := range r.s.Certs {
		if c.TenantID == tenantID && c.SupplierID == supplierID {
			out = append(out, c)
		}
	}
	return out, nil
}

type ContractRepo struct{ s *Store }

func (r *ContractRepo) Save(_ context.Context, c domain.Contract) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Contracts[c.ID] = c
	return nil
}
func (r *ContractRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Contract, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.Contracts[id]
	if !ok || c.TenantID != tenantID {
		return domain.Contract{}, domain.ErrNotFound
	}
	return c, nil
}
func (r *ContractRepo) ListBySupplier(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.Contract, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Contract{}
	for _, c := range r.s.Contracts {
		if c.TenantID == tenantID && c.SupplierID == supplierID {
			out = append(out, c)
		}
	}
	return out, nil
}

type RFQRepo struct{ s *Store }

func (r *RFQRepo) Save(_ context.Context, rfq domain.RFQ) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.RFQs[rfq.ID] = rfq
	return nil
}
func (r *RFQRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.RFQ, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	x, ok := r.s.RFQs[id]
	if !ok || x.TenantID != tenantID {
		return domain.RFQ{}, domain.ErrNotFound
	}
	return x, nil
}
func (r *RFQRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.RFQ, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.RFQ{}
	for _, x := range r.s.RFQs {
		if x.TenantID == tenantID {
			out = append(out, x)
		}
	}
	return out, nil
}

type QuoteRepo struct{ s *Store }

func (r *QuoteRepo) Save(_ context.Context, q domain.Quotation) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Quotes[q.ID] = q
	return nil
}
func (r *QuoteRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.Quotation, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	q, ok := r.s.Quotes[id]
	if !ok || q.TenantID != tenantID {
		return domain.Quotation{}, domain.ErrNotFound
	}
	return q, nil
}
func (r *QuoteRepo) ListByRFQ(_ context.Context, tenantID, rfqID uuid.UUID) ([]domain.Quotation, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Quotation{}
	for _, q := range r.s.Quotes {
		if q.TenantID == tenantID && q.RFQID == rfqID {
			out = append(out, q)
		}
	}
	return out, nil
}

type PORepo struct{ s *Store }

func (r *PORepo) Save(_ context.Context, po domain.SourcingPurchaseOrder) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.POs[po.ID] = po
	return nil
}
func (r *PORepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.SourcingPurchaseOrder, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	po, ok := r.s.POs[id]
	if !ok || po.TenantID != tenantID {
		return domain.SourcingPurchaseOrder{}, domain.ErrNotFound
	}
	return po, nil
}
func (r *PORepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.SourcingPurchaseOrder, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SourcingPurchaseOrder{}
	for _, po := range r.s.POs {
		if po.TenantID == tenantID {
			out = append(out, po)
		}
	}
	return out, nil
}
func (r *PORepo) ListBySupplier(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.SourcingPurchaseOrder, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.SourcingPurchaseOrder{}
	for _, po := range r.s.POs {
		if po.TenantID == tenantID && po.SupplierID == supplierID {
			out = append(out, po)
		}
	}
	return out, nil
}

type ShipmentRepo struct{ s *Store }

func (r *ShipmentRepo) Save(_ context.Context, s domain.InboundShipment) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Shipments[s.ID] = s
	return nil
}
func (r *ShipmentRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.InboundShipment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Shipments[id]
	if !ok || s.TenantID != tenantID {
		return domain.InboundShipment{}, domain.ErrNotFound
	}
	return s, nil
}
func (r *ShipmentRepo) ListByPO(_ context.Context, tenantID, poID uuid.UUID) ([]domain.InboundShipment, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.InboundShipment{}
	for _, s := range r.s.Shipments {
		if s.TenantID == tenantID && s.POID == poID {
			out = append(out, s)
		}
	}
	return out, nil
}

type InvoiceRepo struct{ s *Store }

func (r *InvoiceRepo) Save(_ context.Context, m domain.InvoiceMatchSignal) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Invoices[m.ID] = m
	return nil
}
func (r *InvoiceRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.InvoiceMatchSignal, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.InvoiceMatchSignal{}
	for _, m := range r.s.Invoices {
		if m.TenantID == tenantID {
			out = append(out, m)
		}
	}
	return out, nil
}

type SellerRepo struct{ s *Store }

func (r *SellerRepo) Save(_ context.Context, s domain.MarketplaceSeller) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Sellers[s.ID] = s
	return nil
}
func (r *SellerRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.MarketplaceSeller, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Sellers[id]
	if !ok || s.TenantID != tenantID {
		return domain.MarketplaceSeller{}, domain.ErrNotFound
	}
	return s, nil
}
func (r *SellerRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.MarketplaceSeller, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.MarketplaceSeller{}
	for _, s := range r.s.Sellers {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type ListingRepo struct{ s *Store }

func (r *ListingRepo) Save(_ context.Context, l domain.ListingRef) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Listings[l.ID] = l
	return nil
}
func (r *ListingRepo) ListBySeller(_ context.Context, tenantID, sellerID uuid.UUID) ([]domain.ListingRef, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.ListingRef{}
	for _, l := range r.s.Listings {
		if l.TenantID == tenantID && l.SellerID == sellerID {
			out = append(out, l)
		}
	}
	return out, nil
}

type SubmissionRepo struct{ s *Store }

func (r *SubmissionRepo) Save(_ context.Context, s domain.CatalogSubmission) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Submissions[s.ID] = s
	return nil
}
func (r *SubmissionRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.CatalogSubmission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	s, ok := r.s.Submissions[id]
	if !ok || s.TenantID != tenantID {
		return domain.CatalogSubmission{}, domain.ErrNotFound
	}
	return s, nil
}
func (r *SubmissionRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.CatalogSubmission, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.CatalogSubmission{}
	for _, s := range r.s.Submissions {
		if s.TenantID == tenantID {
			out = append(out, s)
		}
	}
	return out, nil
}

type EDIRepo struct{ s *Store }

func (r *EDIRepo) Save(_ context.Context, d domain.EDIDocument) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.EDI[d.ID] = d
	return nil
}
func (r *EDIRepo) List(_ context.Context, tenantID uuid.UUID) ([]domain.EDIDocument, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.EDIDocument{}
	for _, d := range r.s.EDI {
		if d.TenantID == tenantID {
			out = append(out, d)
		}
	}
	return out, nil
}

type ScorecardRepo struct{ s *Store }

func (r *ScorecardRepo) Save(_ context.Context, sc domain.Scorecard) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Scorecards[sc.ID] = sc
	return nil
}
func (r *ScorecardRepo) ListBySupplier(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.Scorecard, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Scorecard{}
	for _, sc := range r.s.Scorecards {
		if sc.TenantID == tenantID && sc.SupplierID == supplierID {
			out = append(out, sc)
		}
	}
	return out, nil
}

type MessageRepo struct{ s *Store }

func (r *MessageRepo) SaveThread(_ context.Context, t domain.MessageThread) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Threads[t.ID] = t
	return nil
}
func (r *MessageRepo) SaveMessage(_ context.Context, m domain.Message) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Messages[m.ID] = m
	return nil
}
func (r *MessageRepo) ListThreads(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.MessageThread, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.MessageThread{}
	for _, t := range r.s.Threads {
		if t.TenantID == tenantID && t.SupplierID == supplierID {
			out = append(out, t)
		}
	}
	return out, nil
}
func (r *MessageRepo) ListMessages(_ context.Context, tenantID, threadID uuid.UUID) ([]domain.Message, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.Message{}
	for _, m := range r.s.Messages {
		if m.TenantID == tenantID && m.ThreadID == threadID {
			out = append(out, m)
		}
	}
	return out, nil
}

type ChangeRepo struct{ s *Store }

func (r *ChangeRepo) Save(_ context.Context, c domain.ChangeRequest) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Changes[c.ID] = c
	return nil
}
func (r *ChangeRepo) Get(_ context.Context, tenantID, id uuid.UUID) (domain.ChangeRequest, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	c, ok := r.s.Changes[id]
	if !ok || c.TenantID != tenantID {
		return domain.ChangeRequest{}, domain.ErrNotFound
	}
	return c, nil
}

type OutboxRepo struct{ s *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}
func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	out := []domain.OutboxMessage{}
	for _, m := range r.s.Outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.Outbox[m.ID] = m
	return nil
}

type MockERP struct{}

func (MockERP) SyncSupplier(_ context.Context, _ uuid.UUID, s domain.Supplier) (string, error) {
	return "erp-sup-" + s.Code, nil
}
func (MockERP) EnsurePO(_ context.Context, _ uuid.UUID, po domain.SourcingPurchaseOrder) (string, error) {
	return "erp-po-" + po.Number, nil
}

type MockCatalog struct{}

func (MockCatalog) SubmitProduct(context.Context, uuid.UUID, domain.CatalogSubmission) error { return nil }

type MockInventory struct{}

func (MockInventory) AnnounceASN(context.Context, uuid.UUID, domain.InboundShipment) error { return nil }

type MockSettlement struct{}

func (MockSettlement) ScheduleSellerPayout(context.Context, uuid.UUID, uuid.UUID, int64, string) error {
	return nil
}

type MockAI struct{}

func (MockAI) RecommendSuppliers(_ context.Context, _ uuid.UUID, _ string, limit int) ([]uuid.UUID, error) {
	return nil, nil
}
func (MockAI) PredictRisk(context.Context, domain.Supplier) (float64, error) { return 0.2, nil }
func (MockAI) RankSuppliers(_ context.Context, cards []domain.Scorecard) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		out = append(out, c.SupplierID)
	}
	return out, nil
}

type MockMetrics struct{}

func (MockMetrics) Record(context.Context, string, map[string]string, float64) error { return nil }

var (
	_ ports.SupplierRepo    = (*SupplierRepo)(nil)
	_ ports.OutboxRepository = (*OutboxRepo)(nil)
)
