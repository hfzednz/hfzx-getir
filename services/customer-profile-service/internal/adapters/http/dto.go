package httpadapter

import (
	"github.com/nexora/customer-profile-service/internal/domain"
)

func profileDTO(p domain.CustomerProfile) map[string]any {
	out := map[string]any{
		"id":          p.ID,
		"principalId": p.PrincipalID,
		"tenantId":    p.TenantID,
		"displayName": p.DisplayName,
		"fullName":    p.FullName,
		"nickname":    p.Nickname,
		"avatarUrl":   p.AvatarURL,
		"gender":      p.Gender,
		"language":    p.Language,
		"countryCode": p.CountryCode,
		"city":        p.City,
		"timezone":    p.Timezone,
		"occupation":  p.Occupation,
		"familySize":  p.FamilySize,
		"dietary":     p.Dietary,
		"accessibility": p.Accessibility,
		"status":      p.Status,
		"createdAt":   p.CreatedAt,
		"updatedAt":   p.UpdatedAt,
	}
	if p.Birthday != nil {
		out["birthday"] = p.Birthday.Format("2006-01-02")
	}
	if p.DeletedAt != nil {
		out["deletedAt"] = p.DeletedAt
	}
	return out
}

func addressDTO(a domain.Address) map[string]any {
	out := map[string]any{
		"id":          a.ID,
		"profileId":   a.ProfileID,
		"tenantId":    a.TenantID,
		"label":       a.Label,
		"customLabel": a.CustomLabel,
		"line1":       a.Line1,
		"building":    a.Building,
		"apartment":   a.Apartment,
		"entrance":    a.Entrance,
		"floor":       a.Floor,
		"door":        a.Door,
		"notes":       a.Notes,
		"lat":         a.Lat,
		"lng":         a.Lng,
		"isDefault":   a.IsDefault,
		"createdAt":   a.CreatedAt,
		"updatedAt":   a.UpdatedAt,
	}
	if a.CityID != nil {
		out["cityId"] = a.CityID
	}
	if a.ZoneValidatedAt != nil {
		out["zoneValidatedAt"] = a.ZoneValidatedAt
	}
	return out
}

func prefsDTO(p domain.Preferences) map[string]any {
	return map[string]any{
		"profileId":          p.ProfileID,
		"favoriteBrands":     p.FavoriteBrands,
		"favoriteCategories": p.FavoriteCategories,
		"favoriteProducts":   p.FavoriteProducts,
		"favoriteStores":     p.FavoriteStores,
		"delivery":           p.Delivery,
		"payment":            p.Payment,
		"notification":       p.Notification,
		"shopping":           p.Shopping,
		"theme":              p.Theme,
		"language":           p.Language,
		"accessibility":      p.Accessibility,
		"createdAt":          p.CreatedAt,
		"updatedAt":          p.UpdatedAt,
	}
}

func tagAssignDTO(pt domain.ProfileTag) map[string]any {
	out := map[string]any{
		"profileId":  pt.ProfileID,
		"tagId":      pt.TagID,
		"assignedAt": pt.AssignedAt,
		"note":       pt.Note,
	}
	if pt.AssignedBy != nil {
		out["assignedBy"] = pt.AssignedBy
	}
	return out
}

func consentDTO(c domain.Consent) map[string]any {
	return map[string]any{
		"id":         c.ID,
		"profileId":  c.ProfileID,
		"tenantId":   c.TenantID,
		"channel":    c.Channel,
		"granted":    c.Granted,
		"source":     c.Source,
		"recordedAt": c.RecordedAt,
		"createdAt":  c.CreatedAt,
		"updatedAt":  c.UpdatedAt,
	}
}

func householdDTO(h domain.Household) map[string]any {
	return map[string]any{
		"id":             h.ID,
		"tenantId":       h.TenantID,
		"name":           h.Name,
		"ownerProfileId": h.OwnerProfileID,
		"sharing": map[string]any{
			"addresses": h.Sharing.Addresses,
			"payments":  h.Sharing.Payments,
			"lists":     h.Sharing.Lists,
			"wallet":    h.Sharing.Wallet,
			"loyalty":   h.Sharing.Loyalty,
		},
		"createdAt": h.CreatedAt,
		"updatedAt": h.UpdatedAt,
	}
}

func noteDTO(n domain.CRMNote) map[string]any {
	return map[string]any{
		"id":        n.ID,
		"profileId": n.ProfileID,
		"tenantId":  n.TenantID,
		"authorId":  n.AuthorID,
		"body":      n.Body,
		"pinned":    n.Pinned,
		"createdAt": n.CreatedAt,
		"updatedAt": n.UpdatedAt,
	}
}

func timelineDTO(e domain.TimelineEvent) map[string]any {
	out := map[string]any{
		"id":         e.ID,
		"profileId":  e.ProfileID,
		"tenantId":   e.TenantID,
		"type":       e.Type,
		"payload":    e.Payload,
		"occurredAt": e.OccurredAt,
		"createdAt":  e.CreatedAt,
	}
	if e.ActorID != nil {
		out["actorId"] = e.ActorID
	}
	return out
}

func segmentDTO(s domain.Segment) map[string]any {
	return map[string]any{
		"id":          s.ID,
		"tenantId":    s.TenantID,
		"name":        s.Name,
		"kind":        s.Kind,
		"description": s.Description,
		"rules":       s.Rules,
		"active":      s.Active,
		"createdAt":   s.CreatedAt,
		"updatedAt":   s.UpdatedAt,
	}
}

func membershipDTO(m domain.SegmentMembership) map[string]any {
	return map[string]any{
		"segmentId": m.SegmentID,
		"profileId": m.ProfileID,
		"joinedAt":  m.JoinedAt,
		"source":    m.Source,
	}
}

func personalizationDTO(p domain.Personalization) map[string]any {
	return map[string]any{
		"profileId":       p.ProfileID,
		"homepage":        p.Homepage,
		"category":        p.Category,
		"recommendation":  p.Recommendation,
		"search":          p.Search,
		"delivery":        p.Delivery,
		"promotion":       p.Promotion,
		"shoppingHabits":  p.ShoppingHabits,
		"createdAt":       p.CreatedAt,
		"updatedAt":       p.UpdatedAt,
	}
}

func aiModelDTO(m domain.AICustomerModel) map[string]any {
	return map[string]any{
		"profileId":              m.ProfileID,
		"frequency":              m.Frequency,
		"avgOrderValueMinor":     m.AvgOrderValueMinor,
		"churnProb":              m.ChurnProb,
		"preferredOrderHours":    m.PreferredOrderHours,
		"preferredOrderWeekdays": m.PreferredOrderWeekdays,
		"priceSensitivity":       m.PriceSensitivity,
		"brandAffinity":          m.BrandAffinity,
		"categoryAffinity":       m.CategoryAffinity,
		"modelVersion":           m.ModelVersion,
		"createdAt":              m.CreatedAt,
		"updatedAt":              m.UpdatedAt,
	}
}

func privacyDTO(r domain.PrivacyRequest) map[string]any {
	out := map[string]any{
		"id":         r.ID,
		"profileId":  r.ProfileID,
		"tenantId":   r.TenantID,
		"kind":       r.Kind,
		"status":     r.Status,
		"payloadRef": r.PayloadRef,
		"createdAt":  r.CreatedAt,
		"updatedAt":  r.UpdatedAt,
	}
	if r.CompletedAt != nil {
		out["completedAt"] = r.CompletedAt
	}
	return out
}
