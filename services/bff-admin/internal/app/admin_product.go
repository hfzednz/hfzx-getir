package app

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func (d *Deps) ListCoupons(ctx context.Context, tenant string, q url.Values) (map[string]any, error) {
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	return d.Promo.ListCoupons(ctx, tenant, q)
}

func (d *Deps) GetCoupon(ctx context.Context, tenant, code string) (map[string]any, error) {
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	return d.Promo.GetCoupon(ctx, tenant, code)
}

func (d *Deps) CreateCoupon(ctx context.Context, tenant string, body map[string]any) (map[string]any, error) {
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	return d.Promo.CreateCoupon(ctx, tenant, body)
}

func (d *Deps) UpdateCoupon(ctx context.Context, tenant, code string, body map[string]any) (map[string]any, error) {
	if d.Promo == nil {
		return nil, fmt.Errorf("promo gateway not configured")
	}
	return d.Promo.UpdateCoupon(ctx, tenant, code, body)
}

func (d *Deps) NotificationsSnapshot(ctx context.Context, tenant, principalID string) (map[string]any, error) {
	alerts := []map[string]any{}
	if d.Notify != nil && principalID != "" {
		if inbox, err := d.Notify.Inbox(ctx, tenant, principalID); err == nil {
			for _, it := range asMaps(inbox["items"]) {
				alerts = append(alerts, map[string]any{
					"id": firstString(it, "id"),
					"category": firstNonEmpty(firstString(it, "category", "channel"), "operational"),
					"title": firstString(it, "title", "subject"),
					"body": firstString(it, "body", "preview", "message"),
					"severity": firstNonEmpty(firstString(it, "severity"), "info"),
					"read": it["read"] == true,
					"createdAt": firstString(it, "createdAt", "sentAt"),
				})
			}
		}
	}
	if d.CRM != nil {
		if tickets, err := d.CRM.ListTickets(ctx, tenant); err == nil {
			for _, t := range asMaps(tickets["items"]) {
				st := strings.ToLower(firstString(t, "status", "state"))
				if st != "escalated" && st != "open" && firstString(t, "priority") == "" {
					continue
				}
				alerts = append(alerts, map[string]any{
					"id": "ticket-" + firstString(t, "id"),
					"category": "operational",
					"title": "Support " + firstNonEmpty(st, "ticket"),
					"body": firstString(t, "subject", "title"),
					"severity": "warning",
					"read": false,
					"createdAt": firstString(t, "updatedAt", "createdAt"),
				})
			}
		}
	}
	if d.Orders != nil {
		if list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"50"}}); err == nil {
			for _, o := range asMaps(list["items"]) {
				st := strings.ToLower(firstString(o, "status"))
				if st == "cancelled" || st == "failed" || st == "payment_failed" {
					alerts = append(alerts, map[string]any{
						"id": "order-" + firstString(o, "id"),
						"category": "financial",
						"title": "Order " + st,
						"body": firstString(o, "id"),
						"severity": "danger",
						"read": false,
						"createdAt": firstString(o, "updatedAt", "createdAt"),
					})
				}
			}
		}
	}
	unread := 0
	for _, a := range alerts {
		if a["read"] != true {
			unread++
		}
	}
	return map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"alerts":      alerts,
		"unreadCount": unread,
	}, nil
}

func (d *Deps) MarkNotificationRead(ctx context.Context, tenant, id string) (map[string]any, error) {
	if d.Notify == nil {
		return map[string]any{"ok": false, "id": id, "message": "notification-service not configured"}, nil
	}
	out, err := d.Notify.MarkRead(ctx, tenant, id)
	if err != nil {
		return nil, err
	}
	return map[string]any{"ok": true, "id": id, "item": out}, nil
}

