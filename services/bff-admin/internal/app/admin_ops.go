package app

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func (d *Deps) ListCampaigns(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	raw, err := d.Promo.ListCampaigns(ctx, tenant, q)
	if err != nil {
		return nil, err
	}
	items := asMaps(raw["items"])
	mapped := make([]map[string]any, 0, len(items))
	for _, c := range items {
		mapped = append(mapped, mapCampaign(c))
	}
	return map[string]any{
		"items": mapped, "page": 1, "pageSize": 50, "total": len(mapped), "hasMore": false,
	}, nil
}

func (d *Deps) GetCampaign(ctx context.Context, tenant, id string) (map[string]any, error) {
	if tenant == "" || id == "" {
		return nil, ErrInvalid
	}
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	raw, err := d.Promo.GetCampaign(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	return mapCampaign(raw), nil
}

func (d *Deps) CreateCampaign(ctx context.Context, tenant string, name, description string, startsAt, endsAt any) (map[string]any, error) {
	if tenant == "" || name == "" {
		return nil, ErrInvalid
	}
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	raw, err := d.Promo.CreateCampaign(ctx, tenant, map[string]any{
		"name": name, "description": description, "startsAt": startsAt, "endsAt": endsAt,
	})
	if err != nil {
		return nil, err
	}
	return mapCampaign(raw), nil
}

func mapCampaign(c map[string]any) map[string]any {
	st := strings.ToLower(firstString(c, "status"))
	if st == "expired" {
		st = "ended"
	}
	if st == "" {
		st = "draft"
	}
	return map[string]any{
		"id": firstString(c, "id"), "name": firstString(c, "name"),
		"type": "coupon", "status": st, "cityIds": []any{},
		"startsAt": c["startsAt"], "endsAt": c["endsAt"],
		"budgetMinor": 0, "spentMinor": 0, "currency": "TRY",
		"audience": nil, "coupon": nil, "bundle": nil, "flashSale": nil, "personalized": nil,
		"createdAt": firstString(c, "createdAt"), "updatedAt": firstString(c, "updatedAt"),
		"description": firstString(c, "description"),
	}
}

func (d *Deps) ListPricing(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Pricing == nil {
		return nil, fmt.Errorf("pricing gateway not configured")
	}
	raw, err := d.Pricing.AdminList(ctx, tenant)
	if err != nil {
		return nil, err
	}
	entries := asMaps(raw["entries"])
	mapped := make([]map[string]any, 0, len(entries))
	for _, e := range entries {
		scope := strings.ToLower(firstString(e, "scope"))
		kind := "base"
		switch scope {
		case "region", "regional":
			kind = "regional"
		case "warehouse":
			kind = "warehouse"
		case "dynamic":
			kind = "dynamic"
		}
		mapped = append(mapped, map[string]any{
			"id": firstString(e, "id"), "name": firstNonEmpty(firstString(e, "name"), firstString(e, "variantId")),
			"kind": kind, "status": "active",
			"skuId": firstString(e, "variantId"), "categoryId": nil, "cityId": nil,
			"warehouseId": e["scopeId"], "basePriceMinor": asInt(e["amountMinor"]),
			"overridePriceMinor": nil, "adjustmentPct": nil,
			"currency": firstNonEmpty(firstString(e, "currency"), "TRY"),
			"priority": 100, "startsAt": e["validFrom"], "endsAt": e["validTo"],
			"competitorRef": nil, "aiConfidence": nil, "notes": "",
			"updatedAt": firstString(e, "updatedAt"),
		})
	}
	return map[string]any{
		"items": mapped, "page": 1, "pageSize": 100, "total": len(mapped), "hasMore": false,
	}, nil
}

func (d *Deps) InventorySnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Inventory == nil {
		return nil, fmt.Errorf("inventory gateway not configured")
	}
	whRaw, err := d.Inventory.ListWarehouses(ctx, tenant)
	if err != nil {
		return nil, err
	}
	warehouses := asMaps(whRaw["items"])
	stock := make([]map[string]any, 0)
	onHand, reserved, below := 0, 0, 0
	for i, wh := range warehouses {
		if i >= 8 {
			break
		}
		id := firstString(wh, "id", "ID")
		code := firstString(wh, "code", "Code")
		if id == "" {
			continue
		}
		st, err := d.Inventory.ListStock(ctx, tenant, id)
		if err != nil {
			continue
		}
		for _, row := range asMaps(st["items"]) {
			oh := asInt(row["OnHand"])
			if oh == 0 {
				oh = asInt(row["onHand"])
			}
			rs := asInt(row["Reserved"])
			if rs == 0 {
				rs = asInt(row["reserved"])
			}
			avail := asInt(row["Available"])
			if avail == 0 {
				avail = oh - rs
				if avail < 0 {
					avail = 0
				}
			}
			safety := asInt(row["SafetyMin"])
			if safety == 0 {
				safety = asInt(row["safetyMin"])
			}
			kind := "normal"
			if asInt(row["Blocked"]) > 0 || asInt(row["blocked"]) > 0 {
				kind = "damaged"
			}
			if oh < safety {
				below++
			}
			onHand += oh
			reserved += rs
			stock = append(stock, map[string]any{
				"id": firstString(row, "id", "ID"),
				"sku": firstString(row, "SKUCode", "skuCode", "sku"),
				"productName": firstString(row, "SKUCode", "skuCode"),
				"warehouseCode": code, "onHand": oh, "reserved": rs, "available": avail,
				"safetyStock": safety, "forecast7d": 0, "kind": kind,
				"lastCountedAt": firstString(row, "UpdatedAt", "updatedAt"),
			})
		}
	}
	empty := []any{}
	return map[string]any{
		"cityId": nil, "generatedAt": time.Now().UTC().Format(time.RFC3339),
		"stock": stock, "transfers": empty, "cycleCounts": empty, "adjustments": empty,
		"totals": map[string]any{
			"skus": len(stock), "unitsOnHand": onHand, "reserved": reserved,
			"damaged": 0, "expired": 0, "belowSafety": below,
		},
	}, nil
}

