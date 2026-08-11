package memory

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
)

// EventPublisher records published events on the store.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.published = append(p.S.published, PublishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// WalletClient is an in-memory WalletClient.Credit stub.
type WalletClient struct {
	mu    sync.Mutex
	Calls []ports.WalletCreditRequest
	Fail  error
}

func (w *WalletClient) Credit(_ context.Context, req ports.WalletCreditRequest) (ports.WalletCreditResult, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Calls = append(w.Calls, req)
	if w.Fail != nil {
		return ports.WalletCreditResult{}, w.Fail
	}
	return ports.WalletCreditResult{
		WalletID: uuid.NewString(),
		EntryID:  uuid.NewString(),
		Credited: true,
	}, nil
}

var _ ports.EventPublisher = (*EventPublisher)(nil)
var _ ports.WalletClient = (*WalletClient)(nil)
