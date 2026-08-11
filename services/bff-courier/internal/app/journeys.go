package app

import (
	"context"
	"errors"
)

var ErrInvalid = errors.New("invalid argument")

type Deps struct{}

func (Deps) Duty(ctx context.Context, tenant, courierID string, on bool) (map[string]any, error) {
	if tenant == "" || courierID == "" {
		return nil, ErrInvalid
	}
	return map[string]any{"courierId": courierID, "onDuty": on}, nil
}

func (Deps) Offer(ctx context.Context, tenant, courierID, jobID string, accept bool) (map[string]any, error) {
	if jobID == "" {
		return nil, ErrInvalid
	}
	st := "rejected"
	if accept {
		st = "accepted"
	}
	return map[string]any{"jobId": jobID, "status": st, "courierId": courierID}, nil
}

func (Deps) Complete(ctx context.Context, tenant, jobID string) (map[string]any, error) {
	return map[string]any{"jobId": jobID, "status": "delivered"}, nil
}