func (d *Deps) MarkAllNotificationsRead(ctx context.Context, tenant, principalID string) (map[string]any, error) {
	snap, err := d.NotificationsSnapshot(ctx, tenant, principalID)
	if err != nil {
		return nil, err
	}
	n := 0
	for _, a := range asMaps(snap["alerts"]) {
		if a["read"] == true {
			continue
		}
		id := firstString(a, "id")
		if strings.HasPrefix(id, "ticket-") || strings.HasPrefix(id, "order-") {
			continue
		}
		if _, err := d.MarkNotificationRead(ctx, tenant, id); err == nil {
			n++
		}
	}
	return map[string]any{"ok": true, "marked": n}, nil
}

func (d *Deps) ReportsCatalog(ctx context.Context, tenant string, from, to string) (map[string]any, error) {
	orders := []map[string]any{}
	if d.Orders != nil {
		if list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"200"}}); err == nil {
			orders = asMaps(list["items"])
		}
	}
	counts := map[string]int{"total": len(orders), "completed": 0, "cancelled": 0, "refunded": 0, "paid": 0, "failed": 0}
	gmv := 0
	for _, o := range orders {
		st := strings.ToLower(firstString(o, "status", "orderStatus", "paymentStatus"))
		counts["total"] = len(orders)
		switch {
		case strings.Contains(st, "deliver") || st == "completed":
			counts["completed"]++
		case strings.Contains(st, "cancel"):
			counts["cancelled"]++
		case strings.Contains(st, "refund"):
			counts["refunded"]++
		case strings.Contains(st, "fail"):
			counts["failed"]++
		case strings.Contains(st, "paid") || strings.Contains(st, "captured"):
			counts["paid"]++
		}
		gmv += asInt(o["totalMinor"])
	}
	journalCount := 0
	if d.Ledger != nil {
		if j, err := d.Ledger.ListJournals(ctx, tenant); err == nil {
			journalCount = len(asMaps(j["items"]))
		}
	}
	ticketCount := 0
	if d.CRM != nil {
		if t, err := d.CRM.ListTickets(ctx, tenant); err == nil {
			ticketCount = len(asMaps(t["items"]))
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	row := map[string]any{
		"from": from, "to": to, "orders": counts["total"], "gmv_try": float64(gmv) / 100.0,
		"completed": counts["completed"], "cancelled": counts["cancelled"], "refunded": counts["refunded"],
		"paid": counts["paid"], "failed": counts["failed"], "journals": journalCount, "tickets": ticketCount,
	}
	return map[string]any{
		"generatedAt": now,
		"scope":       map[string]any{"tenant": tenant, "from": from, "to": to},
		"templates": []map[string]any{
			{"id": "rep_orders", "domain": "orders", "name": "Orders summary",
				"description": "Order volume, GMV, cancel & refund counts from live orders",
				"columns": []string{"from", "to", "orders", "gmv_try", "completed", "cancelled", "refunded"},
				"sampleRows": []map[string]any{row}},
			{"id": "rep_payments", "domain": "finance", "name": "Payment status",
				"description": "Paid vs failed vs refunded from order payment status",
				"columns": []string{"paid", "failed", "refunded", "journals"},
				"sampleRows": []map[string]any{row}},
			{"id": "rep_support", "domain": "crm", "name": "Support volume",
				"description": "Open ticket volume from CRM",
				"columns": []string{"tickets"},
				"sampleRows": []map[string]any{row}},
		},
	}, nil
}

func (d *Deps) SystemSnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	flags := []map[string]any{}
	if d.LiveOps != nil {
		if raw, err := d.LiveOps.ListFlags(ctx, tenant); err == nil {
			for _, f := range asMaps(raw["items"]) {
				key := firstString(f, "key", "Key")
				flags = append(flags, map[string]any{
					"id": key, "key": key,
					"enabled": f["enabled"] == true || f["Enabled"] == true,
					"killSwitch": f["emergencyOff"] == true || f["EmergencyOff"] == true,
					"description": firstString(f, "description", "Description"),
					"updatedAt": firstString(f, "updatedAt", "UpdatedAt"),
				})
			}
		}
	}
	settings := []map[string]any{
		{"id": "env", "key": "runtime.environment", "value": firstNonEmpty(os.Getenv("NEXORA_ENV"), "local"), "category": "app", "description": "Declared environment"},
		{"id": "locale", "key": "locale.default", "value": "tr-TR", "category": "locale", "description": "Default locale"},
		{"id": "currency", "key": "currency.default", "value": "TRY", "category": "currency", "description": "Default currency"},
	}
	return map[string]any{
		"generatedAt": time.Now().UTC().Format(time.RFC3339),
		"locales":     []string{"tr-TR", "en-US"},
		"currencies":  []string{"TRY"},
		"settings":    settings,
		"flags":       flags,
		"templates":   []any{},
		"zones":       []any{},
	}, nil
}

