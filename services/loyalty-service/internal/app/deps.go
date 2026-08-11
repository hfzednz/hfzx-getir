package app

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app/ports"
	"github.com/nexora/loyalty-service/internal/domain"
)

// Deps aggregates application ports for loyalty use cases.
type Deps struct {
	Accounts     ports.AccountRepo
	Memberships  ports.MembershipRepo
	Rewards      ports.RewardRepo
	Referrals    ports.ReferralRepo
	Missions     ports.MissionRepo
	Achievements ports.AchievementRepo
	Streaks      ports.StreakRepo
	Spins        ports.SpinRepo
	Collectibles ports.CollectibleRepo
	Cashbacks    ports.CashbackRepo
	AIScores     ports.AIScoreRepo
	Outbox       ports.OutboxRepository
	Wallet       ports.WalletClient
	Publisher    ports.EventPublisher
	Clock        ports.Clock
	IDs          ports.IDGen
	Rand         ports.Rand
}

// SystemClock is a real-time Clock.
type SystemClock struct{}

func (SystemClock) Now() time.Time { return time.Now().UTC() }

// UUIDGen generates random UUIDs.
type UUIDGen struct{}

func (UUIDGen) New() uuid.UUID { return uuid.New() }

// SystemRand is a math/rand-backed Rand.
type SystemRand struct{ r *rand.Rand }

// NewSystemRand returns a seeded system rand.
func NewSystemRand(seed int64) *SystemRand {
	return &SystemRand{r: rand.New(rand.NewSource(seed))}
}

func (s *SystemRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	if s.r == nil {
		return rand.Intn(n)
	}
	return s.r.Intn(n)
}

func (d *Deps) now() time.Time {
	if d.Clock != nil {
		return d.Clock.Now().UTC()
	}
	return time.Now().UTC()
}

func (d *Deps) newID() uuid.UUID {
	if d.IDs != nil {
		return d.IDs.New()
	}
	return uuid.New()
}

func (d *Deps) intn(n int) int {
	if d.Rand != nil {
		return d.Rand.Intn(n)
	}
	return rand.Intn(n)
}

func (d *Deps) enqueueOutbox(ctx context.Context, tenantID, accountID uuid.UUID, eventType string, payload map[string]any) {
	if d.Outbox == nil {
		return
	}
	now := d.now()
	_ = d.Outbox.Enqueue(ctx, domain.OutboxMessage{
		ID: d.newID(), TenantID: tenantID, AccountID: accountID,
		Topic: domain.TopicForEvent(eventType), Key: accountID.String(),
		Payload: payload, Status: domain.OutboxStatusPending,
		CreatedAt: now, UpdatedAt: now,
	})
}

func (d *Deps) emit(ctx context.Context, acct domain.Account, eventType string, extra map[string]any) {
	now := d.now()
	payload := map[string]any{
		"type": eventType, "accountId": acct.ID.String(), "tenantId": acct.TenantID.String(),
		"principalId": acct.PrincipalID.String(), "occurredAt": now,
	}
	for k, v := range extra {
		payload[k] = v
	}
	d.enqueueOutbox(ctx, acct.TenantID, acct.ID, eventType, payload)
	if d.Publisher != nil {
		_ = d.Publisher.Publish(ctx, domain.TopicForEvent(eventType), acct.ID.String(), payload)
	}
}

// PublishPending drains the outbox via EventPublisher.
func (d *Deps) PublishPending(ctx context.Context, limit int) (int, error) {
	if d.Outbox == nil || d.Publisher == nil {
		return 0, nil
	}
	if limit <= 0 {
		limit = 100
	}
	msgs, err := d.Outbox.ListPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	n := 0
	now := d.now()
	for _, m := range msgs {
		if err := d.Publisher.Publish(ctx, m.Topic, m.Key, m.Payload); err != nil {
			m.Attempts++
			m.LastError = err.Error()
			m.Status = domain.OutboxStatusFailed
			m.UpdatedAt = now
			_ = d.Outbox.Update(ctx, m)
			continue
		}
		m.Attempts++
		m.Status = domain.OutboxStatusPublished
		m.PublishedAt = &now
		m.UpdatedAt = now
		_ = d.Outbox.Update(ctx, m)
		n++
	}
	return n, nil
}
