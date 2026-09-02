package app

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalid = errors.New("invalid argument")

type OrderGateway interface {
	List(ctx context.Context, tenant string, q url.Values) (map[string]any, error)
	Get(ctx context.Context, tenant, id string) (map[string]any, error)
	Cancel(ctx context.Context, tenant, id, reason string) (map[string]any, error)
	Refund(ctx context.Context, tenant, id string, body map[string]any) (map[string]any, error)
	DispatchEvent(ctx context.Context, tenant, id, eventType, courierRef string) (map[string]any, error)
}

type LiveOpsGateway interface {
	SetFlag(ctx context.Context, tenant, key string, enabled bool) (map[string]any, error)
}

type CatalogGateway interface {
	ListProducts(ctx context.Context, tenant string, q url.Values) (map[string]any, error)
	GetProduct(ctx context.Context, tenant, id string) (map[string]any, error)
}

type CRMGateway interface {
	ListTickets(ctx context.Context, tenant string) (map[string]any, error)
	GetTicket(ctx context.Context, tenant, id string) (map[string]any, error)
	EscalateTicket(ctx context.Context, tenant, id, reason string) (map[string]any, error)
	ResolveTicket(ctx context.Context, tenant, id, note string) (map[string]any, error)
}

type LedgerGateway interface {
	ListJournals(ctx context.Context, tenant string) (map[string]any, error)
}

type PromoGateway interface {
	ListCampaigns(ctx context.Context, tenant string, q url.Values) (map[string]any, error)
	GetCampaign(ctx context.Context, tenant, id string) (map[string]any, error)
	CreateCampaign(ctx context.Context, tenant string, body map[string]any) (map[string]any, error)
}

type PricingGateway interface {
	AdminList(ctx context.Context, tenant string) (map[string]any, error)
}

type InventoryGateway interface {
	ListWarehouses(ctx context.Context, tenant string) (map[string]any, error)
	GetWarehouse(ctx context.Context, tenant, id string) (map[string]any, error)
	ListStock(ctx context.Context, tenant, warehouseID string) (map[string]any, error)
}

type IdentityGateway interface {
	ListRoles(ctx context.Context, tenant string) (map[string]any, error)
}

type Deps struct {
	Orders    OrderGateway
	LiveOps   LiveOpsGateway
	Catalog   CatalogGateway
	CRM       CRMGateway
	Ledger    LedgerGateway
	Promo     PromoGateway
	Pricing   PricingGateway
	Inventory InventoryGateway
	Identity  IdentityGateway
}

func (d *Deps) Dashboard(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	out := map[string]any{
		"ordersLive": 0, "couriersOnDuty": 0, "openIncidents": 0, "sloBurn": 0.0,
		"delayedOrders": 0, "warehouseBacklog": 0,
	}
	if d.Orders != nil {
		list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"100"}})
		if err == nil {
			items := asMaps(list["items"])
			total := asInt(list["total"])
			if total == 0 {
				total = len(items)
			}
			out["ordersLive"] = total
			warehouse, courier, delayed := 0, 0, 0
			for _, o := range items {
				st := strings.ToLower(firstString(o, "status", "orderStatus"))
				switch st {
				case "warehouse_assigned", "picking", "packing", "confirmed":
					warehouse++
				case "courier_assigned", "out_for_delivery":
					courier++
				case "ready_for_dispatch":
					warehouse++
				}
				if st == "picking" || st == "packing" || st == "out_for_delivery" || st == "warehouse_assigned" {
					delayed++
				}
			}
			out["warehouseBacklog"] = warehouse
			out["couriersOnDuty"] = courier
			out["delayedOrders"] = delayed
			if total > 0 {
				out["sloBurn"] = float64(delayed) / float64(total)
			}
		}
	}
	if d.CRM != nil {
		tickets, err := d.CRM.ListTickets(ctx, tenant)
		if err == nil {
			if total := asInt(tickets["total"]); total > 0 {
				out["openIncidents"] = total
			} else {
				out["openIncidents"] = len(asMaps(tickets["items"]))
			}
		}
	}
	return out, nil
}

func (d *Deps) ListOrders(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Orders == nil {
		return nil, errors.New("orders gateway not configured")
	}
	return d.Orders.List(ctx, tenant, q)
}

func (d *Deps) GetOrder(ctx context.Context, tenant, id string) (map[string]any, error) {
	if tenant == "" || id == "" {
		return nil, ErrInvalid
	}
	if d.Orders == nil {
		return nil, errors.New("orders gateway not configured")
	}
	return d.Orders.Get(ctx, tenant, id)
}

