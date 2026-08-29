package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid argument")

type Deps struct {
	HTTP         *http.Client
	OrderURL     string
	TrackingURL  string
	RealtimeURL  string
	PublishToken string
}

func DepsFromEnv() *Deps {
	return &Deps{
		HTTP:         &http.Client{Timeout: 15 * time.Second},
		OrderURL:     strings.TrimRight(os.Getenv("ORDER_URL"), "/"),
		TrackingURL:  strings.TrimRight(os.Getenv("TRACKING_URL"), "/"),
		RealtimeURL:  strings.TrimRight(os.Getenv("REALTIME_URL"), "/"),
		PublishToken: os.Getenv("REALTIME_PUBLISH_TOKEN"),
	}
}

func (d *Deps) Duty(_ context.Context, tenant, courierID string, on bool) (map[string]any, error) {
	if tenant == "" || courierID == "" {
		return nil, ErrInvalid
	}
	return map[string]any{"courierId": courierID, "onDuty": on}, nil
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
	return out, nil
}

func (d *Deps) Enroute(ctx context.Context, tenant, jobID string) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return map[string]any{"jobId": jobID, "status": "out_for_delivery"}, nil
	}
	return d.dispatch(ctx, tenant, jobID, "OutForDelivery", "", "out_for_delivery")
}

func (d *Deps) Complete(ctx context.Context, tenant, jobID string) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	if d.OrderURL == "" {
		return map[string]any{"jobId": jobID, "status": "delivered"}, nil
	}
	o, err := d.getOrder(ctx, tenant, jobID)
	if err != nil {
		return nil, err
	}
	st, _ := o["status"].(string)
	if st == "courier_assigned" {
		if _, err := d.dispatch(ctx, tenant, jobID, "OutForDelivery", "", "out_for_delivery"); err != nil {
			return nil, err
		}
	}
	return d.dispatch(ctx, tenant, jobID, "Delivered", "", "delivered")
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
	return out, nil
}

func (d *Deps) getOrder(ctx context.Context, tenant, orderID string) (map[string]any, error) {
	var out map[string]any
	err := d.getJSON(ctx, d.OrderURL+"/v1/orders/"+orderID, tenant, &out)
	return out, err
}

func (d *Deps) fanout(ctx context.Context, tenant, orderID, status, eventType string) {
	if d.TrackingURL != "" {
		_ = d.postJSON(ctx, d.TrackingURL+"/v1/tracking/orders/"+orderID+"/timeline", tenant, map[string]any{
			"type": "Custom", "message": status,
			"meta": map[string]any{"status": status, "eventType": eventType},
		}, nil)
	}
	if d.RealtimeURL != "" {
		_ = d.postJSON(ctx, d.RealtimeURL+"/v1/realtime/publish", tenant, map[string]any{
			"topic": "order:" + orderID,
			"payload": map[string]any{"orderId": orderID, "status": status, "eventType": eventType},
		}, nil)
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
