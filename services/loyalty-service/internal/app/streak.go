package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// UpdateStreakInput updates daily streak.
type UpdateStreakInput struct {
	TenantID  uuid.UUID
	AccountID uuid.UUID
	Action    string // increment | break | recover
	Date      string // optional YYYY-MM-DD; defaults to today
}

// UpdateStreak increments, breaks, or recovers a streak.
func (d *Deps) UpdateStreak(ctx context.Context, in UpdateStreakInput) (domain.Streak, error) {
	if _, err := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID); err != nil {
		return domain.Streak{}, err
	}
	now := d.now()
	today := in.Date
	if today == "" {
		today = now.Format("2006-01-02")
	}
	action := in.Action
	if action == "" {
		action = "increment"
	}

	s, err := d.Streaks.GetStreak(ctx, in.TenantID, in.AccountID)
	if err != nil {
		s = domain.Streak{
			ID: d.newID(), TenantID: in.TenantID, AccountID: in.AccountID,
			CurrentCount: 0, LongestCount: 0, UpdatedAt: now,
		}
	}

	switch action {
	case "break":
		s.CurrentCount = 0
		s.Broken = true
		s.LastActiveDate = today
	case "recover":
		if !s.Broken {
			return domain.Streak{}, fmt.Errorf("%w: streak not broken", domain.ErrConflict)
		}
		if s.RecoveryUsed {
			return domain.Streak{}, fmt.Errorf("%w: recovery already used", domain.ErrConflict)
		}
		s.Broken = false
		s.RecoveryUsed = true
		if s.CurrentCount < 1 {
			s.CurrentCount = 1
		}
		s.LastActiveDate = today
	case "increment":
		if s.LastActiveDate == today {
			// already counted today
		} else if s.LastActiveDate == "" || isNextDay(s.LastActiveDate, today) {
			s.CurrentCount++
			s.Broken = false
		} else {
			// gap — break then start at 1
			s.CurrentCount = 1
			s.Broken = false
		}
		if s.CurrentCount > s.LongestCount {
			s.LongestCount = s.CurrentCount
		}
		s.LastActiveDate = today
	default:
		return domain.Streak{}, fmt.Errorf("%w: unknown streak action %q", domain.ErrInvalidArgument, action)
	}

	s.UpdatedAt = now
	if err := s.Validate(); err != nil {
		return domain.Streak{}, err
	}
	if err := d.Streaks.UpsertStreak(ctx, s); err != nil {
		return domain.Streak{}, err
	}
	acct, _ := d.Accounts.GetAccount(ctx, in.TenantID, in.AccountID)
	d.emit(ctx, acct, domain.EventStreakUpdated, map[string]any{
		"current": s.CurrentCount, "longest": s.LongestCount, "action": action, "broken": s.Broken,
	})
	return s, nil
}

func isNextDay(prev, today string) bool {
	p, err1 := time.Parse("2006-01-02", prev)
	t, err2 := time.Parse("2006-01-02", today)
	if err1 != nil || err2 != nil {
		return false
	}
	return t.Equal(p.AddDate(0, 0, 1))
}
