package memory

import (
	"context"

	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// EventPublisher records published events on the store.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.published = append(p.S.published, PublishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// FraudClient is a controllable fraud stub.
type FraudClient struct {
	RiskScore int
	Decision  domain.FraudDecision
	Reasons   []string
	Err       error
}

func (f *FraudClient) Score(_ context.Context, _ ports.FraudRequest) (ports.FraudResult, error) {
	if f.Err != nil {
		return ports.FraudResult{}, f.Err
	}
	dec := f.Decision
	if dec == "" {
		dec = domain.FraudAllow
	}
	return ports.FraudResult{Score: f.RiskScore, Decision: dec, Reasons: f.Reasons}, nil
}

// WalletClient is a controllable wallet debit stub.
type WalletClient struct {
	Success bool
	Reason  string
	Err     error
	EntryID string
}

func (w *WalletClient) Debit(_ context.Context, _ ports.WalletDebitRequest) (ports.WalletDebitResult, error) {
	if w.Err != nil {
		return ports.WalletDebitResult{}, w.Err
	}
	entry := w.EntryID
	if entry == "" {
		entry = "wallet-entry-1"
	}
	return ports.WalletDebitResult{WalletID: "w1", EntryID: entry, Success: w.Success, Reason: w.Reason}, nil
}

// LedgerClient is a journal post stub.
type LedgerClient struct {
	Calls int
	Err   error
}

func (l *LedgerClient) PostJournal(_ context.Context, _ ports.PostJournalRequest) (ports.PostJournalResult, error) {
	l.Calls++
	if l.Err != nil {
		return ports.PostJournalResult{}, l.Err
	}
	return ports.PostJournalResult{JournalID: "j1", Posted: true}, nil
}

var _ ports.EventPublisher = (*EventPublisher)(nil)
var _ ports.FraudClient = (*FraudClient)(nil)
var _ ports.WalletClient = (*WalletClient)(nil)
var _ ports.LedgerClient = (*LedgerClient)(nil)
