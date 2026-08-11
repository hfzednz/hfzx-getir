// Package risk scores authentication signals and decides adaptive MFA.
package risk

// Signal identifies a risk indicator.
type Signal string

const (
	SignalNewDevice              Signal = "new_device"
	SignalImpossibleTravel       Signal = "impossible_travel"
	SignalVPN                    Signal = "vpn"
	SignalTOR                    Signal = "tor"
	SignalIPReputation           Signal = "ip_reputation"
	SignalLocationAnomaly        Signal = "location_anomaly"
	SignalBehaviorAnomaly        Signal = "behavior_anomaly"
	SignalDeviceAnomaly          Signal = "device_anomaly"
	SignalCredentialStuffingHint Signal = "credential_stuffing_hint"
)

// defaultWeights maps signals to additive score points (capped at 100).
var defaultWeights = map[Signal]int{
	SignalNewDevice:              15,
	SignalImpossibleTravel:       40,
	SignalVPN:                    10,
	SignalTOR:                    35,
	SignalIPReputation:           30,
	SignalLocationAnomaly:        20,
	SignalBehaviorAnomaly:        25,
	SignalDeviceAnomaly:          20,
	SignalCredentialStuffingHint: 45,
}

// Scorer computes risk scores from signals.
type Scorer struct {
	Weights map[Signal]int
}

// NewScorer returns a Scorer with default weights.
func NewScorer() *Scorer {
	w := make(map[Signal]int, len(defaultWeights))
	for k, v := range defaultWeights {
		w[k] = v
	}
	return &Scorer{Weights: w}
}

// Score returns an integer risk score in [0, 100] for the given signals.
// Duplicate signals are counted once. Unknown signals are ignored.
func (s *Scorer) Score(signals []Signal) int {
	if s == nil {
		s = NewScorer()
	}
	seen := make(map[Signal]struct{}, len(signals))
	total := 0
	for _, sig := range signals {
		if _, ok := seen[sig]; ok {
			continue
		}
		seen[sig] = struct{}{}
		w, ok := s.Weights[sig]
		if !ok {
			continue
		}
		total += w
	}
	if total < 0 {
		return 0
	}
	if total > 100 {
		return 100
	}
	return total
}

// Score is a package-level convenience using default weights.
func Score(signals []Signal) int {
	return NewScorer().Score(signals)
}

// MFAPolicy configures adaptive MFA thresholds.
type MFAPolicy struct {
	// Threshold is the inclusive score at which MFA becomes required.
	Threshold int
	// AlwaysRequire forces MFA regardless of score.
	AlwaysRequire bool
	// NeverRequire disables adaptive MFA (score still computed by callers).
	NeverRequire bool
}

// DefaultMFAPolicy requires MFA at score >= 40.
func DefaultMFAPolicy() MFAPolicy {
	return MFAPolicy{Threshold: 40}
}

// AdaptiveMFARequired reports whether MFA should be challenged for the score and policy.
func AdaptiveMFARequired(score int, policy MFAPolicy) bool {
	if policy.NeverRequire {
		return false
	}
	if policy.AlwaysRequire {
		return true
	}
	threshold := policy.Threshold
	if threshold <= 0 {
		threshold = 40
	}
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return score >= threshold
}
