package memory

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/inventory-service/internal/app/ports"
	"github.com/nexora/inventory-service/internal/domain"
)

// WarehouseRepo implements ports.WarehouseRepository.
type WarehouseRepo struct{ S *Store }

func (r *WarehouseRepo) Create(_ context.Context, w domain.Warehouse) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := tenantKey(w.TenantID, w.Code)
	if _, ok := r.S.WHByCode[key]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Warehouses[w.ID] = w
	r.S.WHByCode[key] = w.ID
	return nil
}

func (r *WarehouseRepo) Update(_ context.Context, w domain.Warehouse) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Warehouses[w.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Warehouses[w.ID] = w
	return nil
}

func (r *WarehouseRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Warehouse, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	w, ok := r.S.Warehouses[id]
	if !ok || w.TenantID != tenantID || w.DeletedAt != nil {
		return domain.Warehouse{}, domain.ErrNotFound
	}
	return w, nil
}

func (r *WarehouseRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Warehouse, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.WHByCode[tenantKey(tenantID, code)]
	if !ok {
		return domain.Warehouse{}, domain.ErrNotFound
	}
	w := r.S.Warehouses[id]
	if w.DeletedAt != nil {
		return domain.Warehouse{}, domain.ErrNotFound
	}
	return w, nil
}

func (r *WarehouseRepo) List(_ context.Context, f ports.WarehouseFilter) ([]domain.Warehouse, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	q := strings.ToLower(strings.TrimSpace(f.Query))
	out := make([]domain.Warehouse, 0)
	for _, w := range r.S.Warehouses {
		if w.TenantID != f.TenantID || w.DeletedAt != nil {
			continue
		}
		if f.Status != nil && w.Status != *f.Status {
			continue
		}
		if q != "" && !strings.Contains(strings.ToLower(w.Code), q) && !strings.Contains(strings.ToLower(w.Name), q) {
			continue
		}
		out = append(out, w)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })
	total := len(out)
	if f.Offset >= len(out) {
		return nil, total, nil
	}
	end := f.Offset + f.Limit
	if f.Limit <= 0 || end > len(out) {
		end = len(out)
	}
	return out[f.Offset:end], total, nil
}

func (r *WarehouseRepo) Delete(_ context.Context, tenantID, id uuid.UUID, at time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	w, ok := r.S.Warehouses[id]
	if !ok || w.TenantID != tenantID {
		return domain.ErrNotFound
	}
	w.Status = domain.WarehouseStatusClosed
	w.DeletedAt = &at
	w.UpdatedAt = at
	r.S.Warehouses[id] = w
	return nil
}

// LocationRepo implements ports.LocationRepository.
type LocationRepo struct{ S *Store }

func (r *LocationRepo) Create(_ context.Context, l domain.Location) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Locations[l.ID] = l
	return nil
}

func (r *LocationRepo) Update(_ context.Context, l domain.Location) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Locations[l.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Locations[l.ID] = l
	return nil
}

func (r *LocationRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Location, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	l, ok := r.S.Locations[id]
	if !ok || l.DeletedAt != nil {
		return domain.Location{}, domain.ErrNotFound
	}
	return l, nil
}

