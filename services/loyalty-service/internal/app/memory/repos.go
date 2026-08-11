package memory

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// Repos bundles all in-memory repositories sharing one store.
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

// NewRepos returns wired memory repositories.
func NewRepos(s *Store) *Repos {
	return &Repos{
		Accounts:     &AccountRepo{S: s},
		Memberships:  &MembershipRepo{S: s},
		Rewards:      &RewardRepo{S: s},
		Referrals:    &ReferralRepo{S: s},
		Missions:     &MissionRepo{S: s},
		Achievements: &AchievementRepo{S: s},
		Streaks:      &StreakRepo{S: s},
		Spins:        &SpinRepo{S: s},
		Collectibles: &CollectibleRepo{S: s},
		Cashbacks:    &CashbackRepo{S: s},
		AIScores:     &AIScoreRepo{S: s},
		Outbox:       &OutboxRepo{S: s},
	}
}

// --- AccountRepo ---

type AccountRepo struct{ S *Store }

func (r *AccountRepo) CreateAccount(_ context.Context, a domain.Account) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := prinKey(a.TenantID, a.PrincipalID)
	if _, ok := r.S.byPrin[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.accounts[a.ID] = a
	r.S.byPrin[k] = a.ID
	return nil
}

func (r *AccountRepo) GetAccount(_ context.Context, tenantID, accountID uuid.UUID) (domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.accounts[accountID]
	if !ok || a.TenantID != tenantID {
		return domain.Account{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *AccountRepo) GetAccountByPrincipal(_ context.Context, tenantID, principalID uuid.UUID) (domain.Account, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.byPrin[prinKey(tenantID, principalID)]
	if !ok {
		return domain.Account{}, domain.ErrNotFound
	}
	return r.S.accounts[id], nil
}

func (r *AccountRepo) UpdateAccount(_ context.Context, a domain.Account) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.accounts[a.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.accounts[a.ID] = a
	return nil
}

func (r *AccountRepo) CreateLedgerEntry(_ context.Context, e domain.PointLedgerEntry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if e.IdempotencyKey != "" {
		k := idemKey(e.TenantID, e.IdempotencyKey)
		if _, ok := r.S.ledgerIdem[k]; ok {
			return domain.ErrAlreadyExists
		}
		r.S.ledgerIdem[k] = e.ID
	}
	if e.OrderID != nil {
		r.S.ledgerOrd[orderKey(e.TenantID, e.AccountID, *e.OrderID)] = e.ID
	}
	r.S.ledger = append(r.S.ledger, e)
	return nil
}

func (r *AccountRepo) GetLedgerByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.PointLedgerEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ledgerIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.PointLedgerEntry{}, domain.ErrNotFound
	}
	for _, e := range r.S.ledger {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.PointLedgerEntry{}, domain.ErrNotFound
}

func (r *AccountRepo) GetLedgerByOrder(_ context.Context, tenantID, accountID, orderID uuid.UUID) (domain.PointLedgerEntry, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ledgerOrd[orderKey(tenantID, accountID, orderID)]
	if !ok {
		return domain.PointLedgerEntry{}, domain.ErrNotFound
	}
	for _, e := range r.S.ledger {
		if e.ID == id {
			return e, nil
		}
	}
	return domain.PointLedgerEntry{}, domain.ErrNotFound
}

func (r *AccountRepo) ListLedger(_ context.Context, tenantID, accountID uuid.UUID, limit, offset int) ([]domain.PointLedgerEntry, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.PointLedgerEntry
	for i := len(r.S.ledger) - 1; i >= 0; i-- {
		e := r.S.ledger[i]
		if e.TenantID == tenantID && e.AccountID == accountID {
			all = append(all, e)
		}
	}
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *AccountRepo) CreateAudit(_ context.Context, a domain.AuditEntry) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.audits = append(r.S.audits, a)
	return nil
}

func (r *AccountRepo) IncrStat(_ context.Context, tenantID, accountID uuid.UUID, key string, delta int64) (int64, error) {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := statKey(tenantID, accountID, key)
	r.S.stats[k] += delta
	return r.S.stats[k], nil
}

func (r *AccountRepo) GetStat(_ context.Context, tenantID, accountID uuid.UUID, key string) (int64, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	return r.S.stats[statKey(tenantID, accountID, key)], nil
}

// --- MembershipRepo ---

type MembershipRepo struct{ S *Store }

