package httpadapter

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nexora/bff-customer/internal/domain"
)

func (h *Handler) search(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Catalog == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	q := r.URL.Query().Get("q")
	items, err := h.Deps.Catalog.Search(r.Context(), tenant(r), q)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{
		"query": q, "items": items, "hits": items, "total_count": len(items), "totalCount": len(items),
	})
}

func (h *Handler) searchSuggestions(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Catalog == nil {
		writeJSON(w, 200, map[string]any{"suggestions": []any{}})
		return
	}
	q := r.URL.Query().Get("q")
	items, err := h.Deps.Catalog.Search(r.Context(), tenant(r), q)
	if err != nil {
		writeErr(w, err)
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		title := firstNonEmpty(asString(item["name"]), asString(item["title"]), asString(item["sku"]))
		if title == "" {
			continue
		}
		out = append(out, map[string]any{"query": title, "title": title, "text": title})
	}
	writeJSON(w, 200, map[string]any{"suggestions": out})
}

func (h *Handler) getProduct(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Catalog == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	id := r.PathValue("id")
	item, err := h.Deps.Catalog.Product(r.Context(), tenant(r), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, item)
}

func (h *Handler) listCategories(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Catalog == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	items, err := h.Deps.Catalog.Categories(r.Context(), tenant(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": items, "categories": items})
}

func (h *Handler) getCategory(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Catalog == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	id := r.PathValue("id")
	items, err := h.Deps.Catalog.Categories(r.Context(), tenant(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	for _, item := range items {
		if asString(item["id"]) == id || asString(item["slug"]) == id {
			writeJSON(w, 200, item)
			return
		}
	}
	writeErr(w, domain.ErrNotFound)
}

func (h *Handler) listStores(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Stores == nil {
		writeJSON(w, 200, map[string]any{"items": []any{}, "stores": []any{}})
		return
	}
	items, err := h.Deps.Stores.ListStores(r.Context(), tenant(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"items": items, "stores": items})
}

func (h *Handler) getStore(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Stores == nil {
		writeErr(w, domain.ErrNotFound)
		return
	}
	id := r.PathValue("id")
	items, err := h.Deps.Stores.ListStores(r.Context(), tenant(r))
	if err != nil {
		writeErr(w, err)
		return
	}
	for _, item := range items {
		if asString(item["id"]) == id || asString(item["code"]) == id {
			writeJSON(w, 200, item)
			return
		}
	}
	writeErr(w, domain.ErrNotFound)
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	seen := map[string]bool{}
	owned := make([]map[string]any, 0)
	if h.Deps.Orders != nil {
		if items, err := h.Deps.Orders.List(r.Context(), tenant(r), pid); err == nil {
			for _, item := range items {
				if ownedByCaller(r, item) || asString(item["customerPrincipalId"]) == pid {
					owned = append(owned, item)
					seen[firstNonEmpty(asString(item["id"]), asString(item["orderId"]))] = true
				}
			}
		}
	}
	if h.Book != nil {
		for _, item := range h.Book.ListOrders(tenant(r), pid) {
			id := firstNonEmpty(asString(item["id"]), asString(item["orderId"]))
			if id == "" || seen[id] {
				continue
			}
			owned = append(owned, item)
			seen[id] = true
		}
	}
	writeJSON(w, 200, owned)
}

func (h *Handler) cancelOrder(w http.ResponseWriter, r *http.Request) {
	if err := h.requireOwnedOrder(r); err != nil {
		writeErr(w, err)
		return
	}
	if h.Deps.Orders == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	out, err := h.Deps.Orders.Cancel(r.Context(), tenant(r), r.PathValue("id"), body.Reason)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) reorder(w http.ResponseWriter, r *http.Request) {
	if err := h.requireOwnedOrder(r); err != nil {
		writeErr(w, err)
		return
	}
	id := r.PathValue("id")
	items := []map[string]any{}
	if order, err := h.lookupOrder(r, id); err == nil {
		items = extractOrderLines(order)
	}
	writeJSON(w, 200, map[string]any{
		"status": "cart_seeded", "id": id, "orderId": id, "items": items,
	})
}

func (h *Handler) lookupOrder(r *http.Request, id string) (map[string]any, error) {
	if h.Deps.Orders != nil {
		if out, err := h.Deps.Orders.Get(r.Context(), tenant(r), id); err == nil {
			return out, nil
		}
	}
	if h.Book != nil {
		pid := callerID(r)
		for _, o := range h.Book.ListOrders(tenant(r), pid) {
			if asString(o["id"]) == id || asString(o["orderId"]) == id {
				return o, nil
			}
		}
	}
	return nil, domain.ErrNotFound
}

func extractOrderLines(order map[string]any) []map[string]any {
	raw, _ := order["items"].([]any)
	if raw == nil {
		if typed, ok := order["items"].([]map[string]any); ok {
			return typed
		}
		raw, _ = order["lines"].([]any)
	}
	out := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		out = append(out, map[string]any{
			"id":               firstNonEmpty(asString(m["id"]), asString(m["sku"])),
			"product_id":       firstNonEmpty(asString(m["product_id"]), asString(m["productId"]), asString(m["sku"])),
			"productId":        firstNonEmpty(asString(m["productId"]), asString(m["product_id"]), asString(m["sku"])),
			"name":             firstNonEmpty(asString(m["name"]), asString(m["title"])),
			"title":            firstNonEmpty(asString(m["title"]), asString(m["name"])),
			"quantity":         m["quantity"],
			"unit_price_minor": firstNonEmptyNum(m, "unit_price_minor", "unitPriceMinor", "priceMinor"),
			"sku":              asString(m["sku"]),
		})
	}
	return out
}

func firstNonEmptyNum(m map[string]any, keys ...string) any {
	for _, k := range keys {
		if m[k] != nil {
			return m[k]
		}
	}
	return 0
}

func (h *Handler) listAddresses(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if h.Book == nil {
		writeJSON(w, 200, []any{})
		return
	}
	writeJSON(w, 200, h.Book.ListAddresses(tenant(r), pid))
}

func (h *Handler) getAddress(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	for _, item := range h.Book.ListAddresses(tenant(r), pid) {
		if asString(item["id"]) == id {
			writeJSON(w, 200, item)
			return
		}
	}
	writeErr(w, domain.ErrNotFound)
}

func addressFromBody(body map[string]any) map[string]any {
	line1 := firstNonEmpty(asString(body["formatted"]), asString(body["line1"]), asString(body["address_line"]))
	lat, _ := toFloat(body["lat"])
	lng, _ := toFloat(body["lng"])
	isDefault, _ := body["is_default"].(bool)
	if v, ok := body["isDefault"].(bool); ok {
		isDefault = v
	}
	isFav, _ := body["is_favorite"].(bool)
	if v, ok := body["isFavorite"].(bool); ok {
		isFav = v
	}
	return map[string]any{
		"id":                     asString(body["id"]),
		"title":                  firstNonEmpty(asString(body["title"]), asString(body["label"]), asString(body["custom_label"])),
		"label":                  asString(body["label"]),
		"custom_label":           firstNonEmpty(asString(body["custom_label"]), asString(body["customLabel"])),
		"formatted":              line1,
		"line1":                  line1,
		"building":               asString(body["building"]),
		"floor":                  asString(body["floor"]),
		"door":                   firstNonEmpty(asString(body["door"]), asString(body["apartment"])),
		"delivery_instructions":  firstNonEmpty(asString(body["delivery_instructions"]), asString(body["notes"])),
		"recipient_name":         firstNonEmpty(asString(body["recipient_name"]), asString(body["recipientName"])),
		"phone":                  firstNonEmpty(asString(body["phone"]), asString(body["recipient_phone"])),
		"lat":                    lat,
		"lng":                    lng,
		"city":                   asString(body["city"]),
		"country":                asString(body["country"]),
		"is_default":             isDefault,
		"isDefault":              isDefault,
		"is_favorite":            isFav,
		"serviceable":            body["serviceable"] != false,
	}
}

func toFloat(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(n, 64)
		return f, err == nil
	default:
		return 0, false
	}
}

func (h *Handler) createAddress(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	addr := addressFromBody(body)
	if asString(addr["line1"]) == "" && addr["lat"] == 0.0 && addr["lng"] == 0.0 {
		writeErr(w, domain.ErrInvalidArgument)
		return
	}
	writeJSON(w, 201, h.Book.PutAddress(tenant(r), pid, addr))
}

func (h *Handler) patchAddress(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	body["id"] = id
	existing := map[string]any{}
	for _, item := range h.Book.ListAddresses(tenant(r), pid) {
		if asString(item["id"]) == id {
			existing = item
			break
		}
	}
	if asString(existing["id"]) == "" {
		writeErr(w, domain.ErrNotFound)
		return
	}
	for k, v := range addressFromBody(body) {
		if v == nil || v == "" || v == 0.0 {
			continue
		}
		existing[k] = v
	}
	existing["id"] = id
	writeJSON(w, 200, h.Book.PutAddress(tenant(r), pid, existing))
}

func (h *Handler) deleteAddress(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if !h.Book.DeleteAddress(tenant(r), pid, r.PathValue("id")) {
		writeErr(w, domain.ErrNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) defaultAddress(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	addr, found := h.Book.SetDefaultAddress(tenant(r), pid, r.PathValue("id"))
	if !found {
		writeErr(w, domain.ErrNotFound)
		return
	}
	writeJSON(w, 200, addr)
}

func (h *Handler) favoriteAddress(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	id := r.PathValue("id")
	var body struct {
		Favorite bool `json:"favorite"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	for _, item := range h.Book.ListAddresses(tenant(r), pid) {
		if asString(item["id"]) == id {
			item["is_favorite"] = body.Favorite
			writeJSON(w, 200, h.Book.PutAddress(tenant(r), pid, item))
			return
		}
	}
	writeErr(w, domain.ErrNotFound)
}

func (h *Handler) validateZone(w http.ResponseWriter, r *http.Request) {
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	ok := true
	if h.Deps.Location != nil {
		var err error
		ok, err = h.Deps.Location.Serviceable(r.Context(), tenant(r), lat, lng)
		if err != nil {
			writeErr(w, err)
			return
		}
	} else if lat == 0 && lng == 0 {
		ok = false
	}
	writeJSON(w, 200, map[string]any{"serviceable": ok, "is_serviceable": ok})
}

func (h *Handler) listFavorites(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	writeJSON(w, 200, h.Book.ListFavorites(tenant(r), pid))
}

func (h *Handler) addFavorite(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	writeJSON(w, 201, h.Book.AddFavorite(tenant(r), pid, body))
}

func (h *Handler) deleteFavorite(w http.ResponseWriter, r *http.Request) {
	pid, ok := requirePrincipal(w, r)
	if !ok {
		return
	}
	if !h.Book.DeleteFavorite(tenant(r), pid, r.PathValue("id")) {
		writeErr(w, domain.ErrNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
