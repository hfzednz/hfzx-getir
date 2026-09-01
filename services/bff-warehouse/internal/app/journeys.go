package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalid      = errors.New("invalid argument")
	ErrNotSupported = errors.New("not supported")
)

type linePick struct {
	Picked int
	Short  int
}

type Deps struct {
	HTTP         *http.Client
	OrderURL     string
	TrackingURL  string
	RealtimeURL  string
	PublishToken string
	mu           sync.Mutex
	picks        map[string]map[string]linePick
}

func DepsFromEnv() *Deps {
	d := &Deps{
		HTTP:         &http.Client{Timeout: 15 * time.Second},
		OrderURL:     strings.TrimRight(os.Getenv("ORDER_URL"), "/"),
		TrackingURL:  strings.TrimRight(os.Getenv("TRACKING_URL"), "/"),
		RealtimeURL:  strings.TrimRight(os.Getenv("REALTIME_URL"), "/"),
		PublishToken: os.Getenv("REALTIME_PUBLISH_TOKEN"),
	}
	return d
}

func (d *Deps) ListTasks(ctx context.Context, tenant string) ([]map[string]any, error) {
	orders, err := d.listOrders(ctx, tenant, "")
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(orders))
	for _, o := range orders {
		st := firstString(o, "status")
		switch st {
		case "warehouse_assigned", "picking", "packing", "confirmed", "inventory_reservation":
			out = append(out, orderToPickTask(o))
		}
	}
	return out, nil
}

func (d *Deps) GetTask(ctx context.Context, tenant, taskID string) (map[string]any, error) {
	if taskID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return d.withPickProgress(tenant, taskID, orderToPickTask(map[string]any{"id": taskID, "status": "queued"})), nil
	}
	o, err := d.getOrder(ctx, tenant, taskID)
	if err != nil {
		return nil, err
	}
	return d.withPickProgress(tenant, taskID, orderToPickTask(o)), nil
}

func (d *Deps) Pick(ctx context.Context, tenant, taskID string) (map[string]any, error) {
	if taskID == "" {
		return nil, ErrInvalid
	}
	return d.lifecycle(ctx, tenant, taskID, "warehouse", "PickingStarted", "picking")
}

func (d *Deps) Pack(ctx context.Context, tenant, taskID string) (map[string]any, error) {
	if taskID == "" {
		return nil, ErrInvalid
	}
	return d.lifecycle(ctx, tenant, taskID, "warehouse", "PackingCompleted", "ready_for_dispatch")
}

func (d *Deps) DispatchReady(ctx context.Context, tenant, taskID string) (map[string]any, error) {
	if taskID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return map[string]any{"taskId": taskID, "status": "ready_for_dispatch"}, nil
	}
	o, err := d.getOrder(ctx, tenant, taskID)
	if err != nil {
		return nil, err
	}
	st, _ := o["status"].(string)
	if st == "ready_for_dispatch" || st == "courier_assigned" || st == "out_for_delivery" {
		return orderToPickTask(o), nil
	}
	return d.lifecycle(ctx, tenant, taskID, "warehouse", "PackingCompleted", "ready_for_dispatch")
}

func (d *Deps) ScanLine(ctx context.Context, tenant, taskID, lineID, barcode string, qty int) (map[string]any, error) {
	if taskID == "" || lineID == "" {
		return nil, ErrInvalid
	}
	if qty <= 0 {
		qty = 1
	}
	task, err := d.GetTask(ctx, tenant, taskID)
	if err != nil {
		return nil, err
	}
	line, ok := findLine(task, lineID, barcode)
	if !ok {
		return nil, fmt.Errorf("%w: line or barcode does not match this task", ErrInvalid)
	}
	d.addPick(tenant, taskID, firstString(line, "id", "sku"), qty, 0)
	return d.GetTask(ctx, tenant, taskID)
}