func (r *LocationRepo) ListByWarehouse(_ context.Context, warehouseID uuid.UUID) ([]domain.Location, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Location, 0)
	for _, l := range r.S.Locations {
		if l.WarehouseID == warehouseID && l.DeletedAt == nil {
			out = append(out, l)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func (r *LocationRepo) ListChildren(_ context.Context, parentID uuid.UUID) ([]domain.Location, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Location, 0)
	for _, l := range r.S.Locations {
		if l.ParentID != nil && *l.ParentID == parentID && l.DeletedAt == nil {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *LocationRepo) Delete(_ context.Context, id uuid.UUID, at time.Time) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	l, ok := r.S.Locations[id]
	if !ok {
		return domain.ErrNotFound
	}
	l.DeletedAt = &at
	l.IsActive = false
	l.UpdatedAt = at
	r.S.Locations[id] = l
	return nil
}

// BalanceRepo implements ports.BalanceRepository.
type BalanceRepo struct{ S *Store }

func (r *BalanceRepo) Create(_ context.Context, b domain.StockBalance) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := balanceKeyStr(ports.BalanceKey{
		TenantID: b.TenantID, WarehouseID: b.WarehouseID,
		VariantID: b.VariantID, LocationID: b.LocationID,
	})
	if _, ok := r.S.BalanceKey[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Balances[b.ID] = b
	r.S.BalanceKey[k] = b.ID
	return nil
}

func (r *BalanceRepo) Update(_ context.Context, b domain.StockBalance) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cur, ok := r.S.Balances[b.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if b.Version < cur.Version {
		return domain.ErrVersionConflict
	}
	r.S.Balances[b.ID] = b
	return nil
}

func (r *BalanceRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.StockBalance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	b, ok := r.S.Balances[id]
	if !ok || b.TenantID != tenantID {
		return domain.StockBalance{}, domain.ErrNotFound
	}
	return b, nil
}

func (r *BalanceRepo) GetByKey(_ context.Context, key ports.BalanceKey) (domain.StockBalance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.BalanceKey[balanceKeyStr(key)]
	if !ok {
		return domain.StockBalance{}, domain.ErrNotFound
	}
	return r.S.Balances[id], nil
}

func (r *BalanceRepo) ListByWarehouse(_ context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.StockBalance, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.StockBalance, 0)
	for _, b := range r.S.Balances {
		if b.TenantID == tenantID && b.WarehouseID == warehouseID {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].VariantID.String() < out[j].VariantID.String() })
	total := len(out)
	if offset >= len(out) {
		return nil, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

func (r *BalanceRepo) ListByVariant(_ context.Context, tenantID, variantID uuid.UUID) ([]domain.StockBalance, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.StockBalance, 0)
	for _, b := range r.S.Balances {
		if b.TenantID == tenantID && b.VariantID == variantID {
			out = append(out, b)
		}
	}
	return out, nil
}

// LotRepo implements ports.LotRepository.
type LotRepo struct{ S *Store }

func (r *LotRepo) Create(_ context.Context, l domain.Lot) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Lots[l.ID] = l
	return nil
}

func (r *LotRepo) Update(_ context.Context, l domain.Lot) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Lots[l.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Lots[l.ID] = l
	return nil
}

func (r *LotRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Lot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	l, ok := r.S.Lots[id]
	if !ok || l.TenantID != tenantID {
		return domain.Lot{}, domain.ErrNotFound
	}
	return l, nil
}

func (r *LotRepo) ListByBalance(_ context.Context, balanceID uuid.UUID) ([]domain.Lot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Lot, 0)
	for _, l := range r.S.Lots {
		if l.BalanceID == balanceID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *LotRepo) ListByWarehouseVariant(_ context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.Lot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Lot, 0)
	for _, l := range r.S.Lots {
		if l.TenantID == tenantID && l.WarehouseID == warehouseID && l.VariantID == variantID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (r *LotRepo) ListNearExpiry(_ context.Context, tenantID uuid.UUID, warehouseID *uuid.UUID, withinDays int, asOf time.Time) ([]domain.Lot, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	deadline := asOf.AddDate(0, 0, withinDays)
	out := make([]domain.Lot, 0)
	for _, l := range r.S.Lots {
		if l.TenantID != tenantID || l.ExpiryDate == nil {
			continue
		}
		if warehouseID != nil && l.WarehouseID != *warehouseID {
			continue
		}
		if l.ExpiryDate.After(deadline) {
			continue
		}
		out = append(out, l)
	}
	domain.SortLotsFEFO(out)
	return out, nil
}

// ReservationRepo implements ports.ReservationRepository.
type ReservationRepo struct{ S *Store }

func (r *ReservationRepo) Create(_ context.Context, res domain.Reservation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Reservations[res.ID] = res
	return nil
}

func (r *ReservationRepo) Update(_ context.Context, res domain.Reservation) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Reservations[res.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Reservations[res.ID] = res
	return nil
}

func (r *ReservationRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Reservation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	res, ok := r.S.Reservations[id]
	if !ok || res.TenantID != tenantID {
		return domain.Reservation{}, domain.ErrNotFound
	}
	return res, nil
}

func (r *ReservationRepo) ListByExternalRef(_ context.Context, tenantID uuid.UUID, externalRef string) ([]domain.Reservation, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Reservation, 0)
	for _, res := range r.S.Reservations {
		if res.TenantID == tenantID && res.ExternalRef == externalRef {
			out = append(out, res)
		}
	}
	return out, nil
}

// MovementRepo implements ports.MovementRepository.
type MovementRepo struct{ S *Store }

func (r *MovementRepo) Create(_ context.Context, m domain.Movement) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.MovByIdem[m.IdempotencyKey]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Movements[m.ID] = m
	r.S.MovByIdem[m.IdempotencyKey] = m.ID
	return nil
}

