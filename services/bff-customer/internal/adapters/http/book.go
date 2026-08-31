package httpadapter

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// CustomerBook holds per-principal customer records for BFF-owned contracts
// (addresses, favorites, orders placed through this edge) when a dedicated
// downstream is not yet populated. Isolation is tenant+principal.
type CustomerBook struct {
	mu            sync.Mutex
	addresses     map[string][]map[string]any
	favorites     map[string][]map[string]any
	orders        map[string][]map[string]any
	profiles      map[string]map[string]any
	notifications map[string][]map[string]any
	tickets       map[string][]map[string]any
	notifyPrefs   map[string]map[string]any
}

func NewCustomerBook() *CustomerBook {
	return &CustomerBook{
		addresses:     make(map[string][]map[string]any),
		favorites:     make(map[string][]map[string]any),
		orders:        make(map[string][]map[string]any),
		profiles:      make(map[string]map[string]any),
		notifications: make(map[string][]map[string]any),
		tickets:       make(map[string][]map[string]any),
		notifyPrefs:   make(map[string]map[string]any),
	}
}

func bookKey(tenantID, principalID string) string {
	return tenantID + "|" + principalID
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func cloneSlice(in []map[string]any) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, item := range in {
		out = append(out, cloneMap(item))
	}
	return out
}

func (b *CustomerBook) ListAddresses(tenantID, principalID string) []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return cloneSlice(b.addresses[bookKey(tenantID, principalID)])
}

func (b *CustomerBook) PutAddress(tenantID, principalID string, addr map[string]any) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	id, _ := addr["id"].(string)
	if id == "" {
		id = uuid.NewString()
		addr["id"] = id
	}
	if addr["createdAt"] == nil {
		addr["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	}
	addr["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	list := b.addresses[key]
	replaced := false
	isDefault, _ := addr["is_default"].(bool)
	if v, ok := addr["isDefault"].(bool); ok {
		isDefault = v
		addr["is_default"] = v
	}
	for i, existing := range list {
		if existing["id"] == id {
			list[i] = addr
			replaced = true
			break
		}
	}
	if !replaced {
		list = append(list, addr)
	}
	if isDefault {
		for i := range list {
			if list[i]["id"] != id {
				list[i]["is_default"] = false
				list[i]["isDefault"] = false
			}
		}
	}
	b.addresses[key] = list
	return cloneMap(addr)
}

func (b *CustomerBook) DeleteAddress(tenantID, principalID, id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	list := b.addresses[key]
	out := list[:0]
	found := false
	for _, item := range list {
		if item["id"] == id {
			found = true
			continue
		}
		out = append(out, item)
	}
	if found {
		b.addresses[key] = out
	}
	return found
}

func (b *CustomerBook) SetDefaultAddress(tenantID, principalID, id string) (map[string]any, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	list := b.addresses[key]
	var found map[string]any
	for i := range list {
		ok := list[i]["id"] == id
		list[i]["is_default"] = ok
		list[i]["isDefault"] = ok
		if ok {
			found = list[i]
		}
	}
	if found == nil {
		return nil, false
	}
	return cloneMap(found), true
}

func (b *CustomerBook) ListFavorites(tenantID, principalID string) []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return cloneSlice(b.favorites[bookKey(tenantID, principalID)])
}

func (b *CustomerBook) AddFavorite(tenantID, principalID string, fav map[string]any) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	id, _ := fav["id"].(string)
	if id == "" {
		id = uuid.NewString()
		fav["id"] = id
	}
	fav["added_at"] = time.Now().UTC().Format(time.RFC3339)
	b.favorites[key] = append(b.favorites[key], fav)
	return cloneMap(fav)
}

func (b *CustomerBook) DeleteFavorite(tenantID, principalID, id string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	list := b.favorites[key]
	out := list[:0]
	found := false
	for _, item := range list {
		if item["id"] == id {
			found = true
			continue
		}
		out = append(out, item)
	}
	if found {
		b.favorites[key] = out
	}
	return found
}

func (b *CustomerBook) RememberOrder(tenantID, principalID string, order map[string]any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	b.orders[key] = append([]map[string]any{cloneMap(order)}, b.orders[key]...)
}

