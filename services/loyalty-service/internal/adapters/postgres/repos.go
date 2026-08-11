package postgres

import "database/sql"

// Repos aggregates all postgres port implementations for loyalty-service.
type Repos struct {
	Accounts     *AccountRepo
	Memberships  *MembershipRepo
	Rewards      *RewardRepo
	Referrals    *ReferralRepo
	Missions     *MissionRepo
	Achievements *AchievementRepo
	Streaks      *StreakRepo
	Spins        *SpinRepo
	Collectibles *CollectibleRepo
	Cashbacks    *CashbackRepo
	AIScores     *AIScoreRepo
	Outbox       *OutboxRepo
}

// NewRepos wires all repository adapters on a shared *sql.DB.
func NewRepos(db *sql.DB) *Repos {
	return &Repos{
		Accounts:     &AccountRepo{DB: db},
		Memberships:  &MembershipRepo{DB: db},
		Rewards:      &RewardRepo{DB: db},
		Referrals:    &ReferralRepo{DB: db},
		Missions:     &MissionRepo{DB: db},
		Achievements: &AchievementRepo{DB: db},
		Streaks:      &StreakRepo{DB: db},
		Spins:        &SpinRepo{DB: db},
		Collectibles: &CollectibleRepo{DB: db},
		Cashbacks:    &CashbackRepo{DB: db},
		AIScores:     &AIScoreRepo{DB: db},
		Outbox:       &OutboxRepo{DB: db},
	}
}
