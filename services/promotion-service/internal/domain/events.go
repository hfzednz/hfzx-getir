package domain

// Kafka / domain event topics for promotion-service.
const (
	TopicCampaign = "promo.campaign"
	TopicCoupon   = "promo.coupon"
	TopicVoucher  = "promo.voucher"
	TopicRule     = "promo.rule"
)

// Event type names.
const (
	EventCampaignCreated   = "CampaignCreated"
	EventCampaignActivated = "CampaignActivated"
	EventCampaignPaused    = "CampaignPaused"
	EventCampaignExpired   = "CampaignExpired"
	EventCouponGenerated   = "CouponGenerated"
	EventCouponRedeemed    = "CouponRedeemed"
	EventVoucherIssued     = "VoucherIssued"
	EventVoucherRedeemed   = "VoucherRedeemed"
	EventPromotionApplied  = "PromotionApplied"
)

// TopicForEvent maps event type to Kafka topic.
func TopicForEvent(eventType string) string {
	switch eventType {
	case EventCampaignCreated, EventCampaignActivated, EventCampaignPaused, EventCampaignExpired:
		return TopicCampaign
	case EventCouponGenerated, EventCouponRedeemed:
		return TopicCoupon
	case EventVoucherIssued, EventVoucherRedeemed:
		return TopicVoucher
	case EventPromotionApplied:
		return TopicRule
	default:
		return TopicCampaign
	}
}