func (b *CustomerBook) ListOrders(tenantID, principalID string) []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return cloneSlice(b.orders[bookKey(tenantID, principalID)])
}

func (b *CustomerBook) UpdateOrder(tenantID, principalID, id string, patch map[string]any) (map[string]any, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	list := b.orders[key]
	for i, o := range list {
		if asString(o["id"]) == id || asString(o["orderId"]) == id {
			for k, v := range patch {
				o[k] = v
			}
			list[i] = o
			return cloneMap(o), true
		}
	}
	return nil, false
}

func (b *CustomerBook) Profile(tenantID, principalID string) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	if existing := b.profiles[bookKey(tenantID, principalID)]; existing != nil {
		return cloneMap(existing)
	}
	return map[string]any{
		"id": principalID, "first_name": "", "last_name": "", "phone": "",
		"display_name": "", "locale": "tr",
	}
}

func (b *CustomerBook) PutProfile(tenantID, principalID string, patch map[string]any) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	cur := b.profiles[key]
	if cur == nil {
		cur = map[string]any{"id": principalID, "locale": "tr"}
	}
	for k, v := range patch {
		if v == nil || v == "" {
			continue
		}
		cur[k] = v
	}
	cur["id"] = principalID
	cur["updatedAt"] = time.Now().UTC().Format(time.RFC3339)
	b.profiles[key] = cur
	return cloneMap(cur)
}

func (b *CustomerBook) ListNotifications(tenantID, principalID string) []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return cloneSlice(b.notifications[bookKey(tenantID, principalID)])
}

func (b *CustomerBook) AddNotification(tenantID, principalID string, n map[string]any) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	if asStringBook(n["id"]) == "" {
		n["id"] = uuid.NewString()
	}
	n["read"] = false
	n["created_at"] = time.Now().UTC().Format(time.RFC3339)
	b.notifications[key] = append([]map[string]any{cloneMap(n)}, b.notifications[key]...)
	return cloneMap(n)
}

func (b *CustomerBook) MarkNotificationRead(tenantID, principalID, id string) (map[string]any, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	list := b.notifications[bookKey(tenantID, principalID)]
	for i := range list {
		if list[i]["id"] == id {
			list[i]["read"] = true
			return cloneMap(list[i]), true
		}
	}
	return nil, false
}

func (b *CustomerBook) MarkAllNotificationsRead(tenantID, principalID string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for i := range b.notifications[bookKey(tenantID, principalID)] {
		b.notifications[bookKey(tenantID, principalID)][i]["read"] = true
	}
}

func (b *CustomerBook) NotificationPrefs(tenantID, principalID string) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	if p := b.notifyPrefs[bookKey(tenantID, principalID)]; p != nil {
		return cloneMap(p)
	}
	return map[string]any{
		"transactional": true, "promo": true, "delivery": true,
		"push_enabled": true, "email_enabled": false,
	}
}

func (b *CustomerBook) PutNotificationPrefs(tenantID, principalID string, prefs map[string]any) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.notifyPrefs[bookKey(tenantID, principalID)] = cloneMap(prefs)
	return cloneMap(prefs)
}

func (b *CustomerBook) ListTickets(tenantID, principalID string) []map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	return cloneSlice(b.tickets[bookKey(tenantID, principalID)])
}

func (b *CustomerBook) AddTicket(tenantID, principalID string, t map[string]any) map[string]any {
	b.mu.Lock()
	defer b.mu.Unlock()
	key := bookKey(tenantID, principalID)
	if asStringBook(t["id"]) == "" {
		t["id"] = uuid.NewString()
	}
	t["ticketId"] = t["id"]
	if asStringBook(t["status"]) == "" {
		t["status"] = "open"
	}
	t["created_at"] = time.Now().UTC().Format(time.RFC3339)
	b.tickets[key] = append([]map[string]any{cloneMap(t)}, b.tickets[key]...)
	return cloneMap(t)
}

func (b *CustomerBook) GetTicket(tenantID, principalID, id string) (map[string]any, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, item := range b.tickets[bookKey(tenantID, principalID)] {
		if item["id"] == id || item["ticketId"] == id {
			return cloneMap(item), true
		}
	}
	return nil, false
}

func asStringBook(v any) string {
	s, _ := v.(string)
	return s
}
