package memory

import (
	"context"

	"github.com/google/uuid"
	"github.com/nexora/catalog-service/internal/app/ports"
	"github.com/nexora/catalog-service/internal/domain"
)

// SupplierRepo is an in-memory SupplierRepository.
type SupplierRepo struct{ S *Store }

func (r *SupplierRepo) Create(_ context.Context, s domain.Supplier) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	k := tenantKey(s.TenantID, s.Code)
	if _, ok := r.S.SupplierCode[k]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Suppliers[s.ID] = s
	r.S.SupplierCode[k] = s.ID
	return nil
}

func (r *SupplierRepo) Update(_ context.Context, s domain.Supplier) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Suppliers[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Suppliers[s.ID] = s
	return nil
}

func (r *SupplierRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Supplier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Suppliers[id]
	if !ok || s.TenantID != tenantID || s.DeletedAt != nil {
		return domain.Supplier{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *SupplierRepo) GetByCode(_ context.Context, tenantID uuid.UUID, code string) (domain.Supplier, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.SupplierCode[tenantKey(tenantID, code)]
	if !ok {
		return domain.Supplier{}, domain.ErrNotFound
	}
	s := r.S.Suppliers[id]
	if s.DeletedAt != nil {
		return domain.Supplier{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *SupplierRepo) List(_ context.Context, tenantID uuid.UUID, limit, offset int) ([]domain.Supplier, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.Supplier
	for _, s := range r.S.Suppliers {
		if s.TenantID == tenantID && s.DeletedAt == nil {
			all = append(all, s)
		}
	}
	total := len(all)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *SupplierRepo) LinkProduct(_ context.Context, sp domain.SupplierProduct) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.SupplierProducts[sp.ID] = sp
	return nil
}

func (r *SupplierRepo) ListProducts(_ context.Context, tenantID, supplierID uuid.UUID) ([]domain.SupplierProduct, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.SupplierProduct
	for _, sp := range r.S.SupplierProducts {
		if sp.TenantID == tenantID && sp.SupplierID == supplierID {
			out = append(out, sp)
		}
	}
	return out, nil
}

var _ ports.SupplierRepository = (*SupplierRepo)(nil)
