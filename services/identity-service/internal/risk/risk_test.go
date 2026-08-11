package risk_test

import (
	"testing"

	"github.com/nexora/identity-service/internal/risk"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name    string
		signals []risk.Signal
		wantMin int
		wantMax int
		want    int // if >= 0, exact
	}{
		{name: "empty", signals: nil, want: 0},
		{name: "new device only", signals: []risk.Signal{risk.SignalNewDevice}, want: 15},
		{
			name: "vpn+tor",
			signals: []risk.Signal{
				risk.SignalVPN,
				risk.SignalTOR,
			},
			want: 45, // 10+35
		},
		{
			name: "duplicates counted once",
			signals: []risk.Signal{
				risk.SignalNewDevice,
				risk.SignalNewDevice,
			},
			want: 15,
		},
		{
			name: "caps at 100",
			signals: []risk.Signal{
				risk.SignalImpossibleTravel,
				risk.SignalTOR,
				risk.SignalIPReputation,
				risk.SignalCredentialStuffingHint,
				risk.SignalBehaviorAnomaly,
				risk.SignalDeviceAnomaly,
				risk.SignalLocationAnomaly,
				risk.SignalNewDevice,
				risk.SignalVPN,
			},
			want: 100,
		},
		{
			name:    "unknown ignored",
			signals: []risk.Signal{risk.Signal("not_a_real_signal")},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := risk.Score(tt.signals)
			if tt.want >= 0 {
				if got != tt.want {
					t.Fatalf("got %d want %d", got, tt.want)
				}
				return
			}
			if got < tt.wantMin || got > tt.wantMax {
				t.Fatalf("got %d want [%d,%d]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestAdaptiveMFARequired(t *testing.T) {
	tests := []struct {
		name   string
		score  int
		policy risk.MFAPolicy
		want   bool
	}{
		{name: "below default", score: 20, policy: risk.DefaultMFAPolicy(), want: false},
		{name: "at threshold", score: 40, policy: risk.DefaultMFAPolicy(), want: true},
		{name: "above", score: 80, policy: risk.DefaultMFAPolicy(), want: true},
		{name: "always", score: 0, policy: risk.MFAPolicy{AlwaysRequire: true}, want: true},
		{name: "never", score: 100, policy: risk.MFAPolicy{NeverRequire: true, Threshold: 10}, want: false},
		{name: "custom threshold", score: 50, policy: risk.MFAPolicy{Threshold: 60}, want: false},
		{name: "custom met", score: 60, policy: risk.MFAPolicy{Threshold: 60}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := risk.AdaptiveMFARequired(tt.score, tt.policy)
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestCustomWeights(t *testing.T) {
	s := &risk.Scorer{Weights: map[risk.Signal]int{
		risk.SignalVPN: 5,
	}}
	if got := s.Score([]risk.Signal{risk.SignalVPN, risk.SignalTOR}); got != 5 {
		t.Fatalf("got %d", got)
	}
}
