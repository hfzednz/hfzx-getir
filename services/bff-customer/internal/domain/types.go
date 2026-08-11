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
	AccessToken  string
	RefreshToken string
	CustomerID   string
	ExpiresIn    int
}

// HomeFeed aggregates search + recommendations.
type HomeFeed struct {
	Query         string
	Products      []map[string]any
	Rails         []map[string]any
	Serviceable   bool
}

// CheckoutPreview is an edge checkout summary (money minor units).
type CheckoutPreview struct {
	SessionID     string
	CartID        string
	Currency      string
	SubtotalMinor int64
	DiscountMinor int64
	TotalMinor    int64
	PaymentReady  bool
}

// OrderTrack is tracking projection for customer.
type OrderTrack struct {
	OrderID   string
	Status    string
	CourierID string
	ETASeconds int
	Lat       float64
	Lng       float64
}