func (r *MembershipRepo) ListTiers(_ context.Context, _ uuid.UUID) ([]domain.TierConfig, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.TierConfig, len(r.S.tiers))
	copy(out, r.S.tiers)
	return out, nil
}

func (r *MembershipRepo) GetMembership(_ context.Context, tenantID, accountID uuid.UUID) (domain.Membership, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.memberships[acctKey(tenantID, accountID)]
	if !ok {
		return domain.Membership{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *MembershipRepo) UpsertMembership(_ context.Context, m domain.Membership) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.memberships[acctKey(m.TenantID, m.AccountID)] = m
	return nil
}

// --- RewardRepo ---

type RewardRepo struct{ S *Store }

func (r *RewardRepo) CreateReward(_ context.Context, rw domain.Reward) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.rewards[rw.ID] = rw
	r.S.rewardCodes[codeKey(rw.TenantID, rw.Code)] = rw.ID
	return nil
}

func (r *RewardRepo) GetReward(_ context.Context, tenantID, rewardID uuid.UUID) (domain.Reward, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	rw, ok := r.S.rewards[rewardID]
	if !ok || rw.TenantID != tenantID {
		return domain.Reward{}, domain.ErrNotFound
	}
	return rw, nil
}

func (r *RewardRepo) GetRewardByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Reward, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.rewardCodes[codeKey(tenantID, code)]
	if !ok {
		return domain.Reward{}, domain.ErrNotFound
	}
	return r.S.rewards[id], nil
}

func (r *RewardRepo) ListRewards(_ context.Context, tenantID uuid.UUID) ([]domain.Reward, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Reward
	for _, rw := range r.S.rewards {
		if rw.TenantID == tenantID {
			out = append(out, rw)
		}
	}
	return out, nil
}

func (r *RewardRepo) CreateRedemption(_ context.Context, red domain.Redemption) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.redemptions[red.ID] = red
	return nil
}

func (r *RewardRepo) GetRedemption(_ context.Context, tenantID, redemptionID uuid.UUID) (domain.Redemption, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	red, ok := r.S.redemptions[redemptionID]
	if !ok || red.TenantID != tenantID {
		return domain.Redemption{}, domain.ErrNotFound
	}
	return red, nil
}

func (r *RewardRepo) UpdateRedemption(_ context.Context, red domain.Redemption) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.redemptions[red.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.redemptions[red.ID] = red
	return nil
}

func (r *RewardRepo) ListRedemptions(_ context.Context, tenantID, accountID uuid.UUID) ([]domain.Redemption, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Redemption
	for _, red := range r.S.redemptions {
		if red.TenantID == tenantID && red.AccountID == accountID {
			out = append(out, red)
		}
	}
	return out, nil
}

// --- ReferralRepo ---

type ReferralRepo struct{ S *Store }

func (r *ReferralRepo) CreateCode(_ context.Context, c domain.ReferralCode) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	code := strings.ToUpper(c.Code)
	c.Code = code
	if _, ok := r.S.refByCode[codeKey(c.TenantID, code)]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.refCodes[c.ID] = c
	r.S.refByCode[codeKey(c.TenantID, code)] = c.ID
	r.S.refByAcct[acctKey(c.TenantID, c.AccountID)] = c.ID
	return nil
}

func (r *ReferralRepo) GetCode(_ context.Context, tenantID uuid.UUID, code string) (domain.ReferralCode, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.refByCode[codeKey(tenantID, strings.ToUpper(code))]
	if !ok {
		return domain.ReferralCode{}, domain.ErrNotFound
	}
	return r.S.refCodes[id], nil
}

func (r *ReferralRepo) GetCodeByAccount(_ context.Context, tenantID, accountID uuid.UUID) (domain.ReferralCode, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.refByAcct[acctKey(tenantID, accountID)]
	if !ok {
		return domain.ReferralCode{}, domain.ErrNotFound
	}
	return r.S.refCodes[id], nil
}

func (r *ReferralRepo) CreateEvent(_ context.Context, e domain.ReferralEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.refByReferee[acctKey(e.TenantID, e.RefereeAccount)]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.refEvents[e.ID] = e
	r.S.refByReferee[acctKey(e.TenantID, e.RefereeAccount)] = e.ID
	return nil
}

func (r *ReferralRepo) GetEventByReferee(_ context.Context, tenantID, refereeAccount uuid.UUID) (domain.ReferralEvent, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.refByReferee[acctKey(tenantID, refereeAccount)]
	if !ok {
		return domain.ReferralEvent{}, domain.ErrNotFound
	}
	return r.S.refEvents[id], nil
}