func (d *Deps) ListWarehouses(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	if d.Inventory == nil {
		return nil, fmt.Errorf("inventory gateway not configured")
	}
	raw, err := d.Inventory.ListWarehouses(ctx, tenant)
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0)
	for _, wh := range asMaps(raw["items"]) {
		items = append(items, mapWarehouse(wh))
	}
	return map[string]any{"items": items, "total": firstNonZero(asInt(raw["total"]), len(items)), "generatedAt": time.Now().UTC().Format(time.RFC3339)}, nil
}

func (d *Deps) GetWarehouse(ctx context.Context, tenant, id string) (map[string]any, error) {
	if tenant == "" || id == "" {
		return nil, ErrInvalid
	}
	if d.Inventory == nil {
		return nil, fmt.Errorf("inventory gateway not configured")
	}
	raw, err := d.Inventory.GetWarehouse(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	row := mapWarehouse(raw)
	row["address"] = ""
	row["managerName"] = ""
	row["openedAt"] = firstString(raw, "CreatedAt", "createdAt")
	row["kpis"] = map[string]any{
		"capacityPct": 0, "skuCount": 0, "unitsOnHand": 0, "pickSlaPct": 0,
		"packSlaPct": 0, "dispatchSlaPct": 0, "avgPickMinutes": 0, "avgPackMinutes": 0,
		"openTransfers": 0, "stockAlerts": 0,
	}
	row["inventorySummary"] = []any{}
	row["transfers"] = []any{}
	row["audits"] = []any{}
	row["stockAlerts"] = []any{}
	row["aiOptimization"] = []any{}
	return row, nil
}

func mapWarehouse(wh map[string]any) map[string]any {
	st := strings.ToLower(firstString(wh, "status", "Status"))
	switch st {
	case "active":
		st = "open"
	case "inactive", "closed":
		st = "closed"
	case "maintenance":
		st = "maintenance"
	case "":
		st = "open"
	}
	return map[string]any{
		"id": firstString(wh, "id", "ID"), "code": firstString(wh, "code", "Code"),
		"name": firstString(wh, "name", "Name"), "cityId": firstString(wh, "RegionID", "regionId"),
		"district": "", "status": st, "capacityPct": 0, "skuCount": 0,
		"openOrders": 0, "pickSlaPct": 0, "stockAlerts": 0,
	}
}

func mapLiveStatus(st string) string {
	switch strings.ToLower(st) {
	case "picking", "packing", "warehouse_assigned", "confirmed":
		return "picking"
	case "ready_for_dispatch":
		return "ready"
	case "courier_assigned":
		return "assigned"
	case "out_for_delivery":
		return "en_route"
	case "delivered":
		return "delivered"
	case "cancelled":
		return "cancelled"
	case "failed":
		return "failed"
	default:
		return "created"
	}
}

func (d *Deps) AuditSnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	events := []map[string]any{}
	if d.Orders != nil {
		list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"30"}})
		if err == nil {
			for _, o := range asMaps(list["items"]) {
				events = append(events, map[string]any{
					"id": "ord-" + firstString(o, "id"),
					"who": "system",
					"when": firstString(o, "updatedAt", "UpdatedAt", "createdAt"),
					"where": "order-service", "device": "api",
					"action": "orders.status", "resource": firstString(o, "id"),
					"oldValue": "", "newValue": firstString(o, "status"),
					"ip": "", "sessionId": "",
				})
			}
		}
	}
	if d.CRM != nil {
		tickets, err := d.CRM.ListTickets(ctx, tenant)
		if err == nil {
			for _, t := range asMaps(tickets["items"]) {
				events = append(events, map[string]any{
					"id": "tkt-" + firstString(t, "id"),
					"who": firstString(t, "assigneeId", "assignee"),
					"when": firstString(t, "updatedAt", "createdAt"),
					"where": "crm-service", "device": "api",
					"action": "support.ticket", "resource": firstString(t, "id"),
					"oldValue": "", "newValue": firstString(t, "status"),
					"ip": "", "sessionId": "",
				})
			}
		}
	}
	return map[string]any{"generatedAt": time.Now().UTC().Format(time.RFC3339), "events": events}, nil
}

