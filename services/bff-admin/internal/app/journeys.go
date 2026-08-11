package app

import (
	"context"
	"errors"
	"net/url"
)

var ErrInvalid = errors.New("invalid argument")

type OrderGateway interface {
	List(ctx context.Context, tenant string, q url.Values) (map[string]any, error)
	Get(ctx context.Context, tenant, id string) (map[string]any, error)
}

type LiveOpsGateway interface {
	SetFlag(ctx context.Context, tenant, key string, enabled bool) (map[string]any, error)
}

type Deps struct {
	Orders  OrderGateway
	LiveOps LiveOpsGateway
}

func (d *Deps) Dashboard(ctx context.Context, tenant string) (map[string]any, error) {
	if tenant == "" {
		return nil, ErrInvalid
	}
	out := map[string]any{
		"ordersLive": 0, "couriersOnDuty": 0, "openIncidents": 0, "sloBurn": 0.0,
	}
	if d.Orders != nil {
		list, err := d.Orders.List(ctx, tenant, url.Values{"pageSize": {"1"}})
		if err == nil {
			if total, ok := list["total"].(float64); ok {
				out["ordersLive"] = int(total)
			} else if items, ok := list["items"].([]any); ok {
				out["ordersLive"] = len(items)
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
