package memory

import (
	"context"

	"github.com/nexora/wallet-service/internal/app/ports"
)

// EventPublisher records published events on the store.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.published = append(p.S.published, PublishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// LedgerClient is a journal post stub.
type LedgerClient struct {
	Calls int
}

func (l *LedgerClient) PostJournal(_ context.Context, _ ports.PostJournalRequest) (ports.PostJournalResult, error) {
	l.Calls++
	return ports.PostJournalResult{JournalID: "j1", Posted: true}, nil
}

var _ ports.EventPublisher = (*EventPublisher)(nil)
var _ ports.LedgerClient = (*LedgerClient)(nil)
