package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/identity-service/internal/app"
	"github.com/nexora/identity-service/internal/app/memory"
	"github.com/nexora/identity-service/internal/app/ports"
	"github.com/nexora/identity-service/internal/domain"
	"github.com/nexora/identity-service/internal/domain/policy"
	"github.com/nexora/identity-service/internal/security/jwt"
	"github.com/nexora/identity-service/internal/security/password"
	"github.com/nexora/identity-service/internal/security/webauthn"
)

func testDeps(t *testing.T) (*app.Deps, *memory.Store, *memory.OTPSender) {
	t.Helper()
	store := memory.NewStore()
	otp := &memory.OTPSender{}
	events := &memory.EventPublisher{}
	clock := &memory.Clock{T: time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)}
	km := jwt.NewKeyManager()
	if err := km.Generate(2048); err != nil {
		t.Fatalf("jwt generate: %v", err)
	}

	d := &app.Deps{
		Principals: &memory.PrincipalRepo{S: store},
		Sessions:   &memory.SessionRepo{S: store},
		Devices:    &memory.DeviceRepo{S: store},
		Roles:      &memory.RoleRepo{S: store},
		Audit:      &memory.AuditRepo{S: store},
		OAuth:      &memory.OAuthRepo{S: store},
		Risk:       &memory.RiskRepo{S: store},
		OTP:        otp,
		Events:     events,
		Clock:      clock,
		IDs:        memory.IDGen{},
		Passwords:  password.NewDefaultHasher(),
		WebAuthn:   &webauthn.StubService{},
		JWTKeys:    km,
		Issuer:     "test-issuer",
		Audience:   "test-aud",
		AccessTTL:  15 * time.Minute,
		OTPPepper:  "test-pepper",
		Social:     app.StubSocialProviders(),
	}
	d.Tokens = &app.DefaultTokenIssuer{Deps: d}

	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	(&memory.RiskRepo{S: store}).SeedPolicy(tenant, domain.SecurityPolicy{
		ID:                     uuid.New(),
		Name:                   "test",
		Enabled:                true,
		PasswordMinLength:      12,
		PasswordRequireUpper:   true,
		PasswordRequireLower:   true,
		PasswordRequireDigit:   true,
		PasswordRequireSymbol:  true,
		PasswordHistoryCount:   3,
		MFARequiredAboveRisk:   40,
		SessionIdleSeconds:     1800,
		SessionAbsoluteSeconds: 43200,
		RefreshTokenSeconds:    86400,
		MaxFailedAttempts:      3,
		LockoutSeconds:         900,
		BlockAboveRisk:         90,
	})
	return d, store, otp
}

func TestVerifyOTP_HappyPath(t *testing.T) {
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name    string
		phone   string
		wantErr bool
	}{
		{name: "new principal by phone", phone: "+905551112233"},
		{name: "existing principal reuse", phone: "+905559998877"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, store, otpSender := testDeps(t)
			ctx := context.Background()

			if tt.name == "existing principal reuse" {
				now := d.Clock.Now()
				p := domain.Principal{
					ID: uuid.New(), TenantID: tenant, Kind: domain.PrincipalKindUser,
					Status: domain.PrincipalStatusActive, CreatedAt: now, UpdatedAt: now,
				}
				_ = d.Principals.Create(ctx, p)
				_ = d.Principals.CreateIdentifier(ctx, domain.Identifier{
					ID: uuid.New(), PrincipalID: p.ID, TenantID: tenant,
					Type: domain.IdentifierTypePhone, Value: tt.phone, CreatedAt: now, UpdatedAt: now,
				})
			}

			chalID, err := d.StartOTP(ctx, app.StartOTPInput{TenantID: tenant, Phone: tt.phone})
			if err != nil {
				t.Fatalf("StartOTP: %v", err)
			}
			if otpSender.LastCode == "" {
				t.Fatal("expected OTP code to be sent")
			}

			res, err := d.VerifyOTP(ctx, app.VerifyOTPInput{
				ChallengeID: chalID,
				Code:        otpSender.LastCode,
			})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("VerifyOTP: %v", err)
			}
			if res.Tokens.AccessToken == "" || res.Tokens.RefreshToken == "" {
				t.Fatal("expected token pair")
			}
			if res.Principal.ID == uuid.Nil {
				t.Fatal("expected principal")
			}
			if res.Session.ID == uuid.Nil {
				t.Fatal("expected session")
			}
			if _, ok := store.OTP[chalID]; ok {
				t.Fatal("otp challenge should be deleted")
			}
		})
	}
}

