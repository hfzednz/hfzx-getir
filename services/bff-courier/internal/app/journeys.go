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

type Deps struct {
	HTTP         *http.Client
	OrderURL     string
	TrackingURL  string
	RealtimeURL  string
	PublishToken string

	mu   sync.Mutex
	duty map[string]bool
}

func DepsFromEnv() *Deps {
	return &Deps{
		HTTP:         &http.Client{Timeout: 15 * time.Second},
		OrderURL:     strings.TrimRight(os.Getenv("ORDER_URL"), "/"),
		TrackingURL:  strings.TrimRight(os.Getenv("TRACKING_URL"), "/"),
		RealtimeURL:  strings.TrimRight(os.Getenv("REALTIME_URL"), "/"),
		PublishToken: os.Getenv("REALTIME_PUBLISH_TOKEN"),
		duty:         map[string]bool{},
	}
}

func dutyKey(tenant, courierID string) string {
	return tenant + ":" + courierID
}

func (d *Deps) Duty(_ context.Context, tenant, courierID string, on bool) (map[string]any, error) {
	if tenant == "" || courierID == "" {
		return nil, ErrInvalid
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.duty == nil {
		d.duty = map[string]bool{}
	}
	d.duty[dutyKey(tenant, courierID)] = on
	return map[string]any{"courierId": courierID, "onDuty": on, "persisted": true}, nil
}

func (d *Deps) GetDuty(_ context.Context, tenant, courierID string) (map[string]any, error) {
	if courierID == "" {
		return nil, ErrInvalid
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	on := false
	if d.duty != nil {
		on = d.duty[dutyKey(tenant, courierID)]
	}
	return map[string]any{"courierId": courierID, "onDuty": on}, nil
}

func (d *Deps) ListOffers(ctx context.Context, tenant, courierID string) ([]map[string]any, error) {
	orders, err := d.listOrders(ctx, tenant)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(orders))
	for _, o := range orders {
		st := firstString(o, "status")
		switch st {
		case "ready_for_dispatch", "courier_assigned", "out_for_delivery":
			if courierID != "" {
				ref := firstString(o, "courierRef")
				if ref != "" && ref != courierID && st != "ready_for_dispatch" {
					continue
				}
			}
			out = append(out, orderToJob(o))
		}
	}
	return out, nil
}

func (d *Deps) GetJob(ctx context.Context, tenant, jobID string) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return orderToJob(map[string]any{"id": jobID, "status": "assigned"}), nil
	}
	o, err := d.getOrder(ctx, tenant, jobID)
	if err != nil {
		return nil, err
	}
	return orderToJob(o), nil
}

func (d *Deps) Offer(ctx context.Context, tenant, courierID, jobID string, accept bool) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	st := "rejected"
	if accept {
		st = "accepted"
	}
	out := map[string]any{"jobId": jobID, "status": st, "courierId": courierID}
	if !accept || d.OrderURL == "" {
		return out, nil
	}
	order, err := d.dispatch(ctx, tenant, jobID, "CourierAssigned", courierID, "courier_assigned")
	if err != nil {
		return nil, err
	}
	out["order"] = order
	out["status"] = firstString(order, "status")
	if out["status"] == "" {
		out["status"] = "assigned"
	}
	return out, nil
}

func (d *Deps) Enroute(ctx context.Context, tenant, jobID string) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return orderToJob(map[string]any{"id": jobID, "status": "out_for_delivery"}), nil
	}
	return d.dispatch(ctx, tenant, jobID, "OutForDelivery", "", "out_for_delivery")
}

func (d *Deps) Complete(ctx context.Context, tenant, jobID string) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return orderToJob(map[string]any{"id": jobID, "status": "delivered"}), nil
	}
	o, err := d.getOrder(ctx, tenant, jobID)
	if err != nil {
		return nil, err
	}
	st, _ := o["status"].(string)
	if st == "courier_assigned" || st == "ready_for_dispatch" {
		if _, err := d.dispatch(ctx, tenant, jobID, "OutForDelivery", "", "out_for_delivery"); err != nil {
			return nil, err
		}
	}
	return d.dispatch(ctx, tenant, jobID, "Delivered", "", "delivered")
}

func (d *Deps) Fail(ctx context.Context, tenant, jobID, reason string) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	if reason == "" {
		reason = "delivery_failed"
	}
	if d.OrderURL == "" {
		return map[string]any{"jobId": jobID, "status": "failed", "reason": reason}, nil
	}
	var out map[string]any
	if err := d.postJSON(ctx, d.OrderURL+"/v1/orders/"+jobID+"/cancel", tenant, map[string]any{
		"reason": reason,
	}, &out); err != nil {
		return nil, err
	}
	st := firstString(out, "status")
	if st == "" {
		st = "cancelled"
	}
	d.fanout(ctx, tenant, jobID, st, "Cancelled")
	job := orderToJob(out)
	job["status"] = "failed"
	job["reason"] = reason
	return job, nil
}