func (d *Deps) MonitoringSnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	client := d.HTTP
	if client == nil {
		client = &http.Client{Timeout: 2 * time.Second}
	}
	services := []map[string]any{}
	for _, t := range d.Health {
		status, latency := "down", 0
		if strings.TrimSpace(t.URL) != "" {
			start := time.Now()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(t.URL, "/")+"/health", nil)
			if err != nil {
				status = "down"
			} else {
				if tenant != "" {
					req.Header.Set("X-Tenant-Id", tenant)
				}
				res, err := client.Do(req)
				latency = int(time.Since(start).Milliseconds())
				if err != nil {
					status = "down"
				} else {
					_ = res.Body.Close()
					if res.StatusCode >= 200 && res.StatusCode < 300 {
						status = "healthy"
					} else {
						status = "degraded"
					}
				}
			}
		}
		services = append(services, map[string]any{
			"id": t.Name, "name": t.Name, "status": status, "latencyMs": latency,
		})
	}
	queues := []map[string]any{}
	if d.Orders != nil {
		if list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"50"}}); err == nil {
			depth := len(asMaps(list["items"]))
			st := "ok"
			if depth > 40 {
				st = "warn"
			}
			queues = append(queues, map[string]any{"id": "orders", "name": "order processing", "depth": depth, "lagSec": 0, "status": st})
		}
	}
	if d.CRM != nil {
		if list, err := d.CRM.ListTickets(ctx, tenant); err == nil {
			depth := len(asMaps(list["items"]))
			st := "ok"
			if depth > 20 {
				st = "warn"
			}
			queues = append(queues, map[string]any{"id": "support", "name": "support backlog", "depth": depth, "lagSec": 0, "status": st})
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	return map[string]any{
		"generatedAt": now, "services": services,
		"apiLatency": []any{}, "serverLoad": []any{}, "cpu": []any{}, "memory": []any{},
		"storagePct": 0, "dbConnections": 0, "dbSlowQueries": 0,
		"queues": queues,
		"websocket": map[string]any{"status": "disconnected", "clients": 0, "msgPerSec": 0},
	}, nil
}

func (d *Deps) AICommandSnapshot(ctx context.Context, tenant, cityID string) (map[string]any, error) {
	provider := "unavailable"
	stats := map[string]any{}
	if d.AI != nil {
		if raw, err := d.AI.AdminStats(ctx, tenant); err == nil {
			provider = "ai-platform"
			stats = raw
		}
	}
	insights := []map[string]any{}
	risks := []map[string]any{}
	if d.Orders != nil {
		if list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"50"}}); err == nil {
			failed, delayed := 0, 0
			for _, o := range asMaps(list["items"]) {
				st := strings.ToLower(firstString(o, "status"))
				if strings.Contains(st, "fail") {
					failed++
				}
				if strings.Contains(st, "delay") || st == "picking" {
					delayed++
				}
			}
			insights = append(insights, map[string]any{
				"id": "local-orders", "title": "Order processing summary",
				"detail": fmt.Sprintf("local analysis of %d recent orders: %d failed, %d in-progress", len(asMaps(list["items"])), failed, delayed),
				"confidence": 1, "source": "local_analysis",
			})
			if failed > 0 {
				risks = append(risks, map[string]any{
					"id": "local-fail", "area": "orders", "severity": "medium",
					"summary": fmt.Sprintf("%d recent orders in failed status (local count, not a fraud model)", failed),
				})
			}
		}
	}
	return map[string]any{
		"cityId": cityID, "generatedAt": time.Now().UTC().Format(time.RFC3339),
		"provider": provider, "providerUnavailable": provider == "unavailable",
		"upstream": stats,
		"kpis": []map[string]any{
			{"id": "provider", "label": "AI provider", "value": boolToNum(provider != "unavailable"), "unit": "count", "tone": "neutral"},
		},
		"demandForecast": []any{}, "inventoryForecast": []any{}, "fraudAlerts": []any{},
		"deliveryOptSeries": []any{}, "recommendationCtr": []any{}, "pricingRecs": []any{},
		"campaignOpts": []any{}, "segments": []any{}, "risks": risks, "opsInsights": insights,
		"disclaimer": "Operational facts are counted from live domain APIs. External inference is not fabricated.",
	}, nil
}