func (d *Deps) ShortPick(ctx context.Context, tenant, taskID, lineID string, missingQty int) (map[string]any, error) {
	if taskID == "" || lineID == "" {
		return nil, ErrInvalid
	}
	if missingQty <= 0 {
		missingQty = 1
	}
	task, err := d.GetTask(ctx, tenant, taskID)
	if err != nil {
		return nil, err
	}
	line, ok := findLine(task, lineID, "")
	if !ok {
		return nil, fmt.Errorf("%w: line not on this task", ErrInvalid)
	}
	d.addPick(tenant, taskID, firstString(line, "id", "sku"), 0, missingQty)
	out, err := d.GetTask(ctx, tenant, taskID)
	if err != nil {
		return nil, err
	}
	out["status"] = "short_pick"
	return out, nil
}

func (d *Deps) lifecycle(ctx context.Context, tenant, orderID, kind, eventType, statusHint string) (map[string]any, error) {
	if d.OrderURL == "" {
		return orderToPickTask(map[string]any{"id": orderID, "status": statusHint}), nil
	}
	path := "/v1/orders/" + orderID + "/events/" + kind
	var out map[string]any
	if err := d.postJSON(ctx, d.OrderURL+path, tenant, map[string]any{"eventType": eventType}, &out); err != nil {
		return nil, err
	}
	st, _ := out["status"].(string)
	if st == "" {
		st = statusHint
	}
	d.fanout(ctx, tenant, orderID, st, eventType)
	return orderToPickTask(out), nil
}