func (d *Deps) OrderAction(ctx context.Context, tenant, id, action, reason, courierID string, refundMinor int64) (map[string]any, error) {
	if tenant == "" || id == "" || action == "" {
		return nil, ErrInvalid
	}
	if d.Orders == nil {
		return nil, errors.New("orders gateway not configured")
	}
	var (
		out map[string]any
		err error
	)
	switch strings.ToLower(strings.TrimSpace(action)) {
	case "cancel", "force_cancel":
		out, err = d.Orders.Cancel(ctx, tenant, id, reason)
	case "refund":
		body := map[string]any{"reason": reason}
		if refundMinor > 0 {
			body["amountMinor"] = refundMinor
		}
		out, err = d.Orders.Refund(ctx, tenant, id, body)
	case "reassign":
		out, err = d.Orders.DispatchEvent(ctx, tenant, id, "CourierAssigned", courierID)
	case "force_complete":
		out, err = d.Orders.DispatchEvent(ctx, tenant, id, "Delivered", courierID)
	case "replace":
		return map[string]any{
			"orderId": id, "action": action, "ok": false,
			"message": "line replacement is not supported for this order",
		}, nil
	default:
		return nil, fmt.Errorf("%w: unknown action %s", ErrInvalid, action)
	}
	if err != nil {
		return map[string]any{
			"orderId": id, "action": action, "ok": false, "message": err.Error(),
		}, nil
	}
	if out == nil {
		out = map[string]any{}
	}
	out["orderId"] = id
	out["action"] = action
	out["ok"] = true
	if _, ok := out["message"]; !ok {
		out["message"] = action + " applied"
	}
	return out, nil
}

func (d *Deps) ListCatalogProducts(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Catalog == nil {
		return nil, errors.New("catalog gateway not configured")
	}
	raw, err := d.Catalog.ListProducts(ctx, tenant, q)
	if err != nil {
		return nil, err
	}
	items := asMaps(raw["items"])
	mapped := make([]map[string]any, 0, len(items))
	brands := map[string]struct{}{}
	categories := map[string]struct{}{}
	for _, p := range items {
		row := mapCatalogProduct(p)
		mapped = append(mapped, row)
		if b := asString(row["brand"]); b != "" {
			brands[b] = struct{}{}
		}
		if c := asString(row["category"]); c != "" {
			categories[c] = struct{}{}
		}
	}
	return map[string]any{
		"items": mapped, "total": firstNonZero(asInt(raw["total"]), len(mapped)),
		"brands": setKeys(brands), "categories": setKeys(categories),
	}, nil
}