func (r *ReferralRepo) UpdateEvent(_ context.Context, e domain.ReferralEvent) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.refEvents[e.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.refEvents[e.ID] = e
	return nil
}

func (r *ReferralRepo) CountCompletedByReferrer(_ context.Context, tenantID, referrerAccount uuid.UUID) (int64, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var n int64
	for _, e := range r.S.refEvents {
		if e.TenantID == tenantID && e.ReferrerAccount == referrerAccount && e.Status == domain.ReferralCompleted {
			n++
		}
	}
	return n, nil
}

// --- MissionRepo ---

type MissionRepo struct{ S *Store }

func (r *MissionRepo) CreateMission(_ context.Context, m domain.Mission) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.missions[m.ID] = m
	r.S.missionCode[codeKey(m.TenantID, m.Code)] = m.ID
	return nil
}

func (r *MissionRepo) GetMission(_ context.Context, tenantID, missionID uuid.UUID) (domain.Mission, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	m, ok := r.S.missions[missionID]
	if !ok || m.TenantID != tenantID {
		return domain.Mission{}, domain.ErrNotFound
	}
	return m, nil
}

func (r *MissionRepo) GetMissionByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Mission, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.missionCode[codeKey(tenantID, code)]
	if !ok {
		return domain.Mission{}, domain.ErrNotFound
	}
	return r.S.missions[id], nil
}

func (r *MissionRepo) GetProgress(_ context.Context, tenantID, accountID, missionID uuid.UUID) (domain.MissionProgress, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	p, ok := r.S.progress[progKey(tenantID, accountID, missionID)]
	if !ok {
		return domain.MissionProgress{}, domain.ErrNotFound
	}
	return p, nil
}

func (r *MissionRepo) UpsertProgress(_ context.Context, p domain.MissionProgress) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.progress[progKey(p.TenantID, p.AccountID, p.MissionID)] = p
	return nil
}

// --- AchievementRepo ---

type AchievementRepo struct{ S *Store }

func (r *AchievementRepo) CreateAchievement(_ context.Context, a domain.Achievement) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.achievements[a.ID] = a
	r.S.achByCode[codeKey(a.TenantID, a.Code)] = a.ID
	return nil
}

func (r *AchievementRepo) GetAchievement(_ context.Context, tenantID, id uuid.UUID) (domain.Achievement, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	a, ok := r.S.achievements[id]
	if !ok || a.TenantID != tenantID {
		return domain.Achievement{}, domain.ErrNotFound
	}
	return a, nil
}

func (r *AchievementRepo) GetAchievementByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Achievement, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.achByCode[codeKey(tenantID, code)]
	if !ok {
		return domain.Achievement{}, domain.ErrNotFound
	}
	return r.S.achievements[id], nil
}

func (r *AchievementRepo) ListAchievements(_ context.Context, tenantID uuid.UUID) ([]domain.Achievement, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Achievement
	for _, a := range r.S.achievements {
		if a.TenantID == tenantID {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *AchievementRepo) CreateUnlock(_ context.Context, u domain.AchievementUnlock) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := unlockKey(u.TenantID, u.AccountID, u.AchievementID)
	if _, ok := r.S.unlocks[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.unlocks[k] = u
	return nil
}

func (r *AchievementRepo) GetUnlock(_ context.Context, tenantID, accountID, achievementID uuid.UUID) (domain.AchievementUnlock, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	u, ok := r.S.unlocks[unlockKey(tenantID, accountID, achievementID)]
	if !ok {
		return domain.AchievementUnlock{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *AchievementRepo) ListUnlocks(_ context.Context, tenantID, accountID uuid.UUID) ([]domain.AchievementUnlock, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.AchievementUnlock
	for _, u := range r.S.unlocks {
		if u.TenantID == tenantID && u.AccountID == accountID {
			out = append(out, u)
		}
	}
	return out, nil
}

// --- StreakRepo ---

type StreakRepo struct{ S *Store }

func (r *StreakRepo) GetStreak(_ context.Context, tenantID, accountID uuid.UUID) (domain.Streak, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.streaks[acctKey(tenantID, accountID)]
	if !ok {
		return domain.Streak{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *StreakRepo) UpsertStreak(_ context.Context, s domain.Streak) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.streaks[acctKey(s.TenantID, s.AccountID)] = s
	return nil
}

// --- SpinRepo ---

type SpinRepo struct{ S *Store }

func (r *SpinRepo) CreateCampaign(_ context.Context, c domain.SpinCampaign) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.campaigns[c.ID] = c
	r.S.campByCode[codeKey(c.TenantID, c.Code)] = c.ID
	return nil
}

func (r *SpinRepo) GetCampaign(_ context.Context, tenantID, campaignID uuid.UUID) (domain.SpinCampaign, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	c, ok := r.S.campaigns[campaignID]
	if !ok || c.TenantID != tenantID {
		return domain.SpinCampaign{}, domain.ErrNotFound
	}
	return c, nil
}

func (r *SpinRepo) GetCampaignByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.SpinCampaign, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.campByCode[codeKey(tenantID, code)]
	if !ok {
		return domain.SpinCampaign{}, domain.ErrNotFound
	}
	return r.S.campaigns[id], nil
}

func (r *SpinRepo) CreateSpin(_ context.Context, s domain.SpinResult) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.spins = append(r.S.spins, s)
	return nil
}

// --- CollectibleRepo ---

type CollectibleRepo struct{ S *Store }

func (r *CollectibleRepo) CreateCollectible(_ context.Context, c domain.Collectible) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.collectibles[c.ID] = c
	return nil
}

func (r *CollectibleRepo) Grant(_ context.Context, o domain.OwnedCollectible) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.owned = append(r.S.owned, o)
	return nil
}