func (d *Deps) RbacSnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	roles := []map[string]any{}
	if d.Identity != nil {
		raw, err := d.Identity.ListRoles(ctx, tenant)
		if err == nil {
			for _, r := range asMaps(raw["items"]) {
				key := firstString(r, "name")
				roles = append(roles, map[string]any{
					"id": firstString(r, "id"), "key": key, "label": strings.ReplaceAll(key, "_", " "),
					"members": 0, "description": firstString(r, "description"),
				})
			}
		}
	}
	if len(roles) == 0 {
		for _, name := range []string{"customer", "courier", "picker", "admin", "super_admin", "support_agent", "finance_analyst", "city_ops", "supplier"} {
			roles = append(roles, map[string]any{
				"id": "role-" + name, "key": name, "label": strings.ReplaceAll(name, "_", " "),
				"members": 0, "description": "platform " + name,
			})
		}
	}
	matrixRoles := []string{"city_ops", "support_agent", "admin", "super_admin", "finance_analyst"}
	perms := []string{"orders:read", "orders:cancel", "orders:force_complete", "finance:payout:approve", "system:flags", "rbac:write", "audit:read"}
	grants := map[string][]string{
		"city_ops": {"orders:read", "audit:read"},
		"support_agent": {"orders:read", "orders:cancel"},
		"admin": {"orders:read", "orders:cancel", "orders:force_complete", "audit:read", "rbac:write"},
		"super_admin": perms,
		"finance_analyst": {"orders:read", "finance:payout:approve", "audit:read"},
	}
	matrix := []map[string]any{}
	for _, role := range matrixRoles {
		allowed := map[string]struct{}{}
		for _, p := range grants[role] {
			allowed[p] = struct{}{}
		}
		for _, p := range perms {
			_, ok := allowed[p]
			matrix = append(matrix, map[string]any{"role": role, "permission": p, "granted": ok})
		}
	}
	return map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"departments": []any{}, "roles": roles, "matrix": matrix,
		"matrixPermissions": perms, "matrixRoles": matrixRoles,
		"customPermissions": []any{}, "temporaryGrants": []any{}, "approvals": []any{},
	}, nil
}

func (d *Deps) ListCustomers(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	needle := strings.TrimSpace(q.Get("q"))
	if d.Profile != nil {
		raw, err := d.Profile.Search(ctx, tenant, needle, 100)
		if err != nil {
			return nil, err
		}
		items := make([]map[string]any, 0)
		for _, p := range asMaps(raw["items"]) {
			pid := firstString(p, "id", "principalId")
			phone := maskPhone(firstString(p, "phone", "msisdn"))
			items = append(items, map[string]any{
				"id": pid, "principalId": firstString(p, "principalId"),
				"name": firstNonEmpty(firstString(p, "fullName", "displayName"), pid),
				"email": firstString(p, "email"), "phone": phone,
				"cityId": firstString(p, "city"), "status": firstString(p, "status"),
				"tenant": firstString(p, "tenantId"),
				"segment": "", "orderCount": 0, "lifetimeValueMinor": 0,
				"currency": "TRY", "riskScore": 0, "fraudScore": 0, "loyaltyTier": "",
				"walletBalanceMinor": 0,
				"createdAt": firstString(p, "createdAt"), "lastOrderAt": firstString(p, "updatedAt"),
			})
		}
		if d.Orders != nil {
			if list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"100"}}); err == nil {
				counts := map[string]int{}
				last := map[string]string{}
				for _, o := range asMaps(list["items"]) {
					cid := firstString(o, "customerId", "CustomerID", "principalId")
					counts[cid]++
					last[cid] = firstString(o, "createdAt", "updatedAt")
				}
				for _, it := range items {
					id := firstString(it, "id", "principalId")
					it["orderCount"] = counts[id]
					if last[id] != "" {
						it["lastOrderAt"] = last[id]
					}
				}
			}
		}
		return map[string]any{"items": items, "total": len(items), "page": 1, "pageSize": len(items), "hasMore": false}, nil
	}
	if d.Orders == nil {
		return nil, fmt.Errorf("customer directory gateway not configured")
	}
	list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"100"}})
	if err != nil {
		return nil, err
	}
	type agg struct {
		id, last string
		count    int
		total    int
	}
	seen := map[string]*agg{}
	order := []string{}
	nl := strings.ToLower(needle)
	for _, o := range asMaps(list["items"]) {
		cid := firstString(o, "customerId", "CustomerID", "principalId")
		if cid == "" {
			continue
		}
		if nl != "" && !strings.Contains(strings.ToLower(cid), nl) {
			continue
		}
		a, ok := seen[cid]
		if !ok {
			a = &agg{id: cid}
			seen[cid] = a
			order = append(order, cid)
		}
		a.count++
		a.total += asInt(o["totalMinor"])
		a.last = firstString(o, "createdAt", "updatedAt")
	}
	items := make([]map[string]any, 0, len(order))
	for _, id := range order {
		a := seen[id]
		items = append(items, map[string]any{
			"id": a.id, "name": a.id, "email": "", "phone": "", "cityId": "",
			"segment": "loyal", "orderCount": a.count, "lifetimeValueMinor": a.total,
			"currency": "TRY", "riskScore": 0, "fraudScore": 0, "loyaltyTier": "",
			"walletBalanceMinor": 0, "createdAt": a.last, "lastOrderAt": a.last,
		})
	}
	return map[string]any{"items": items, "total": len(items), "page": 1, "pageSize": len(items), "hasMore": false}, nil
}

func maskPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if len(phone) < 4 {
		return phone
	}
	return strings.Repeat("*", len(phone)-4) + phone[len(phone)-4:]
}

func (d *Deps) GetCustomer(ctx context.Context, tenant, id string) (map[string]any, error) {
	list, err := d.ListCustomers(ctx, tenant, url.Values{})
	if err != nil {
		return nil, err
	}
	for _, c := range asMaps(list["items"]) {
		if firstString(c, "id") == id {
			c["addresses"] = []any{}
			c["recentOrders"] = []any{}
			c["walletTxns"] = []any{}
			c["loyalty"] = map[string]any{"tier": "", "points": 0, "pointsToNextTier": 0}
			c["coupons"] = []any{}
			c["supportHistory"] = []any{}
			c["notes"] = []any{}
			return c, nil
		}
	}
	return nil, fmt.Errorf("%w: customer not found", ErrInvalid)
}

func (d *Deps) CustomerAdjustment(_ context.Context, customerID, kind, note string) (map[string]any, error) {
	return map[string]any{
		"customerId": customerID, "ok": false,
		"message": kind + " adjustment is not supported without wallet/loyalty write APIs: " + note,
	}, nil
}

func (d *Deps) FinanceMutation(ctx context.Context, kind, id string) (map[string]any, error) {
	if d.Settlement == nil {
		return map[string]any{
			"ok": false, "id": id, "kind": kind,
			"message": kind + " requires settlement-service and is not silently mocked",
			"provider": "unavailable",
		}, nil
	}
	tenant := ""
	var (
		out map[string]any
		err error
	)
	switch kind {
	case "payout_approve":
		out, err = d.Settlement.Approve(ctx, tenant, id)
	case "courier_settle":
		out, err = d.Settlement.Execute(ctx, tenant, id)
	default:
		return map[string]any{
			"ok": false, "id": id, "kind": kind,
			"message": kind + " is not a settlement batch mutation",
			"provider": "internal",
		}, nil
	}
	if err != nil {
		return map[string]any{
			"ok": false, "id": id, "kind": kind, "message": err.Error(), "provider": "settlement",
		}, err
	}
	status := firstString(out, "status", "Status")
	return map[string]any{
		"ok": true, "id": id, "kind": kind, "status": status, "batch": out,
	}, nil
}

func (d *Deps) FinanceMutationTenant(ctx context.Context, tenant, kind, id string) (map[string]any, error) {
	if d.Settlement == nil {
		return d.FinanceMutation(ctx, kind, id)
	}
	var (
		out map[string]any
		err error
	)
	switch kind {
	case "payout_approve":
		out, err = d.Settlement.Approve(ctx, tenant, id)
	case "courier_settle":
		out, err = d.Settlement.Execute(ctx, tenant, id)
	case "refund_approve":
		if d.Orders != nil {
			out, err = d.Orders.Refund(ctx, tenant, id, map[string]any{"reason": "admin_refund_approve"})
			break
		}
		return map[string]any{"ok": false, "id": id, "kind": kind, "message": "order refund gateway not configured"}, nil
	default:
		return d.FinanceMutation(ctx, kind, id)
	}
	if err != nil {
		return map[string]any{"ok": false, "id": id, "kind": kind, "message": err.Error()}, err
	}
	return map[string]any{"ok": true, "id": id, "kind": kind, "status": firstString(out, "status"), "batch": out}, nil
}
