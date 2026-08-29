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
	HTTP       *http.Client
	OrderURL   string
	TrackingURL string
	RealtimeURL string
	PublishToken string
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
		return o, nil
	}
	return d.lifecycle(ctx, tenant, taskID, "warehouse", "PackingCompleted", "ready_for_dispatch")
}

func (d *Deps) lifecycle(ctx context.Context, tenant, orderID, kind, eventType, statusHint string) (map[string]any, error) {
	if d.OrderURL == "" {
		return map[string]any{"taskId": orderID, "status": statusHint}, nil
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