func (r *CollectibleRepo) ListOwned(_ context.Context, tenantID, accountID uuid.UUID) ([]domain.OwnedCollectible, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.OwnedCollectible
	for _, o := range r.S.owned {
		if o.TenantID == tenantID && o.AccountID == accountID {
			out = append(out, o)
		}
	}
	return out, nil
}

// --- CashbackRepo ---

type CashbackRepo struct{ S *Store }

func (r *CashbackRepo) CreateGrant(_ context.Context, g domain.CashbackGrant) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if g.IdempotencyKey != "" {
		k := idemKey(g.TenantID, g.IdempotencyKey)
		if _, ok := r.S.cashbackIdem[k]; ok {
			return domain.ErrAlreadyExists
		}
		r.S.cashbackIdem[k] = g.ID
	}
	r.S.cashbacks[g.ID] = g
	return nil
}

func (r *CashbackRepo) GetGrant(_ context.Context, tenantID, grantID uuid.UUID) (domain.CashbackGrant, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	g, ok := r.S.cashbacks[grantID]
	if !ok || g.TenantID != tenantID {
		return domain.CashbackGrant{}, domain.ErrNotFound
	}
	return g, nil
}

func (r *CashbackRepo) GetGrantByIdempotency(_ context.Context, tenantID uuid.UUID, key string) (domain.CashbackGrant, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.cashbackIdem[idemKey(tenantID, key)]
	if !ok {
		return domain.CashbackGrant{}, domain.ErrNotFound
	}
	return r.S.cashbacks[id], nil
}

func (r *CashbackRepo) UpdateGrant(_ context.Context, g domain.CashbackGrant) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.cashbacks[g.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.cashbacks[g.ID] = g
	return nil
}

// --- AIScoreRepo ---

type AIScoreRepo struct{ S *Store }

func (r *AIScoreRepo) Upsert(_ context.Context, s domain.AIScore) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.aiScores[acctKey(s.TenantID, s.AccountID)] = s
	return nil
}

func (r *AIScoreRepo) Get(_ context.Context, tenantID, accountID uuid.UUID) (domain.AIScore, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.aiScores[acctKey(tenantID, accountID)]
	if !ok {
		return domain.AIScore{}, domain.ErrNotFound
	}
	return s, nil
}

// --- OutboxRepo ---

type OutboxRepo struct{ S *Store }

func (r *OutboxRepo) Enqueue(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.outbox = append(r.S.outbox, m)
	return nil
}

func (r *OutboxRepo) Update(_ context.Context, m domain.OutboxMessage) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	for i := range r.S.outbox {
		if r.S.outbox[i].ID == m.ID {
			r.S.outbox[i] = m
			return nil
		}
	}
	return domain.ErrNotFound
}

func (r *OutboxRepo) ListPending(_ context.Context, limit int) ([]domain.OutboxMessage, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.OutboxMessage
	for _, m := range r.S.outbox {
		if m.Status == domain.OutboxStatusPending {
			out = append(out, m)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
