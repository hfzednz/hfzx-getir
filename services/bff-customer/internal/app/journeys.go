package app

import (
	"context"

	"github.com/nexora/bff-customer/internal/app/ports"
	"github.com/nexora/bff-customer/internal/domain"
)

type Deps struct {
	Identity ports.IdentityClient
	Catalog  ports.CatalogClient
	Recs     ports.RecClient
	Cart     ports.CartClient
	Checkout ports.CheckoutClient
	Orders   ports.OrderClient
	Tracking ports.TrackingClient
	Location ports.LocationClient
	Notify   ports.NotificationClient
	CRM      ports.CRMClient
	Reviews  ports.ReviewClient
}

func (d *Deps) StartOTP(ctx context.Context, tenantID, phone string) (string, error) {
	if tenantID == "" || phone == "" {
		return "", domain.ErrInvalidArgument
	}
	if d.Identity == nil {
		return "", domain.ErrUpstream
	}
	return d.Identity.StartOTP(ctx, tenantID, phone)
}

func (d *Deps) Login(ctx context.Context, tenantID, challengeID, code string) (domain.SessionView, error) {
	if tenantID == "" || challengeID == "" || code == "" {
		return domain.SessionView{}, domain.ErrInvalidArgument
	}
	if d.Identity == nil {
		return domain.SessionView{}, domain.ErrUpstream
	}
	return d.Identity.VerifyOTP(ctx, tenantID, challengeID, code)
}

func (d *Deps) Home(ctx context.Context, tenantID, customerID, query string, lat, lng float64) (domain.HomeFeed, error) {
	feed := domain.HomeFeed{Query: query, Serviceable: true}
	if d.Location != nil {
		ok, err := d.Location.Serviceable(ctx, tenantID, lat, lng)
		if err != nil {
			return feed, err
		}
		feed.Serviceable = ok
	}
	if d.Catalog != nil && query != "" {
		items, err := d.Catalog.Search(ctx, tenantID, query)
		if err != nil {
			return feed, err
		}
		feed.Products = items
	}
	if d.Recs != nil && customerID != "" {
		rails, err := d.Recs.ForYou(ctx, tenantID, customerID)
		if err != nil {
			return feed, err
		}
		feed.Rails = rails
	}
	return feed, nil
}

func (d *Deps) AddToCart(ctx context.Context, tenantID, cartID, sku string, qty, unitMinor int64) (map[string]any, error) {
	if d.Cart == nil {
		return nil, domain.ErrUpstream
	}
	return d.Cart.AddItem(ctx, tenantID, cartID, sku, qty, unitMinor)
}

func (d *Deps) PreviewCheckout(ctx context.Context, tenantID, cartID string) (domain.CheckoutPreview, error) {
	if d.Checkout == nil {
		return domain.CheckoutPreview{}, domain.ErrUpstream
	}
	return d.Checkout.Preview(ctx, tenantID, cartID)
}

func (d *Deps) PlaceOrder(ctx context.Context, tenantID, cartID, paymentMethod, sessionID string) (string, error) {
	if d.Checkout == nil {
		return "", domain.ErrUpstream
	}
	return d.Checkout.Place(ctx, tenantID, cartID, paymentMethod, sessionID)
}

func (d *Deps) TrackOrder(ctx context.Context, tenantID, orderID string) (domain.OrderTrack, error) {
	if d.Tracking == nil {
		return domain.OrderTrack{}, domain.ErrUpstream
	}
	return d.Tracking.Track(ctx, tenantID, orderID)
}

func (d *Deps) OpenSupport(ctx context.Context, tenantID, customerID, subject string) (string, error) {
	if d.CRM == nil {
		return "", domain.ErrUpstream
	}
	return d.CRM.OpenTicket(ctx, tenantID, customerID, subject)
}

func (d *Deps) SubmitReview(ctx context.Context, tenantID, orderID string, rating int, body string) error {
	if rating < 1 || rating > 5 {
		return domain.ErrInvalidArgument
	}
	if d.Reviews == nil {
		return domain.ErrUpstream
	}
	return d.Reviews.Submit(ctx, tenantID, orderID, rating, body)
}