func boolToNum(v bool) int {
	if v {
		return 1
	}
	return 0
}

func (d *Deps) LoyaltySnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	rewards := []map[string]any{}
	if d.Loyalty != nil {
		if raw, err := d.Loyalty.ListRewards(ctx, tenant); err == nil {
			for _, r := range asMaps(raw["items"]) {
				rewards = append(rewards, map[string]any{
					"id": firstString(r, "id"),
					"title": firstString(r, "name", "title"),
					"pointsCost": asInt(r["pointsCost"]),
					"stock": asInt(r["stock"]),
					"redeemed": asInt(r["redeemedCount"]),
					"active": r["active"] != false,
				})
			}
		} else {
			return nil, err
		}
	}
	return map[string]any{
		"totalMembers": 0, "pointsIssued": 0, "pointsRedeemed": 0,
		"levels": []any{}, "rewards": rewards, "cashback": []any{},
		"referral": map[string]any{"id": "", "referrerBonusPoints": 0, "refereeBonusPoints": 0, "conversions": 0, "active": false},
		"vipBenefits": []any{}, "achievements": []any{}, "challenges": []any{},
		"source": "loyalty-service",
	}, nil
}

func (d *Deps) LiveSnapshot(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	dash, err := d.Dashboard(ctx, tenant)
	if err != nil {
		return nil, err
	}
	stream := []map[string]any{}
	if d.Orders != nil {
		list, err := d.Orders.List(ctx, tenant, url.Values{"limit": {"50"}})
		if err == nil {
			for _, o := range asMaps(list["items"]) {
				st := mapLiveStatus(firstString(o, "status", "orderStatus"))
				stream = append(stream, map[string]any{
					"id": firstString(o, "id"), "orderId": firstString(o, "id"),
					"status": st, "customerName": firstString(o, "customerId", "CustomerID"),
					"warehouseCode": "", "zone": "", "etaMinutes": nil, "delayMinutes": 0,
					"amountMinor": asInt(o["totalMinor"]), "currency": firstNonEmpty(firstString(o, "currency"), "TRY"),
					"updatedAt": firstString(o, "updatedAt", "UpdatedAt"),
				})
			}
		}
	}
	couriers := []any{}
	if d.Tracking != nil {
		if nearby, err := d.Tracking.Nearby(ctx, tenant, 41.01, 28.97); err == nil {
			for _, loc := range asMaps(nearby["items"]) {
				couriers = append(couriers, map[string]any{
					"id": firstString(loc, "courierId"),
					"lat": loc["lat"], "lng": loc["lon"],
					"updatedAt": firstString(loc, "updatedAt", "recordedAt"),
					"accuracyM": loc["accuracyM"],
				})
			}
		}
	}
	empty := []any{}
	return map[string]any{
		"cityId": nil, "generatedAt": time.Now().UTC().Format(time.RFC3339),
		"connection": "polling", "orderStream": stream,
		"couriers": couriers, "warehouses": empty, "delays": empty, "failedDeliveries": empty,
		"bottlenecks": empty, "emergencies": empty, "alerts": empty,
		"counts": map[string]any{
			"activeOrders": dash["ordersLive"], "delayedOrders": dash["delayedOrders"],
			"availableCouriers": len(couriers), "openEmergencies": dash["openIncidents"],
		},
	}, nil
}
