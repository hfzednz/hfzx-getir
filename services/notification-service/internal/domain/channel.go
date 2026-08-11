package domain

// Channel is a delivery channel.
type Channel string

const (
	ChannelPush     Channel = "push"
	ChannelEmail    Channel = "email"
	ChannelSMS      Channel = "sms"
	ChannelWhatsApp Channel = "whatsapp"
	ChannelInApp    Channel = "in_app"
	ChannelWeb      Channel = "web"
)

// Valid reports whether the channel is known.
func (c Channel) Valid() bool {
	switch c {
	case ChannelPush, ChannelEmail, ChannelSMS, ChannelWhatsApp, ChannelInApp, ChannelWeb:
		return true
	default:
		return false
	}
}

// Priority classifies message urgency / compliance tier.
type Priority string

const (
	PriorityOTP           Priority = "otp"
	PriorityTransactional Priority = "transactional"
	PrioritySystem        Priority = "system"
	PriorityMarketing     Priority = "marketing"
)

// Valid reports whether the priority is known.
func (p Priority) Valid() bool {
	switch p {
	case PriorityOTP, PriorityTransactional, PrioritySystem, PriorityMarketing:
		return true
	default:
		return false
	}
}

// OverridesMarketingOptOut is true for transactional/otp/system.
func (p Priority) OverridesMarketingOptOut() bool {
	switch p {
	case PriorityOTP, PriorityTransactional, PrioritySystem:
		return true
	default:
		return false
	}
}

// RespectsQuietHours is true for marketing (and similar promotional) traffic.
func (p Priority) RespectsQuietHours() bool {
	return p == PriorityMarketing
}
