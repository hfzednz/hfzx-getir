package httpadapter

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/erp-service/internal/app"
	"github.com/nexora/erp-service/internal/domain"
	"github.com/nexora/erp-service/internal/ratelimit"
)

type Handler struct{ Deps *app.Deps }

type ServerConfig struct {
	Addr               string
	Deps               *app.Deps
	Limiter            ratelimit.Limiter
	RateLimitPerMinute int
	CORSOrigins        []string
	Log                *slog.Logger
}

func NewHandler(cfg ServerConfig) http.Handler {
	h := &Handler{Deps: cfg.Deps}
	mux := http.NewServeMux()
	const base = "/v1/erp"

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /ready", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, map[string]string{"status": "ready"})
	})

	mux.HandleFunc("POST "+base+"/companies", tenant(h.upsertCompany))
	mux.HandleFunc("GET "+base+"/companies", tenant(h.listCompanies))
	mux.HandleFunc("POST "+base+"/fiscal-years", tenant(h.openFiscalYear))
	mux.HandleFunc("POST "+base+"/periods/{id}/close", tenant(h.closePeriod))
	mux.HandleFunc("GET "+base+"/periods", tenant(h.listPeriods))
	mux.HandleFunc("POST "+base+"/accounts", tenant(h.upsertAccount))
	mux.HandleFunc("GET "+base+"/accounts", tenant(h.listAccounts))
	mux.HandleFunc("POST "+base+"/journals", tenant(h.postJournal))

	mux.HandleFunc("POST "+base+"/suppliers", tenant(h.upsertSupplier))
	mux.HandleFunc("GET "+base+"/suppliers", tenant(h.listSuppliers))
	mux.HandleFunc("POST "+base+"/purchase-requests", tenant(h.createPR))
	mux.HandleFunc("POST "+base+"/purchase-orders", tenant(h.createPO))
	mux.HandleFunc("POST "+base+"/goods-receipts", tenant(h.receiveGoods))

	mux.HandleFunc("POST "+base+"/ap/invoices", tenant(h.createAP))
	mux.HandleFunc("POST "+base+"/ap/invoices/{id}/approve", tenant(h.approveAP))
	mux.HandleFunc("POST "+base+"/ap/invoices/{id}/schedule-payment", tenant(h.scheduleAP))
	mux.HandleFunc("GET "+base+"/ap/invoices", tenant(h.listAP))
	mux.HandleFunc("POST "+base+"/ar/invoices", tenant(h.createAR))

	mux.HandleFunc("POST "+base+"/treasury/banks", tenant(h.upsertBank))
	mux.HandleFunc("GET "+base+"/treasury/banks", tenant(h.listBanks))
	mux.HandleFunc("POST "+base+"/treasury/transactions", tenant(h.importTxn))
	mux.HandleFunc("POST "+base+"/treasury/transactions/{id}/reconcile", tenant(h.reconcileTxn))

	mux.HandleFunc("POST "+base+"/budgets", tenant(h.upsertBudget))
	mux.HandleFunc("POST "+base+"/budgets/{id}/approve", tenant(h.approveBudget))
	mux.HandleFunc("GET "+base+"/budgets", tenant(h.listBudgets))

	mux.HandleFunc("POST "+base+"/assets", tenant(h.createAsset))
	mux.HandleFunc("POST "+base+"/assets/depreciate", tenant(h.depreciate))
	mux.HandleFunc("GET "+base+"/assets", tenant(h.listAssets))

	mux.HandleFunc("POST "+base+"/tax/returns", tenant(h.calcTax))
	mux.HandleFunc("GET "+base+"/tax/returns", tenant(h.listTax))

	mux.HandleFunc("POST "+base+"/expenses", tenant(h.submitExpense))
	mux.HandleFunc("GET "+base+"/approvals", tenant(h.listApprovals))
	mux.HandleFunc("POST "+base+"/approvals/{id}/decide", tenant(h.decideApproval))

	mux.HandleFunc("POST "+base+"/payroll/batches", tenant(h.exportPayroll))
	mux.HandleFunc("GET "+base+"/ai/cashflow", tenant(h.cashflow))
	mux.HandleFunc("GET "+base+"/admin/stats", tenant(h.stats))
	mux.HandleFunc("POST "+base+"/outbox/publish", tenant(h.outbox))

	return chain(mux,
		requestIDMiddleware,
		recoverMiddleware(cfg.Log),
		loggingMiddleware(cfg.Log),
		corsMiddleware(cfg.CORSOrigins),
		rateLimitMiddleware(cfg.Limiter, cfg.RateLimitPerMinute),
	)
}

func NewServer(cfg ServerConfig) *http.Server {
	return &http.Server{
		Addr: cfg.Addr, Handler: NewHandler(cfg),
		ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second,
		WriteTimeout: 60 * time.Second, IdleTimeout: 60 * time.Second,
	}
}

func tenant(next func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenantMiddleware(http.HandlerFunc(next)).ServeHTTP(w, r)
	}
}

func requireTenant(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	tid, ok := TenantIDFromContext(r.Context())
	if !ok {
		writeErr(w, r, domain.ErrInvalidArgument)
		return uuid.Nil, false
	}
	return tid, true
}

func parseUUID(s string) (uuid.UUID, error) { return uuid.Parse(s) }

func (h *Handler) upsertCompany(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Company
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	c, err := h.Deps.UpsertCompany(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, c)
}

