package httpadapter

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/nexora/bff-customer/internal/domain"
	"github.com/nexora/bff-customer/internal/reqctx"
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
	storeID := firstNonEmpty(r.URL.Query().Get("storeId"), r.URL.Query().Get("store_id"))
	if storeID != "" && h.Deps.Stores != nil {
		stock, err := h.Deps.Stores.StoreStock(r.Context(), tenant(r), storeID)
		if err != nil {
			writeErr(w, err)
			return
		}
		items = filterItemsByStock(items, stock)
	}
	writeJSON(w, 200, map[string]any{
		"query": q, "items": items, "hits": items, "total_count": len(items), "totalCount": len(items),
		"storeId": storeID,
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

func (h *Handler) listStoreProducts(w http.ResponseWriter, r *http.Request) {
	if h.Deps.Stores == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	id := r.PathValue("id")
	var store map[string]any
	if items, err := h.Deps.Stores.ListStores(r.Context(), tenant(r)); err == nil {
		for _, item := range items {
			if asString(item["id"]) == id || asString(item["code"]) == id {
				store = item
				break
			}
		}
	}
	if store == nil {
		writeErr(w, domain.ErrNotFound)
		return
	}
	stock, err := h.Deps.Stores.StoreStock(r.Context(), tenant(r), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	var catalog []map[string]any
	if h.Deps.Catalog != nil {
		catalog, _ = h.Deps.Catalog.Search(r.Context(), tenant(r), "")
	}
	products := joinStockWithCatalog(stock, catalog)
	writeJSON(w, 200, map[string]any{
		"id": id, "store": store, "items": products, "products": products,
		"open":       store["open"] == true || asString(store["status"]) == "open" || asString(store["status"]) == "active",
		"etaMinutes": store["etaMinutes"], "minOrderMinor": store["minOrderMinor"],
		"deliveryFeeMinor": store["deliveryFeeMinor"],
	})
}

func filterItemsByStock(items, stock []map[string]any) []map[string]any {
	avail := map[string]int64{}
	for _, row := range stock {
		sku := strings.ToLower(firstNonEmpty(asString(row["sku"]), asString(row["skuCode"])))
		if sku == "" {
			continue
		}
		qty := asIntFromAny(row["available"])
		oos, _ := row["outOfStock"].(bool)
		if oos || qty <= 0 {
			continue
		}
		avail[sku] = qty
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		sku := strings.ToLower(firstNonEmpty(asString(item["sku"]), asString(item["id"]), asString(item["productId"])))
		if qty, ok := avail[sku]; ok {
			item["available"] = qty
			item["outOfStock"] = false
			item["stock_status"] = "in_stock"
			out = append(out, item)
		}
	}
	return out
}

func joinStockWithCatalog(stock, catalog []map[string]any) []map[string]any {
	bySKU := map[string]map[string]any{}
	for _, p := range catalog {
		key := strings.ToLower(firstNonEmpty(asString(p["sku"]), asString(p["id"]), asString(p["productId"])))
		if key != "" {
			bySKU[key] = p
		}
	}
	out := make([]map[string]any, 0, len(stock))
	for _, row := range stock {
		sku := firstNonEmpty(asString(row["sku"]), asString(row["skuCode"]))
		item := map[string]any{}
		if p, ok := bySKU[strings.ToLower(sku)]; ok {
			for k, v := range p {
				item[k] = v
			}
		}
		item["id"] = firstNonEmpty(asString(item["id"]), sku)
		item["sku"] = sku
		name := firstNonEmpty(asString(item["name"]), asString(item["title"]), asString(row["name"]), asString(row["title"]), sku)
		item["name"] = name
		item["title"] = name
		avail := asIntFromAny(row["available"])
		oos, _ := row["outOfStock"].(bool)
		if avail <= 0 {
			oos = true
		}
		item["available"] = avail
		item["outOfStock"] = oos
		if oos {
			item["stock_status"] = "out_of_stock"
		} else {
			item["stock_status"] = "in_stock"
		}
		if pm := asIntFromAny(row["priceMinor"]); pm > 0 {
			item["priceMinor"] = pm
			item["price_minor"] = pm
		}
		out = append(out, item)
	}
	return out
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
	writeJSON(w, 200, map[string]any{"items": owned})
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
	if h.Deps == nil || h.Deps.Cart == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	id := r.PathValue("id")
	order, err := h.lookupOrder(r, id)
	if err != nil {
		writeErr(w, err)
		return
	}
	items := extractOrderLines(order)
	if len(items) == 0 {
		writeJSON(w, 409, map[string]any{
			"error": map[string]any{
				"code": "conflict", "message": "No items available to reorder",
			},
		})
		return
	}
	if pid := callerID(r); pid != "" {
		r = r.WithContext(reqctx.WithUserID(r.Context(), pid))
	}
	cartID := ""
	added := 0
	for _, line := range items {
		sku := firstNonEmpty(asString(line["sku"]), asString(line["productId"]), asString(line["product_id"]))
		if sku == "" {
			continue
		}
		qty := asInt64(line["quantity"])
		if qty == 0 {
			qty = asInt64(line["qty"])
		}
		if qty < 1 {
			qty = 1
		}
		unit := asInt64(line["unit_price_minor"])
		if unit == 0 {
			unit = asInt64(line["unitPriceMinor"])
		}
		out, err := h.Deps.Cart.AddItem(r.Context(), tenant(r), cartID, sku, qty, unit)
		if err != nil {
			writeErr(w, err)
			return
		}
		cartID = firstNonEmpty(asString(out["cartId"]), asString(out["id"]), asString(out["ID"]), cartID)
		added++
	}
	if cartID == "" || added == 0 {
		writeJSON(w, 409, map[string]any{
			"error": map[string]any{
				"code": "conflict", "message": "Could not add items to a new cart",
			},
		})
		return
	}
	writeJSON(w, 200, map[string]any{
		"status":  "cart_created",
		"cartId":  cartID,
		"id":      cartID,
		"orderId": id,
		"items":   items,
	})
}

func (h *Handler) refundOrder(w http.ResponseWriter, r *http.Request) {
	if err := h.requireOwnedOrder(r); err != nil {
		writeErr(w, err)
		return
	}
	if h.Deps == nil || h.Deps.Orders == nil {
		writeErr(w, domain.ErrUpstream)
		return
	}
	var body struct {
		Reason      string `json:"reason"`
		AmountMinor int64  `json:"amountMinor"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	out, err := h.Deps.Orders.Refund(r.Context(), tenant(r), r.PathValue("id"), body.Reason, body.AmountMinor)
	if err != nil {
		writeErr(w, err)
		return
	}
	writeJSON(w, 200, out)
}

func (h *Handler) favoriteOrder(w http.ResponseWriter, r *http.Request) {
	if err := h.requireOwnedOrder(r); err != nil {
		writeErr(w, err)
		return
	}
	var body struct {
		Favorite bool `json:"favorite"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	id := r.PathValue("id")
	if h.Book != nil {
		if out, ok := h.Book.UpdateOrder(tenant(r), callerID(r), id, map[string]any{
			"isFavorite": body.Favorite, "is_favorite": body.Favorite,
		}); ok {
			writeJSON(w, 200, out)
			return
		}
	}
	order, err := h.lookupOrder(r, id)
	if err != nil {
		writeErr(w, err)
		return
	}
	order["isFavorite"] = body.Favorite
	order["is_favorite"] = body.Favorite
	writeJSON(w, 200, order)
}

func (h *Handler) orderDocument(kind string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h.requireOwnedOrder(r); err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, 404, map[string]any{
			"error": map[string]any{
				"code":    "not_found",
				"message": kind + " is not available for this order yet",
			},
		})
	}
}

func (h *Handler) lookupOrder(r *http.Request, id string) (map[string]any, error) {
	if h.Book != nil {
		pid := callerID(r)
		for _, o := range h.Book.ListOrders(tenant(r), pid) {
			if asString(o["id"]) == id || asString(o["orderId"]) == id {
				return o, nil
			}
		}
	}
	if h.Deps.Orders != nil {
		if out, err := h.Deps.Orders.Get(r.Context(), tenant(r), id); err == nil && ownedByCaller(r, out) {
			return out, nil
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
		sku := firstNonEmpty(asString(m["sku"]), asString(m["skuCode"]), asString(m["variantId"]), asString(m["productId"]), asString(m["product_id"]))
		qty := firstNonEmptyNum(m, "quantity", "qty")
		out = append(out, map[string]any{
			"id":               firstNonEmpty(asString(m["id"]), sku),
			"product_id":       firstNonEmpty(asString(m["product_id"]), asString(m["productId"]), sku),
			"productId":        firstNonEmpty(asString(m["productId"]), asString(m["product_id"]), sku),
			"name":             firstNonEmpty(asString(m["name"]), asString(m["title"]), asString(m["titleSnapshot"])),
			"title":            firstNonEmpty(asString(m["title"]), asString(m["name"]), asString(m["titleSnapshot"])),
			"quantity":         qty,
			"qty":              qty,
			"unit_price_minor": firstNonEmptyNum(m, "unit_price_minor", "unitPriceMinor", "priceMinor", "unit_price_minor"),
			"sku":              sku,
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

func asInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case float64:
		return int64(n)
	case json.Number:
		i, _ := n.Int64()
		return i
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	default:
		return 0
	}
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
		"id":                    asString(body["id"]),
		"title":                 firstNonEmpty(asString(body["title"]), asString(body["label"]), asString(body["custom_label"])),
		"label":                 asString(body["label"]),
		"custom_label":          firstNonEmpty(asString(body["custom_label"]), asString(body["customLabel"])),
		"formatted":             line1,
		"line1":                 line1,
		"building":              asString(body["building"]),
		"floor":                 asString(body["floor"]),
		"door":                  firstNonEmpty(asString(body["door"]), asString(body["apartment"])),
		"delivery_instructions": firstNonEmpty(asString(body["delivery_instructions"]), asString(body["notes"])),
		"recipient_name":        firstNonEmpty(asString(body["recipient_name"]), asString(body["recipientName"])),
		"phone":                 firstNonEmpty(asString(body["phone"]), asString(body["recipient_phone"])),
		"lat":                   lat,
		"lng":                   lng,
		"city":                  asString(body["city"]),
		"country":               asString(body["country"]),
		"is_default":            isDefault,
		"isDefault":             isDefault,
		"is_favorite":           isFav,
		"serviceable":           body["serviceable"] != false,
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