func (d *Deps) GetCatalogProduct(ctx context.Context, tenant, id string) (map[string]any, error) {
	if tenant == "" || id == "" {
		return nil, ErrInvalid
	}
	if d.Catalog == nil {
		return nil, errors.New("catalog gateway not configured")
	}
	raw, err := d.Catalog.GetProduct(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	return mapCatalogProduct(raw), nil
}

func (d *Deps) ListSupportTickets(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.CRM == nil {
		return nil, errors.New("crm gateway not configured")
	}
	raw, err := d.CRM.ListTickets(ctx, tenant)
	if err != nil {
		return nil, err
	}
	items := asMaps(raw["items"])
	tickets := make([]map[string]any, 0, len(items))
	complaints, refunds := 0, 0
	for _, t := range items {
		row := mapSupportTicket(t)
		tickets = append(tickets, row)
		if asString(row["category"]) == "complaint" {
			complaints++
		}
		if asString(row["category"]) == "refund" {
			refunds++
		}
	}
	return map[string]any{
		"tickets": tickets,
		"liveChat": map[string]any{"activeSessions": 0, "queued": 0, "avgWaitSec": 0, "agentsOnline": 0},
		"aiChatbot": map[string]any{
			"enabled": false, "containmentRatePct": 0, "handoffRatePct": 0, "topIntents": []any{},
		},
		"complaintCount": complaints,
		"openRefunds":    refunds,
	}, nil
}

func (d *Deps) GetSupportTicket(ctx context.Context, tenant, id string) (map[string]any, error) {
	if tenant == "" || id == "" {
		return nil, ErrInvalid
	}
	if d.CRM == nil {
		return nil, errors.New("crm gateway not configured")
	}
	raw, err := d.CRM.GetTicket(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	return mapSupportTicket(raw), nil
}

func (d *Deps) EscalateSupportTicket(ctx context.Context, tenant, id, reason string) (map[string]any, error) {
	if d.CRM == nil {
		return nil, errors.New("crm gateway not configured")
	}
	raw, err := d.CRM.EscalateTicket(ctx, tenant, id, reason)
	if err != nil {
		return nil, err
	}
	if t, ok := raw["ticket"].(map[string]any); ok {
		return mapSupportTicket(t), nil
	}
	return mapSupportTicket(raw), nil
}

func (d *Deps) ResolveSupportTicket(ctx context.Context, tenant, id, note string) (map[string]any, error) {
	if d.CRM == nil {
		return nil, errors.New("crm gateway not configured")
	}
	raw, err := d.CRM.ResolveTicket(ctx, tenant, id, note)
	if err != nil {
		return nil, err
	}
	return mapSupportTicket(raw), nil
}

func (d *Deps) FinanceSnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Ledger == nil {
		return nil, errors.New("ledger gateway not configured")
	}
	raw, err := d.Ledger.ListJournals(ctx, tenant)
	if err != nil {
		return nil, err
	}
	return financeFromJournals(raw), nil
}

func (d *Deps) KillSwitch(ctx context.Context, tenant, flag string, enabled bool) (map[string]any, error) {
	if flag == "" {
		return nil, ErrInvalid
	}
	if d.LiveOps == nil {
		return nil, errors.New("liveops gateway not configured")
	}
	res, err := d.LiveOps.SetFlag(ctx, tenant, flag, enabled)
	if err != nil {
		return nil, err
	}
	if res == nil {
		res = map[string]any{}
	}
	res["dualControl"] = true
	return res, nil
}

func mapCatalogProduct(p map[string]any) map[string]any {
	id := firstString(p, "id", "ID")
	sku := firstString(p, "sku", "skuCode", "SKUCode")
	name := firstString(p, "name", "title", "slug", "Slug")
	status := strings.ToLower(firstString(p, "status", "Status"))
	switch status {
	case "published", "active":
		status = "active"
	case "draft", "scheduled":
		status = "draft"
	case "archived", "deleted":
		status = "archived"
	case "":
		status = "draft"
	}
	return map[string]any{
		"id": id, "sku": sku, "name": name,
		"brand": firstString(p, "brand", "brandName"), "category": firstString(p, "category", "kind", "Kind"),
		"status": status, "price": asInt(p["price"]), "currency": firstNonEmpty(firstString(p, "currency"), "TRY"),
		"inventoryLinked": p["inventoryLinked"] == true, "variantCount": asInt(p["variantCount"]),
		"hasBundle": p["hasBundle"] == true,
	}
}

func mapSupportTicket(t map[string]any) map[string]any {
	status := strings.ToLower(firstString(t, "status", "Status"))
	if status == "in_progress" || status == "waiting_customer" || status == "reopened" {
		status = "pending"
	}
	priority := strings.ToLower(firstString(t, "priority", "Priority"))
	if priority == "normal" {
		priority = "medium"
	}
	category := strings.ToLower(firstString(t, "category", "Category"))
	switch category {
	case "payment":
		category = "refund"
	case "fraud", "technical", "legal":
		category = "other"
	case "":
		category = "other"
	}
	assignee := firstString(t, "assignee", "assigneeId", "AssigneeID")
	var assigneeAny any
	if assignee != "" {
		assigneeAny = assignee
	}
	orderID := firstString(t, "orderId", "order_id")
	var orderAny any
	if orderID != "" {
		orderAny = orderID
	}
	return map[string]any{
		"id": firstString(t, "id", "ID"),
		"subject": firstString(t, "subject", "Subject"),
		"status": status, "priority": firstNonEmpty(priority, "medium"),
		"category": category,
		"customerId": firstString(t, "customerId", "CustomerID"),
		"customerName": firstNonEmpty(firstString(t, "customerName"), firstString(t, "customerId", "CustomerID")),
		"orderId": orderAny, "assignee": assigneeAny,
		"createdAt": firstString(t, "createdAt", "CreatedAt"),
		"updatedAt": firstString(t, "updatedAt", "UpdatedAt"),
		"messages": []any{}, "refund": nil, "qc": []any{},
		"escalated": status == "escalated" || t["slaBreached"] == true,
		"aiSuggestedReply": nil,
	}
}

func financeFromJournals(raw map[string]any) map[string]any {
	items := asMaps(raw["items"])
	var debit int64
	currency := "TRY"
	payments := make([]map[string]any, 0, len(items))
	for _, j := range items {
		amt := int64(asInt(j["debitTotal"]))
		debit += amt
		if c := firstString(j, "currency"); c != "" {
			currency = c
		}
		payments = append(payments, map[string]any{
			"id": firstString(j, "id"), "method": "ledger",
			"amountMinor": amt, "currency": currency,
			"status": "captured", "at": firstString(j, "postedAt"),
		})
	}
	empty := []any{}
	return map[string]any{
		"kpis": []map[string]any{
			{"id": "posted", "label": "Posted journals", "valueMinor": debit, "currency": currency, "deltaPct": 0},
			{"id": "count", "label": "Journal count", "valueMinor": int64(len(items)) * 100, "currency": currency, "deltaPct": 0},
		},
		"revenue": empty, "refunds": empty, "taxes": empty, "invoices": empty,
		"payments": payments, "payouts": empty, "courierSettlements": empty, "supplierPayments": empty,
		"profit": map[string]any{
			"gmvMinor": debit, "cogsMinor": 0, "deliveryCostMinor": 0, "promoCostMinor": 0,
			"contributionMinor": debit, "currency": currency,
		},
		"reports": []map[string]any{
			{"id": "journals", "title": "Ledger journals", "href": "/v1/admin/finance/journals", "description": "Posted double-entry journals for this tenant"},
		},
		"journals": items,
	}
}

func asMaps(v any) []map[string]any {
	switch arr := v.(type) {
	case []map[string]any:
		return arr
	case []any:
		out := make([]map[string]any, 0, len(arr))
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out
	default:
		return nil
	}
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func asString(v any) string {
	s, _ := v.(string)
	return s
}

func asInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

func firstNonZero(n, fallback int) int {
	if n > 0 {
		return n
	}
	return fallback
}

func firstNonEmpty(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}

func setKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