func (r *MovementRepo) GetByIdempotencyKey(_ context.Context, key string) (domain.Movement, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.MovByIdem[key]
	if !ok {
		return domain.Movement{}, domain.ErrNotFound
	}
	return r.S.Movements[id], nil
}

func (r *MovementRepo) ListByBalance(_ context.Context, balanceID uuid.UUID, limit, offset int) ([]domain.Movement, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Movement, 0)
	for _, m := range r.S.Movements {
		if m.BalanceID != nil && *m.BalanceID == balanceID {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].OccurredAt.After(out[j].OccurredAt) })
	total := len(out)
	if offset >= len(out) {
		return nil, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

// TransferRepo implements ports.TransferRepository.
type TransferRepo struct{ S *Store }

func (r *TransferRepo) Create(_ context.Context, t domain.Transfer) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Transfers[t.ID] = t
	return nil
}

func (r *TransferRepo) Update(_ context.Context, t domain.Transfer) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Transfers[t.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Transfers[t.ID] = t
	return nil
}

func (r *TransferRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Transfer, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.Transfers[id]
	if !ok || t.TenantID != tenantID {
		return domain.Transfer{}, domain.ErrNotFound
	}
	return t, nil
}

func (r *TransferRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Transfer, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.Transfer, 0)
	for _, t := range r.S.Transfers {
		if t.TenantID == tenantID {
			out = append(out, t)
		}
	}
	total := len(out)
	if offset >= len(out) {
		return nil, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

// CountRepo implements ports.CountRepository.
type CountRepo struct{ S *Store }

func (r *CountRepo) Create(_ context.Context, s domain.CountSession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Counts[s.ID] = s
	return nil
}

func (r *CountRepo) Update(_ context.Context, s domain.CountSession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Counts[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Counts[s.ID] = s
	return nil
}

func (r *CountRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.CountSession, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Counts[id]
	if !ok || s.TenantID != tenantID {
		return domain.CountSession{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *CountRepo) List(_ context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.CountSession, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.CountSession, 0)
	for _, s := range r.S.Counts {
		if s.TenantID == tenantID && (warehouseID == uuid.Nil || s.WarehouseID == warehouseID) {
			out = append(out, s)
		}
	}
	total := len(out)
	if offset >= len(out) {
		return nil, total, nil
	}
	end := offset + limit
	if limit <= 0 || end > len(out) {
		end = len(out)
	}
	return out[offset:end], total, nil
}

// ReturnRepo implements ports.ReturnRepository.
type ReturnRepo struct{ S *Store }

func (r *ReturnRepo) Create(_ context.Context, ret domain.InventoryReturn) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Returns[ret.ID] = ret
	return nil
}

func (r *ReturnRepo) Update(_ context.Context, ret domain.InventoryReturn) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Returns[ret.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Returns[ret.ID] = ret
	return nil
}

func (r *ReturnRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.InventoryReturn, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	ret, ok := r.S.Returns[id]
	if !ok || ret.TenantID != tenantID {
		return domain.InventoryReturn{}, domain.ErrNotFound
	}
	return ret, nil
}

// ForecastRepo implements ports.ForecastRepository.
type ForecastRepo struct{ S *Store }

func (r *ForecastRepo) Upsert(_ context.Context, f domain.StockForecast) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	key := tenantKey(f.TenantID, f.WarehouseID.String(), f.VariantID.String(), f.HorizonStart.Format("2006-01-02"), f.HorizonEnd.Format("2006-01-02"), f.ModelID)
	if id, ok := r.S.ForecastKey[key]; ok {
		f.ID = id
		cur := r.S.Forecasts[id]
		f.CreatedAt = cur.CreatedAt
	} else {
		r.S.ForecastKey[key] = f.ID
	}
	r.S.Forecasts[f.ID] = f
	return nil
}

func (r *ForecastRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.StockForecast, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	f, ok := r.S.Forecasts[id]
	if !ok || f.TenantID != tenantID {
		return domain.StockForecast{}, domain.ErrNotFound
	}
	return f, nil
}

func (r *ForecastRepo) List(_ context.Context, tenantID, warehouseID, variantID uuid.UUID) ([]domain.StockForecast, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := make([]domain.StockForecast, 0)
	for _, f := range r.S.Forecasts {
		if f.TenantID == tenantID && f.WarehouseID == warehouseID && f.VariantID == variantID {
			out = append(out, f)
		}
	}
	return out, nil
}
