package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/loyalty-service/internal/app"
	"github.com/nexora/loyalty-service/internal/app/memory"
	"github.com/nexora/loyalty-service/internal/domain"
)

type testEnv struct {
	Deps      *app.Deps
	Store     *memory.Store
	Wallet    *memory.WalletClient
	Repos     *memory.Repos
	Tenant    uuid.UUID
	Principal uuid.UUID
}

func testDeps(t *testing.T) *testEnv {
	t.Helper()
	store := memory.NewStore()
	repos := memory.NewRepos(store)
	wallet := &memory.WalletClient{}
	deps := &app.Deps{
		Accounts: repos.Accounts, Memberships: repos.Memberships, Rewards: repos.Rewards,
		Referrals: repos.Referrals, Missions: repos.Missions, Achievements: repos.Achievements,
		Streaks: repos.Streaks, Spins: repos.Spins, Collectibles: repos.Collectibles,
		Cashbacks: repos.Cashbacks, AIScores: repos.AIScores, Outbox: repos.Outbox,
		Wallet: wallet, Publisher: &memory.EventPublisher{S: store},
		Clock: &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)},
		IDs:   memory.IDGen{},
		Rand:  &memory.FixedRand{Values: []int{0}},
	}
	return &testEnv{
		Deps: deps, Store: store, Wallet: wallet, Repos: repos,
		Tenant:    uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Principal: uuid.MustParse("22222222-2222-2222-2222-222222222222"),
	}
}

func TestEarnThenRedeemPoints(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	orderID := uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	acct, _, err := env.Deps.EarnPoints(ctx, app.EarnPointsInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, OrderID: orderID, Points: 100,
	})
	if err != nil {
		t.Fatalf("earn: %v", err)
	}
	if acct.Points != 100 {
		t.Fatalf("points=%d", acct.Points)
	}
	acct, _, err = env.Deps.RedeemPoints(ctx, app.RedeemPointsInput{
		TenantID: env.Tenant, AccountID: acct.ID, Points: 40, IdempotencyKey: "r1",
	})
	if err != nil {
		t.Fatalf("redeem: %v", err)
	}
	if acct.Points != 60 {
		t.Fatalf("after redeem points=%d", acct.Points)
	}
}

func TestRedeemMoreThanBalanceFails(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	acct, _ := env.Deps.EnsureAccount(ctx, app.EnsureAccountInput{TenantID: env.Tenant, PrincipalID: env.Principal})
	_, _, err := env.Deps.EarnPoints(ctx, app.EarnPointsInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		OrderID: uuid.New(), Points: 10,
	})
	if err != nil {
		t.Fatalf("earn: %v", err)
	}
	_, _, err = env.Deps.RedeemPoints(ctx, app.RedeemPointsInput{
		TenantID: env.Tenant, AccountID: acct.ID, Points: 50, IdempotencyKey: "over",
	})
	if !errors.Is(err, domain.ErrInsufficientPoints) {
		t.Fatalf("expected insufficient points, got %v", err)
	}
}

func TestMembershipUpgradesAtThreshold(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	acct, _ := env.Deps.EnsureAccount(ctx, app.EnsureAccountInput{TenantID: env.Tenant, PrincipalID: env.Principal})

	_, _, err := env.Deps.EarnPoints(ctx, app.EarnPointsInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		OrderID: uuid.New(), Points: 1000,
	})
	if err != nil {
		t.Fatalf("earn: %v", err)
	}
	m, upgraded, err := env.Deps.EvaluateMembership(ctx, app.EvaluateMembershipInput{
		TenantID: env.Tenant, AccountID: acct.ID,
	})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if m.Tier != domain.TierSilver {
		t.Fatalf("tier=%s want silver", m.Tier)
	}
	if !upgraded && m.Tier != domain.TierSilver {
		t.Fatalf("expected upgrade to silver")
	}
}