func (d *Deps) Transition(ctx context.Context, tenant, courierID, jobID, status string) (map[string]any, error) {
	switch strings.ToLower(strings.ReplaceAll(status, "-", "_")) {
	case "assigned", "accepted":
		return d.Offer(ctx, tenant, courierID, jobID, true)
	case "en_route_store", "en_route_customer", "picked_up", "at_store", "arrived", "out_for_delivery":
		return d.Enroute(ctx, tenant, jobID)
	case "delivered":
		return d.Complete(ctx, tenant, jobID)
	case "failed", "cancelled", "canceled":
		return d.Fail(ctx, tenant, jobID, status)
	default:
		return nil, ErrInvalid
	}
}

func (d *Deps) dispatch(ctx context.Context, tenant, orderID, eventType, courierRef, statusHint string) (map[string]any, error) {
	body := map[string]any{"eventType": eventType}
	if courierRef != "" {
		body["courierRef"] = courierRef
	}
	var out map[string]any
	if err := d.postJSON(ctx, d.OrderURL+"/v1/orders/"+orderID+"/events/dispatch", tenant, body, &out); err != nil {
		return nil, err
	}
	st, _ := out["status"].(string)
	if st == "" {
		st = statusHint
	}
	d.fanout(ctx, tenant, orderID, st, eventType)
	return orderToJob(out), nil
}

func (d *Deps) listOrders(ctx context.Context, tenant string) ([]map[string]any, error) {
	if d.OrderURL == "" {
		return []map[string]any{}, nil
	}
	q := url.Values{}
	q.Set("limit", "100")
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
			log.Warn("courier.fanout.tracking", "err", err, "orderId", orderID)
		}
	}
	if d.RealtimeURL != "" {
		if err := d.postJSON(ctx, d.RealtimeURL+"/v1/realtime/publish", tenant, map[string]any{
			"topic":   "order:" + orderID,
			"payload": map[string]any{"orderId": orderID, "status": status, "eventType": eventType},
		}, nil); err != nil {
			log.Warn("courier.fanout.realtime", "err", err, "orderId", orderID)
		}
	}
}

func (d *Deps) postJSON(ctx context.Context, url, tenant string, body, out any) error {
	raw, _ := json.Marshal(body)
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
		if resp.StatusCode == 400 {
			return fmt.Errorf("%w: %s", ErrInvalid, string(b))
		}
		return fmt.Errorf("upstream %s: %d %s", url, resp.StatusCode, string(b))
	}
	if out == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, out)
}

func (d *Deps) UpdateLocation(ctx context.Context, tenant, courierID string, lat, lon, accuracyM float64) (map[string]any, error) {
	if tenant == "" || courierID == "" {
		return nil, ErrInvalid
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return nil, fmt.Errorf("%w: invalid lat/lon", ErrInvalid)
	}
	if d.TrackingURL == "" {
		return nil, fmt.Errorf("tracking-service not configured")
	}
	var out map[string]any
	if err := d.postJSON(ctx, d.TrackingURL+"/v1/tracking/locations", tenant, map[string]any{
		"courierId": courierID, "lat": lat, "lon": lon, "accuracyM": accuracyM,
		"recordedAt": time.Now().UTC().Format(time.RFC3339),
	}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (d *Deps) LiveLocation(ctx context.Context, tenant, courierID string) (map[string]any, error) {
	if tenant == "" || courierID == "" {
		return nil, ErrInvalid
	}
	if d.TrackingURL == "" {
		return nil, fmt.Errorf("tracking-service not configured")
	}
	var out map[string]any
	if err := d.getJSON(ctx, d.TrackingURL+"/v1/tracking/couriers/"+url.PathEscape(courierID)+"/live", tenant, &out); err != nil {
		return nil, err
	}
	return out, nil
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

func orderToJob(o map[string]any) map[string]any {
	id := firstString(o, "id", "jobId", "orderId")
	st := firstString(o, "status", "orderStatus")
	jobSt := "assigned"
	switch st {
	case "out_for_delivery":
		jobSt = "en_route_customer"
	case "delivered", "completed":
		jobSt = "delivered"
	case "cancelled", "failed":
		jobSt = "failed"
	case "courier_assigned", "ready_for_dispatch", "accepted":
		jobSt = "assigned"
	}
	return map[string]any{
		"id":            id,
		"jobId":         id,
		"order_id":      id,
		"orderId":       id,
		"status":        jobSt,
		"orderStatus":   st,
		"store_name":    firstString(o, "storeName"),
		"customer_area": firstString(o, "customerArea"),
		"courierRef":    firstString(o, "courierRef"),
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