func (d *Deps) listOrders(ctx context.Context, tenant, status string) ([]map[string]any, error) {
	if d.OrderURL == "" {
		return []map[string]any{}, nil
	}
	q := url.Values{}
	q.Set("limit", "100")
	if status != "" {
		q.Set("status", status)
	}
	var raw map[string]any
	if err := d.getJSON(ctx, d.OrderURL+"/v1/orders?"+q.Encode(), tenant, &raw); err != nil {
		return nil, err
	}
	arr, _ := raw["items"].([]any)
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func (d *Deps) getOrder(ctx context.Context, tenant, orderID string) (map[string]any, error) {
	var out map[string]any
	err := d.getJSON(ctx, d.OrderURL+"/v1/orders/"+orderID, tenant, &out)
	return out, err
}

func (d *Deps) fanout(ctx context.Context, tenant, orderID, status, eventType string) {
	log := slog.Default()
	if d.TrackingURL != "" {
		if err := d.postJSON(ctx, d.TrackingURL+"/v1/tracking/orders/"+orderID+"/timeline", tenant, map[string]any{
			"type": "Custom", "message": status,
			"meta": map[string]any{"status": status, "eventType": eventType},
		}, nil); err != nil {
			log.Warn("warehouse.fanout.tracking", "err", err, "orderId", orderID)
		}
	}
	if d.RealtimeURL != "" {
		if err := d.postJSON(ctx, d.RealtimeURL+"/v1/realtime/publish", tenant, map[string]any{
			"topic":   "order:" + orderID,
			"payload": map[string]any{"orderId": orderID, "status": status, "eventType": eventType},
		}, nil); err != nil {
			log.Warn("warehouse.fanout.realtime", "err", err, "orderId", orderID)
		}
	}
}

func (d *Deps) postJSON(ctx context.Context, url, tenant string, body, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if tenant != "" {
		req.Header.Set("X-Tenant-Id", tenant)
	}
	if d.PublishToken != "" && strings.Contains(url, "/v1/realtime/publish") {
		req.Header.Set("X-Realtime-Publish-Token", d.PublishToken)
	}
	resp, err := d.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upstream %s: %d %s", url, resp.StatusCode, string(b))
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

func (d *Deps) getJSON(ctx context.Context, url, tenant string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if tenant != "" {
		req.Header.Set("X-Tenant-Id", tenant)
	}
	resp, err := d.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upstream %s: %d %s", url, resp.StatusCode, string(b))
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

func (d *Deps) addPick(tenant, taskID, lineID string, picked, short int) {
	if lineID == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.picks == nil {
		d.picks = map[string]map[string]linePick{}
	}
	key := tenant + "|" + taskID
	if d.picks[key] == nil {
		d.picks[key] = map[string]linePick{}
	}
	cur := d.picks[key][lineID]
	cur.Picked += picked
	cur.Short += short
	d.picks[key][lineID] = cur
}

func (d *Deps) withPickProgress(tenant, taskID string, task map[string]any) map[string]any {
	d.mu.Lock()
	prog := d.picks[tenant+"|"+taskID]
	d.mu.Unlock()
	if prog == nil {
		return task
	}
	raw, _ := task["lines"].([]map[string]any)
	if raw == nil {
		raw = []map[string]any{}
	}
	seen := map[string]bool{}
	for i, line := range raw {
		id := firstString(line, "id", "sku")
		seen[id] = true
		if p, ok := prog[id]; ok {
			line["picked_qty"] = p.Picked
			line["pickedQty"] = p.Picked
			line["shorted"] = p.Short > 0
			line["short_qty"] = p.Short
			raw[i] = line
		}
	}
	for id, p := range prog {
		if seen[id] {
			continue
		}
		raw = append(raw, map[string]any{
			"id": id, "sku": id, "barcode": id, "qty": p.Picked + p.Short,
			"picked_qty": p.Picked, "pickedQty": p.Picked,
			"shorted": p.Short > 0, "short_qty": p.Short,
		})
	}
	task["lines"] = raw
	return task
}

func findLine(task map[string]any, lineID, barcode string) (map[string]any, bool) {
	raw, _ := task["lines"].([]map[string]any)
	wantLine := strings.EqualFold(strings.TrimSpace(lineID), "")
	wantCode := strings.ToUpper(strings.TrimSpace(barcode))
	for _, line := range raw {
		id := firstString(line, "id", "sku")
		sku := strings.ToUpper(firstString(line, "sku", "barcode"))
		if !wantLine && (strings.EqualFold(id, lineID) || strings.EqualFold(sku, lineID)) {
			if wantCode == "" || sku == wantCode || strings.EqualFold(firstString(line, "barcode"), barcode) {
				return line, true
			}
		}
		if wantCode != "" && (sku == wantCode || strings.EqualFold(firstString(line, "barcode"), barcode)) {
			return line, true
		}
	}
	if lineID != "" && len(raw) == 0 {
		code := barcode
		if code == "" {
			code = lineID
		}
		return map[string]any{"id": lineID, "sku": lineID, "barcode": code}, true
	}
	return nil, false
}

func orderToPickTask(o map[string]any) map[string]any {
	id := firstString(o, "id", "taskId", "orderId")
	st := firstString(o, "status", "orderStatus")
	pick := "queued"
	switch st {
	case "picking":
		pick = "in_progress"
	case "packing":
		pick = "picked"
	case "ready_for_dispatch", "courier_assigned", "out_for_delivery":
		pick = "staged"
	}
	return map[string]any{
		"id":          id,
		"taskId":      id,
		"order_id":    id,
		"orderId":     id,
		"status":      pick,
		"orderStatus": st,
		"lines":       orderLines(o),
	}
}

func orderLines(o map[string]any) []map[string]any {
	raw, _ := o["lines"].([]any)
	if raw == nil {
		raw, _ = o["items"].([]any)
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		sku := firstString(m, "sku", "skuCode", "variantId", "productId", "product_id")
		qty := asInt(m["qty"])
		if qty == 0 {
			qty = asInt(m["quantity"])
		}
		if qty == 0 {
			qty = 1
		}
		name := firstString(m, "titleSnapshot", "name", "title")
		out = append(out, map[string]any{
			"id":           firstString(m, "id"),
			"sku":          sku,
			"barcode":      sku,
			"bin":          firstString(m, "bin"),
			"qty":          qty,
			"name":         name,
			"product_name": name,
		})
	}
	return out
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
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