func (h *Handler) listCompanies(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Companies.List(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) openFiscalYear(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		domain.FiscalYear
		Months int `json:"months"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	y, periods, err := h.Deps.OpenFiscalYear(r.Context(), body.FiscalYear, body.Months)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, map[string]any{"fiscalYear": y, "periods": periods})
}

func (h *Handler) closePeriod(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	p, err := h.Deps.ClosePeriod(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, p)
}

func (h *Handler) listPeriods(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Periods.ListPeriods(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) upsertAccount(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ChartAccount
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.UpsertAccount(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) listAccounts(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Accounts.List(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) postJournal(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.JournalEntry
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	if key := r.Header.Get("Idempotency-Key"); key != "" && body.IdempotencyKey == "" {
		body.IdempotencyKey = key
	}
	j, err := h.Deps.PostJournal(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, j)
}

func (h *Handler) upsertSupplier(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Supplier
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	s, err := h.Deps.UpsertSupplier(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, s)
}

func (h *Handler) listSuppliers(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Suppliers.List(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) createPR(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PurchaseRequest
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.CreatePR(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) createPO(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PurchaseOrder
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.CreatePO(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) receiveGoods(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		domain.GoodsReceipt
		WarehouseRef string `json:"warehouseRef"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	g, err := h.Deps.ReceiveGoods(r.Context(), body.GoodsReceipt, body.WarehouseRef)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, g)
}

func (h *Handler) createAP(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		domain.APInvoice
		TaxRateBps int64 `json:"taxRateBps"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	inv, err := h.Deps.CreateAPInvoice(r.Context(), body.APInvoice, body.TaxRateBps)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, inv)
}

func (h *Handler) approveAP(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	approver, _ := UserIDFromContext(r.Context())
	inv, err := h.Deps.ApproveAPInvoice(r.Context(), tid, id, approver)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, inv)
}

func (h *Handler) scheduleAP(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	inv, ref, err := h.Deps.ScheduleAPPayment(r.Context(), tid, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"invoice": inv, "batchRef": ref})
}

func (h *Handler) listAP(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.AP.List(r.Context(), tid, cid, r.URL.Query().Get("status"))
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) createAR(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ARInvoice
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	inv, err := h.Deps.CreateARInvoice(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, inv)
}

func (h *Handler) upsertBank(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.BankAccount
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	b, err := h.Deps.UpsertBank(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, b)
}

func (h *Handler) listBanks(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Treasury.ListBanks(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) importTxn(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.BankTransaction
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	t, err := h.Deps.ImportBankTxn(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, t)
}

func (h *Handler) reconcileTxn(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	bankID, err := parseUUID(r.URL.Query().Get("bankAccountId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	t, err := h.Deps.ReconcileTxn(r.Context(), tid, bankID, id)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, t)
}

func (h *Handler) upsertBudget(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.Budget
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	b, err := h.Deps.UpsertBudget(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, b)
}

func (h *Handler) approveBudget(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	approver, _ := UserIDFromContext(r.Context())
	b, err := h.Deps.ApproveBudget(r.Context(), tid, id, approver)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, b)
}

func (h *Handler) listBudgets(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Budgets.List(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) createAsset(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.FixedAsset
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	a, err := h.Deps.CreateAsset(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, a)
}

func (h *Handler) depreciate(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		CompanyID uuid.UUID `json:"companyId"`
		PeriodID  uuid.UUID `json:"periodId"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	n, err := h.Deps.RunDepreciation(r.Context(), tid, body.CompanyID, body.PeriodID)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"depreciatedAssets": n})
}

func (h *Handler) listAssets(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Assets.List(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) calcTax(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body struct {
		domain.TaxReturn
		RateBps int64 `json:"rateBps"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	t, err := h.Deps.CalculateTax(r.Context(), body.TaxReturn, body.RateBps)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, t)
}

func (h *Handler) listTax(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	items, err := h.Deps.Tax.List(r.Context(), tid, cid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) submitExpense(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.ExpenseReport
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	e, err := h.Deps.SubmitExpense(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, e)
}

func (h *Handler) listApprovals(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	items, err := h.Deps.Approvals.ListPending(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"items": items})
}

func (h *Handler) decideApproval(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	id, err := parseUUID(r.PathValue("id"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	var body struct {
		Approve bool   `json:"approve"`
		Note    string `json:"note"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	actor, _ := UserIDFromContext(r.Context())
	a, err := h.Deps.DecideApproval(r.Context(), tid, id, actor, body.Approve, body.Note)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, a)
}

func (h *Handler) exportPayroll(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	var body domain.PayrollBatch
	if err := decodeJSON(r, &body); err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	body.TenantID = tid
	p, err := h.Deps.ExportPayroll(r.Context(), body)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeCreated(w, p)
}

func (h *Handler) cashflow(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	cid, err := parseUUID(r.URL.Query().Get("companyId"))
	if err != nil {
		writeErr(w, r, domain.ErrInvalidArgument)
		return
	}
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days <= 0 {
		days = 30
	}
	m, err := h.Deps.CashflowForecast(r.Context(), tid, cid, days)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, m)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	tid, ok := requireTenant(w, r)
	if !ok {
		return
	}
	s, err := h.Deps.AdminStats(r.Context(), tid)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, s)
}

func (h *Handler) outbox(w http.ResponseWriter, r *http.Request) {
	n, err := h.Deps.PublishPending(r.Context(), 100)
	if err != nil {
		writeErr(w, r, err)
		return
	}
	writeOK(w, map[string]any{"published": n})
}
