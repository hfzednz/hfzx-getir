package domain

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUpstream        = errors.New("upstream failure")
	ErrNotFound        = errors.New("not found")
	ErrConflict        = errors.New("conflict")
)

// SessionView is a BFF session projection (opaque tokens from IAM).
type SessionView struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	CustomerID   string `json:"customerId"`
	ExpiresIn    int    `json:"expiresIn"`
}

// HomeFeed aggregates search + recommendations.
type HomeFeed struct {
	Query       string           `json:"query"`
	Products    []map[string]any `json:"products"`
	Rails       []map[string]any `json:"rails"`
	Serviceable bool             `json:"serviceable"`
}

// CheckoutAddress is the delivery snapshot supplied by customer-web (location store).
type CheckoutAddress struct {
	Label   string  `json:"label"`
	Line1   string  `json:"line1"`
	City    string  `json:"city"`
	Country string  `json:"country"`
	Phone   string  `json:"phone"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
}

// Empty reports whether the address cannot satisfy checkout validation.
func (a CheckoutAddress) Empty() bool {
	if a.Line1 != "" {
		return false
	}
	return a.Lat == 0 && a.Lng == 0
}

// Map is the PATCH payload for checkout-service sessions.
func (a CheckoutAddress) Map() map[string]any {
	return map[string]any{
		"label": a.Label, "line1": a.Line1, "city": a.City,
		"country": a.Country, "phone": a.Phone, "lat": a.Lat, "lng": a.Lng,
	}
}

// CheckoutPreview is an edge checkout summary (money minor units).
type CheckoutPreview struct {
	SessionID     string `json:"sessionId"`
	CartID        string `json:"cartId"`
	Currency      string `json:"currency"`
	SubtotalMinor int64  `json:"subtotalMinor"`
	DiscountMinor int64  `json:"discountMinor"`
	TotalMinor    int64  `json:"totalMinor"`
	PaymentReady  bool   `json:"paymentReady"`
}

// OrderTrack is tracking projection for customer.
type OrderTrack struct {
	OrderID    string  `json:"orderId"`
	Status     string  `json:"status"`
	CourierID  string  `json:"courierId,omitempty"`
	ETASeconds int     `json:"etaSeconds"`
	Lat        float64 `json:"lat"`
	Lng        float64 `json:"lng"`
}
