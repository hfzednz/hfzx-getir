package ports

import (
	"context"

	"github.com/nexora/bff-customer/internal/domain"
)

type IdentityClient interface {
	StartOTP(ctx context.Context, tenantID, phone string) (challengeID string, err error)
	VerifyOTP(ctx context.Context, tenantID, challengeID, code string) (domain.SessionView, error)
}

type CatalogClient interface {
	Search(ctx context.Context, tenantID, query string) ([]map[string]any, error)
	Categories(ctx context.Context, tenantID string) ([]map[string]any, error)
	Product(ctx context.Context, tenantID, productID string) (map[string]any, error)
}

type StoreClient interface {
	ListStores(ctx context.Context, tenantID string) ([]map[string]any, error)
}

type RecClient interface {
	ForYou(ctx context.Context, tenantID, customerID string) ([]map[string]any, error)
}

type CartClient interface {
	Get(ctx context.Context, tenantID, cartID string) (map[string]any, error)
	AddItem(ctx context.Context, tenantID, cartID, sku string, qty int64, unitMinor int64) (map[string]any, error)
}

type CheckoutClient interface {
	Preview(ctx context.Context, tenantID, cartID string) (domain.CheckoutPreview, error)
	Place(ctx context.Context, tenantID, cartID, paymentMethod, sessionID string, addr domain.CheckoutAddress) (orderID string, err error)
}

type OrderClient interface {
	Get(ctx context.Context, tenantID, orderID string) (map[string]any, error)
	List(ctx context.Context, tenantID, principalID string) ([]map[string]any, error)
	Cancel(ctx context.Context, tenantID, orderID, reason string) (map[string]any, error)
}

type TrackingClient interface {
	Track(ctx context.Context, tenantID, orderID string) (domain.OrderTrack, error)
}

type LocationClient interface {
	Serviceable(ctx context.Context, tenantID string, lat, lng float64) (bool, error)
}

type NotificationClient interface {
	RegisterDevice(ctx context.Context, tenantID, customerID, token string) error
}

type CRMClient interface {
	OpenTicket(ctx context.Context, tenantID, customerID, subject string) (ticketID string, err error)
}

type ReviewClient interface {
	Submit(ctx context.Context, tenantID, orderID string, rating int, body string) error
}