func TestReferralSelfApplyRejected(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	code, err := env.Deps.CreateReferralCode(ctx, app.CreateReferralCodeInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, Code: "SELFME",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = env.Deps.ApplyReferral(ctx, app.ApplyReferralInput{
		TenantID: env.Tenant, RefereePrincipal: env.Principal, Code: code.Code,
	})
	if !errors.Is(err, domain.ErrSelfReferral) {
		t.Fatalf("expected self referral, got %v", err)
	}
}

func TestReferralCompleteGrantsPoints(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	referrer := env.Principal
	referee := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	code, err := env.Deps.CreateReferralCode(ctx, app.CreateReferralCodeInput{
		TenantID: env.Tenant, PrincipalID: referrer, Code: "FRIEND1",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	_, err = env.Deps.ApplyReferral(ctx, app.ApplyReferralInput{
		TenantID: env.Tenant, RefereePrincipal: referee, Code: code.Code,
	})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	ev, err := env.Deps.CompleteReferral(ctx, app.CompleteReferralInput{
		TenantID: env.Tenant, RefereePrincipal: referee, OrderID: uuid.New(), Points: 500,
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if ev.Status != domain.ReferralCompleted || ev.PointsGranted != 500 {
		t.Fatalf("event=%+v", ev)
	}
	refAcct, _ := env.Deps.GetAccountByPrincipal(ctx, env.Tenant, referrer)
	if refAcct.Points != 500 {
		t.Fatalf("referrer points=%d", refAcct.Points)
	}
}

func TestCashbackConfirmCallsWalletCredit(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	g, err := env.Deps.GrantCashback(ctx, app.GrantCashbackInput{
		TenantID: env.Tenant, PrincipalID: env.Principal,
		AmountMinor: 2500, Currency: "TRY", AccountType: "cashback",
		IdempotencyKey: "cb1",
	})
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if g.Status != domain.CashbackPending {
		t.Fatalf("status=%s", g.Status)
	}
	g, err = env.Deps.ConfirmCashbackToWallet(ctx, app.ConfirmCashbackInput{
		TenantID: env.Tenant, GrantID: g.ID,
	})
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if g.Status != domain.CashbackIssued {
		t.Fatalf("status=%s", g.Status)
	}
	if len(env.Wallet.Calls) != 1 {
		t.Fatalf("wallet calls=%d", len(env.Wallet.Calls))
	}
	if env.Wallet.Calls[0].AmountMinor != 2500 || env.Wallet.Calls[0].AccountType != "cashback" {
		t.Fatalf("credit req=%+v", env.Wallet.Calls[0])
	}
}

func TestMissionCompleteUnlocksAchievement(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	acct, _ := env.Deps.EnsureAccount(ctx, app.EnsureAccountInput{TenantID: env.Tenant, PrincipalID: env.Principal})

	achID := uuid.New()
	_ = env.Repos.Achievements.CreateAchievement(ctx, domain.Achievement{
		ID: achID, TenantID: env.Tenant, Code: "first_mission", Title: "First Mission",
		RuleType: domain.RuleMissionCode, Threshold: 1, Active: true, CreatedAt: time.Now(),
	})
	missionID := uuid.New()
	_ = env.Repos.Missions.CreateMission(ctx, domain.Mission{
		ID: missionID, TenantID: env.Tenant, Code: "buy_3", Title: "Buy 3",
		TargetCount: 3, RewardPoints: 50, Achievement: "first_mission", Active: true,
		CreatedAt: time.Now(),
	})

	_, err := env.Deps.TrackMission(ctx, app.TrackMissionInput{
		TenantID: env.Tenant, AccountID: acct.ID, MissionID: missionID, Delta: 2,
	})
	if err != nil {
		t.Fatalf("track: %v", err)
	}
	prog, err := env.Deps.TrackMission(ctx, app.TrackMissionInput{
		TenantID: env.Tenant, AccountID: acct.ID, MissionID: missionID, Delta: 1,
	})
	if err != nil {
		t.Fatalf("track complete: %v", err)
	}
	if prog.Status != domain.MissionCompleted {
		t.Fatalf("status=%s", prog.Status)
	}
	u, err := env.Repos.Achievements.GetUnlock(ctx, env.Tenant, acct.ID, achID)
	if err != nil {
		t.Fatalf("unlock missing: %v", err)
	}
	if u.Code != "first_mission" {
		t.Fatalf("code=%s", u.Code)
	}
	fresh, _ := env.Deps.GetAccount(ctx, env.Tenant, acct.ID)
	if fresh.Points != 50 {
		t.Fatalf("reward points=%d", fresh.Points)
	}
}

func TestStreakIncrementsAndBreakResets(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	acct, _ := env.Deps.EnsureAccount(ctx, app.EnsureAccountInput{TenantID: env.Tenant, PrincipalID: env.Principal})

	s, err := env.Deps.UpdateStreak(ctx, app.UpdateStreakInput{
		TenantID: env.Tenant, AccountID: acct.ID, Action: "increment", Date: "2026-08-01",
	})
	if err != nil || s.CurrentCount != 1 {
		t.Fatalf("day1: %+v err=%v", s, err)
	}
	s, err = env.Deps.UpdateStreak(ctx, app.UpdateStreakInput{
		TenantID: env.Tenant, AccountID: acct.ID, Action: "increment", Date: "2026-08-02",
	})
	if err != nil || s.CurrentCount != 2 {
		t.Fatalf("day2: %+v err=%v", s, err)
	}
	s, err = env.Deps.UpdateStreak(ctx, app.UpdateStreakInput{
		TenantID: env.Tenant, AccountID: acct.ID, Action: "break", Date: "2026-08-03",
	})
	if err != nil {
		t.Fatalf("break: %v", err)
	}
	if s.CurrentCount != 0 || !s.Broken {
		t.Fatalf("after break: %+v", s)
	}
}

func TestSpinRespectsWeights(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	acct, _ := env.Deps.EnsureAccount(ctx, app.EnsureAccountInput{TenantID: env.Tenant, PrincipalID: env.Principal})

	// weights: A=1, B=9 → total 10. FixedRand returns 0 → always A.
	env.Deps.Rand = &memory.FixedRand{Values: []int{0}}
	campID := uuid.New()
	_ = env.Repos.Spins.CreateCampaign(ctx, domain.SpinCampaign{
		ID: campID, TenantID: env.Tenant, Code: "daily", Title: "Daily",
		Prizes: []domain.SpinPrize{
			{Code: "A", Weight: 1, Points: 10},
			{Code: "B", Weight: 9, Points: 100},
		},
		Active: true, CreatedAt: time.Now(),
	})
	res, err := env.Deps.Spin(ctx, app.SpinInput{
		TenantID: env.Tenant, AccountID: acct.ID, CampaignID: campID,
	})
	if err != nil {
		t.Fatalf("spin: %v", err)
	}
	if res.PrizeCode != "A" || res.PointsWon != 10 {
		t.Fatalf("expected A/10 got %s/%d", res.PrizeCode, res.PointsWon)
	}

	// roll=1 still in A (weight 1, roll 0 only); roll=1 → B
	env.Deps.Rand = &memory.FixedRand{Values: []int{1}}
	res, err = env.Deps.Spin(ctx, app.SpinInput{
		TenantID: env.Tenant, AccountID: acct.ID, CampaignID: campID,
	})
	if err != nil {
		t.Fatalf("spin2: %v", err)
	}
	if res.PrizeCode != "B" {
		t.Fatalf("expected B got %s", res.PrizeCode)
	}
}

func TestIdempotentEarnPointsSameOrderID(t *testing.T) {
	env := testDeps(t)
	ctx := context.Background()
	orderID := uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb")

	acct1, e1, err := env.Deps.EarnPoints(ctx, app.EarnPointsInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, OrderID: orderID, Points: 75,
	})
	if err != nil {
		t.Fatalf("earn1: %v", err)
	}
	acct2, e2, err := env.Deps.EarnPoints(ctx, app.EarnPointsInput{
		TenantID: env.Tenant, PrincipalID: env.Principal, OrderID: orderID, Points: 75,
	})
	if err != nil {
		t.Fatalf("earn2: %v", err)
	}
	if e1.ID != e2.ID {
		t.Fatalf("entry ids differ: %s vs %s", e1.ID, e2.ID)
	}
	if acct2.Points != 75 || acct1.Points != 75 {
		t.Fatalf("points doubled: %d", acct2.Points)
	}
}
