package memory

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/domain"
)

// Store is the in-memory aggregate for loyalty-service.
type Store struct {
	mu sync.RWMutex

	accounts   map[uuid.UUID]domain.Account
	byPrin     map[string]uuid.UUID
	ledger     []domain.PointLedgerEntry
	ledgerIdem map[string]uuid.UUID
	ledgerOrd  map[string]uuid.UUID
	stats      map[string]int64
	audits     []domain.AuditEntry

	tiers        []domain.TierConfig
	memberships  map[string]domain.Membership
	rewards      map[uuid.UUID]domain.Reward
	rewardCodes  map[string]uuid.UUID
	redemptions  map[uuid.UUID]domain.Redemption

	refCodes     map[uuid.UUID]domain.ReferralCode
	refByCode    map[string]uuid.UUID
	refByAcct    map[string]uuid.UUID
	refEvents    map[uuid.UUID]domain.ReferralEvent
	refByReferee map[string]uuid.UUID

	missions   map[uuid.UUID]domain.Mission
	missionCode map[string]uuid.UUID
	progress   map[string]domain.MissionProgress

	achievements map[uuid.UUID]domain.Achievement
	achByCode    map[string]uuid.UUID
	unlocks      map[string]domain.AchievementUnlock

	streaks map[string]domain.Streak

	campaigns   map[uuid.UUID]domain.SpinCampaign
	campByCode  map[string]uuid.UUID
	spins       []domain.SpinResult

	collectibles map[uuid.UUID]domain.Collectible
	owned        []domain.OwnedCollectible

	cashbacks     map[uuid.UUID]domain.CashbackGrant
	cashbackIdem  map[string]uuid.UUID

	aiScores map[string]domain.AIScore
	outbox   []domain.OutboxMessage
	published []PublishedEvent
}

// PublishedEvent is a recorded EventPublisher call.
type PublishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore creates an empty memory store with default tiers.
func NewStore() *Store {
	return &Store{
		accounts:     make(map[uuid.UUID]domain.Account),
		byPrin:       make(map[string]uuid.UUID),
		ledgerIdem:   make(map[string]uuid.UUID),
		ledgerOrd:    make(map[string]uuid.UUID),
		stats:        make(map[string]int64),
		tiers:        domain.DefaultTiers(),
		memberships:  make(map[string]domain.Membership),
		rewards:      make(map[uuid.UUID]domain.Reward),
		rewardCodes:  make(map[string]uuid.UUID),
		redemptions:  make(map[uuid.UUID]domain.Redemption),
		refCodes:     make(map[uuid.UUID]domain.ReferralCode),
		refByCode:    make(map[string]uuid.UUID),
		refByAcct:    make(map[string]uuid.UUID),
		refEvents:    make(map[uuid.UUID]domain.ReferralEvent),
		refByReferee: make(map[string]uuid.UUID),
		missions:     make(map[uuid.UUID]domain.Mission),
		missionCode:  make(map[string]uuid.UUID),
		progress:     make(map[string]domain.MissionProgress),
		achievements: make(map[uuid.UUID]domain.Achievement),
		achByCode:    make(map[string]uuid.UUID),
		unlocks:      make(map[string]domain.AchievementUnlock),
		streaks:      make(map[string]domain.Streak),
		campaigns:    make(map[uuid.UUID]domain.SpinCampaign),
		campByCode:   make(map[string]uuid.UUID),
		collectibles: make(map[uuid.UUID]domain.Collectible),
		cashbacks:    make(map[uuid.UUID]domain.CashbackGrant),
		cashbackIdem: make(map[string]uuid.UUID),
		aiScores:     make(map[string]domain.AIScore),
	}
}

func prinKey(tenant, principal uuid.UUID) string {
	return tenant.String() + "|" + principal.String()
}

func idemKey(tenant uuid.UUID, key string) string {
	return tenant.String() + "|" + key
}

func acctKey(tenant, account uuid.UUID) string {
	return tenant.String() + "|" + account.String()
}

func codeKey(tenant uuid.UUID, code string) string {
	return tenant.String() + "|" + code
}

func progKey(tenant, account, mission uuid.UUID) string {
	return tenant.String() + "|" + account.String() + "|" + mission.String()
}

func unlockKey(tenant, account, ach uuid.UUID) string {
	return tenant.String() + "|" + account.String() + "|" + ach.String()
}

func orderKey(tenant, account, order uuid.UUID) string {
	return tenant.String() + "|" + account.String() + "|" + order.String()
}

func statKey(tenant, account uuid.UUID, key string) string {
	return tenant.String() + "|" + account.String() + "|" + key
}

// Clock is a fixed clock for tests.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time { return c.T }

// IDGen returns random UUIDs.
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }

// FixedRand returns a fixed Intn sequence for spin tests.
type FixedRand struct {
	Values []int
	i      int
}

func (r *FixedRand) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	if len(r.Values) == 0 {
		return 0
	}
	v := r.Values[r.i%len(r.Values)]
	r.i++
	if v < 0 {
		v = 0
	}
	return v % n
}
