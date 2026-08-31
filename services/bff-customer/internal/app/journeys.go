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
	Stores   ports.StoreClient
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
	if d.Catalog != nil {
		items, err := d.Catalog.Search(ctx, tenantID, query)
		if err != nil {
			return feed, err
		}
		feed.Products = items
		if cats, err := d.Catalog.Categories(ctx, tenantID); err == nil {
			feed.Categories = cats
		}
	}
	if d.Stores != nil {
		if stores, err := d.Stores.ListStores(ctx, tenantID); err == nil {
			feed.Stores = stores
		}
	}
	if d.Recs != nil && customerID != "" {
		// Recommendation rails are supplementary: a recs outage must not take down
		// the whole home feed, which would leave the storefront unusable.
		if rails, err := d.Recs.ForYou(ctx, tenantID, customerID); err == nil {
			feed.Rails = rails
		}
	}
	feed.Widgets = composeHomeWidgets(feed)
	return feed, nil
}

func composeHomeWidgets(feed domain.HomeFeed) []map[string]any {
	widgets := make([]map[string]any, 0, 6)
	if feed.Serviceable {
		widgets = append(widgets, map[string]any{
			"id": "campaign-welcome", "type": "campaign", "title": "WELCOME10",
			"payload": map[string]any{"subtitle": "10% off when the basket is over 150 TL", "deep_link": "/coupons"},
		})
	}
	if len(feed.Stores) > 0 {
		items := make([]map[string]any, 0, len(feed.Stores))
		for _, store := range feed.Stores {
			id := firstString(store, "id", "code", "ID")
			name := firstString(store, "name", "title", "Name")
			if id == "" {
				continue
			}
			items = append(items, map[string]any{
				"id": id, "title": name, "deep_link": "/stores/" + id,
			})
		}
		if len(items) > 0 {
			widgets = append(widgets, map[string]any{
				"id": "nearby-stores", "type": "brands", "title": "Nearby stores", "items": items,
			})
		}
	}
	if len(feed.Categories) > 0 {
		items := make([]map[string]any, 0, len(feed.Categories))
		for _, cat := range feed.Categories {
			id := firstString(cat, "id", "slug", "categoryId")
			title := firstString(cat, "title", "name", "Name")
			if id == "" {
				continue
			}
			items = append(items, map[string]any{
				"id": id, "title": title, "deep_link": "/categories/" + id,
			})
		}
		if len(items) > 0 {
			widgets = append(widgets, map[string]any{
				"id": "categories", "type": "favorite_categories", "title": "Categories", "items": items,
			})
		}
	}
	if products := productTiles(feed.Products); len(products) > 0 {
		widgets = append(widgets, map[string]any{
			"id": "popular", "type": "trending", "title": "Popular", "items": products,
		})
	}
	if recs := productTiles(feed.Rails); len(recs) > 0 {
		widgets = append(widgets, map[string]any{
			"id": "recommended", "type": "recommendation", "title": "Recommended", "items": recs,
		})
	}
	return widgets
}

func productTiles(in []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, item := range in {
		id := firstString(item, "productId", "id", "sku", "ProductID")
		title := firstString(item, "name", "title", "Title", "sku")
		if id == "" {
			continue
		}
		if title == "" {
			title = id
		}
		tile := map[string]any{
			"id": id, "title": title, "deep_link": "/p/" + id,
			"price_minor": firstInt(item, "priceMinor", "price_minor", "unitPriceMinor"),
			"currency":    firstString(item, "currency", "Currency"),
		}
		if tile["currency"] == "" {
			tile["currency"] = "TRY"
		}
		out = append(out, tile)
	}
	return out
}

func firstString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if s, ok := m[k].(string); ok && s != "" {
			return s
		}
	}
	return ""
}

func firstInt(m map[string]any, keys ...string) int64 {
	for _, k := range keys {
		switch v := m[k].(type) {
		case int:
			return int64(v)
		case int64:
			return v
		case float64:
			return int64(v)
		}
	}
	return 0
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

func (d *Deps) PlaceOrder(ctx context.Context, tenantID, cartID, paymentMethod, sessionID string, addr domain.CheckoutAddress) (string, error) {
	if d.Checkout == nil {
		return "", domain.ErrUpstream
	}
	return d.Checkout.Place(ctx, tenantID, cartID, paymentMethod, sessionID, addr)
}

func (d *Deps) TrackOrder(ctx context.Context, tenantID, orderID string) (domain.OrderTrack, error) {
	if d.Tracking != nil {
		tr, err := d.Tracking.Track(ctx, tenantID, orderID)
		if err == nil && tr.Status != "" && tr.Status != "unknown" && tr.Status != "Custom" {
			return tr, nil
		}
	}
	if d.Orders == nil {
		return domain.OrderTrack{}, domain.ErrUpstream
	}
	raw, err := d.Orders.Get(ctx, tenantID, orderID)
	if err != nil {
		return domain.OrderTrack{}, err
	}
	status, _ := raw["status"].(string)
	if status == "" {
		status = "unknown"
	}
	return domain.OrderTrack{OrderID: orderID, Status: status}, nil
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