func TestLogin_FailLockoutHint(t *testing.T) {
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	email := "user@example.com"
	goodPass := "Str0ng!Passw0rd"

	tests := []struct {
		name          string
		failuresFirst int
		password      string
		wantLocked    bool
	}{
		{name: "single bad password", failuresFirst: 0, password: "Wrong!Pass1", wantLocked: false},
		{name: "lockout on threshold", failuresFirst: 2, password: "Wrong!Pass1", wantLocked: true},
		{name: "already at lockout", failuresFirst: 3, password: goodPass, wantLocked: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _, _ := testDeps(t)
			ctx := context.Background()
			now := d.Clock.Now()

			p := domain.Principal{
				ID: uuid.New(), TenantID: tenant, Kind: domain.PrincipalKindUser,
				Status: domain.PrincipalStatusActive, CreatedAt: now, UpdatedAt: now,
			}
			_ = d.Principals.Create(ctx, p)
			_ = d.Principals.CreateIdentifier(ctx, domain.Identifier{
				ID: uuid.New(), PrincipalID: p.ID, TenantID: tenant,
				Type: domain.IdentifierTypeEmail, Value: email, CreatedAt: now, UpdatedAt: now,
			})
			hash, err := d.Passwords.Hash(goodPass)
			if err != nil {
				t.Fatal(err)
			}
			_ = d.Principals.UpsertCredential(ctx, domain.Credential{
				ID: uuid.New(), PrincipalID: p.ID, PasswordHash: hash,
				Algorithm: domain.CredentialAlgorithmArgon2id, PasswordChangedAt: now,
				CreatedAt: now, UpdatedAt: now,
			})

			for i := 0; i < tt.failuresFirst; i++ {
				_ = d.Risk.AppendLoginAttempt(ctx, domain.LoginAttempt{
					ID: uuid.New(), TenantID: &tenant, PrincipalID: &p.ID,
					Identifier: email, Result: domain.LoginAttemptInvalidCredentials,
					CreatedAt: now,
				})
			}

			_, err = d.Login(ctx, app.LoginInput{
				TenantID: tenant, Email: email, Password: tt.password,
			})
			if tt.wantLocked {
				if err == nil {
					t.Fatal("expected lockout error")
				}
				if !errors.Is(err, domain.ErrPrincipalLocked) {
					t.Fatalf("want ErrPrincipalLocked, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected unauthorized for bad password")
			}
			if errors.Is(err, domain.ErrPrincipalLocked) {
				t.Fatalf("did not expect lockout yet: %v", err)
			}
			if !errors.Is(err, domain.ErrUnauthorized) {
				t.Fatalf("want ErrUnauthorized, got %v", err)
			}
		})
	}
}

func TestRefresh_RotationAndReuseDetection(t *testing.T) {
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name string
		run  func(t *testing.T, d *app.Deps)
	}{
		{
			name: "rotate once succeeds",
			run: func(t *testing.T, d *app.Deps) {
				ctx := context.Background()
				res := mustRegister(t, d, tenant, "rotate@example.com")
				next, err := d.Refresh(ctx, res.Tokens.RefreshToken)
				if err != nil {
					t.Fatalf("Refresh: %v", err)
				}
				if next.RefreshToken == "" || next.RefreshToken == res.Tokens.RefreshToken {
					t.Fatal("expected new refresh token")
				}
				if next.AccessToken == "" {
					t.Fatal("expected new access token")
				}
			},
		},
		{
			name: "reuse of rotated refresh revokes family",
			run: func(t *testing.T, d *app.Deps) {
				ctx := context.Background()
				res := mustRegister(t, d, tenant, "reuse@example.com")
				old := res.Tokens.RefreshToken
				next, err := d.Refresh(ctx, old)
				if err != nil {
					t.Fatalf("first Refresh: %v", err)
				}
				_, err = d.Refresh(ctx, old)
				if !errors.Is(err, domain.ErrTokenReuse) {
					t.Fatalf("want ErrTokenReuse, got %v", err)
				}
				// New token from rotation should also be dead after family revoke.
				_, err = d.Refresh(ctx, next.RefreshToken)
				if err == nil {
					t.Fatal("expected refresh after family revoke to fail")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, _, _ := testDeps(t)
			tt.run(t, d)
		})
	}
}

func TestCheckPermission_AllowDeny(t *testing.T) {
	tenant := uuid.MustParse("11111111-1111-1111-1111-111111111111")

	tests := []struct {
		name       string
		permission string
		assignRole bool
		maxRisk    float64
		riskScore  float64
		wantAllow  bool
	}{
		{name: "allow with role", permission: "orders:read", assignRole: true, maxRisk: 100, riskScore: 10, wantAllow: true},
		{name: "deny without role", permission: "orders:read", assignRole: false, maxRisk: 100, riskScore: 10, wantAllow: false},
		{name: "deny high risk abac", permission: "orders:read", assignRole: true, maxRisk: 20, riskScore: 50, wantAllow: false},
		{name: "deny wrong permission", permission: "orders:write", assignRole: true, maxRisk: 100, riskScore: 0, wantAllow: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, store, _ := testDeps(t)
			ctx := context.Background()
			now := d.Clock.Now()

			p := domain.Principal{
				ID: uuid.New(), TenantID: tenant, Kind: domain.PrincipalKindUser,
				Status: domain.PrincipalStatusActive, CreatedAt: now, UpdatedAt: now,
			}
			_ = d.Principals.Create(ctx, p)

			permID := uuid.New()
			perm := domain.Permission{ID: permID, Resource: "orders", Action: "read", CreatedAt: now}
			roleID := uuid.New()
			role := domain.Role{
				ID: roleID, TenantID: &tenant, Name: "customer", Kind: domain.RoleKindTenant, CreatedAt: now, UpdatedAt: now,
			}
			roles := &memory.RoleRepo{S: store}
			roles.SeedRole(role, nil, []domain.Permission{perm})

			if tt.assignRole {
				_ = d.Roles.AssignRole(ctx, domain.PrincipalRole{
					ID: uuid.New(), PrincipalID: p.ID, RoleID: roleID, CreatedAt: now,
				})
			}

			dec, err := d.CheckPermission(ctx, app.CheckPermissionInput{
				PrincipalID: p.ID,
				Permission:  tt.permission,
				Attrs: policy.Attributes{
					TenantID:  &tenant,
					RiskScore: tt.riskScore,
					MFALevel:  1,
				},
				Requirement: policy.Requirement{MaxRisk: tt.maxRisk},
			})
			if err != nil {
				t.Fatalf("CheckPermission: %v", err)
			}
			if dec.Allow != tt.wantAllow {
				t.Fatalf("Allow=%v want %v reasons=%v", dec.Allow, tt.wantAllow, dec.Reasons)
			}
		})
	}
}

func mustRegister(t *testing.T, d *app.Deps, tenant uuid.UUID, email string) app.AuthResult {
	t.Helper()
	res, err := d.Register(context.Background(), app.RegisterInput{
		TenantID: tenant, Email: email, Password: "Str0ng!Passw0rd", DisplayName: "Test",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	return res
}

// Ensure ports compile against memory.
var (
	_ ports.PrincipalRepository = (*memory.PrincipalRepo)(nil)
	_ ports.SessionRepository   = (*memory.SessionRepo)(nil)
)
