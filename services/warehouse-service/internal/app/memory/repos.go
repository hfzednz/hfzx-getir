package memory

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// NewRepos returns all memory repository implementations sharing one store.
func NewRepos(s *Store) (
	ports.FulfillmentRepo,
	ports.TaskRepo,
	ports.PickRepo,
	ports.PackRepo,
	ports.DispatchRepo,
	ports.StationRepo,
	ports.WorkforceRepo,
	ports.EquipmentRepo,
	ports.QCRepo,
	ports.LabelRepo,
) {
	return &fulfillmentRepo{s}, &taskRepo{s}, &pickRepo{s}, &packRepo{s},
		&dispatchRepo{s}, &stationRepo{s}, &workforceRepo{s}, &equipmentRepo{s},
		&qcRepo{s}, &labelRepo{s}
}

// --- Fulfillment ---

type fulfillmentRepo struct{ S *Store }

func (r *fulfillmentRepo) Create(_ context.Context, o domain.FulfillmentOrder) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Fulfillments[o.ID]; ok {
		return domain.ErrAlreadyExists
	}
	key := tenantKey(o.TenantID, o.ExternalOrderID)
	if _, ok := r.S.FulfillByExt[key]; ok {
		return domain.ErrAlreadyExists
	}
	cp := cloneFulfillment(o)
	r.S.Fulfillments[o.ID] = cp
	r.S.FulfillByExt[key] = o.ID
	return nil
}

func (r *fulfillmentRepo) Update(_ context.Context, o domain.FulfillmentOrder) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Fulfillments[o.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Fulfillments[o.ID] = cloneFulfillment(o)
	return nil
}

func (r *fulfillmentRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.FulfillmentOrder, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	o, ok := r.S.Fulfillments[id]
	if !ok || o.TenantID != tenantID {
		return domain.FulfillmentOrder{}, domain.ErrNotFound
	}
	return cloneFulfillment(o), nil
}

func (r *fulfillmentRepo) GetByExternalOrderID(_ context.Context, tenantID uuid.UUID, externalOrderID string) (domain.FulfillmentOrder, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.FulfillByExt[tenantKey(tenantID, externalOrderID)]
	if !ok {
		return domain.FulfillmentOrder{}, domain.ErrNotFound
	}
	o := r.S.Fulfillments[id]
	return cloneFulfillment(o), nil
}

func (r *fulfillmentRepo) List(_ context.Context, f ports.FulfillmentFilter) ([]domain.FulfillmentOrder, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.FulfillmentOrder
	for _, o := range r.S.Fulfillments {
		if o.TenantID != f.TenantID {
			continue
		}
		if f.WarehouseID != nil && o.WarehouseID != *f.WarehouseID {
			continue
		}
		if f.Status != nil && o.Status != *f.Status {
			continue
		}
		all = append(all, cloneFulfillment(o))
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.After(all[j].CreatedAt) })
	total := len(all)
	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 50
	}
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func cloneFulfillment(o domain.FulfillmentOrder) domain.FulfillmentOrder {
	cp := o
	if o.Lines != nil {
		cp.Lines = append([]domain.FulfillmentLine(nil), o.Lines...)
	}
	if o.Metadata != nil {
		cp.Metadata = map[string]any{}
		for k, v := range o.Metadata {
			cp.Metadata[k] = v
		}
	}
	if o.ReservationID != nil {
		id := *o.ReservationID
		cp.ReservationID = &id
	}
	return cp
}

// --- Task ---

type taskRepo struct{ S *Store }

func (r *taskRepo) Create(_ context.Context, t domain.Task) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Tasks[t.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Tasks[t.ID] = cloneTask(t)
	return nil
}

func (r *taskRepo) Update(_ context.Context, t domain.Task) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Tasks[t.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Tasks[t.ID] = cloneTask(t)
	return nil
}

func (r *taskRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Task, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	t, ok := r.S.Tasks[id]
	if !ok || t.TenantID != tenantID {
		return domain.Task{}, domain.ErrNotFound
	}
	return cloneTask(t), nil
}

func (r *taskRepo) List(_ context.Context, f ports.TaskFilter) ([]domain.Task, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.Task
	for _, t := range r.S.Tasks {
		if t.TenantID != f.TenantID || t.WarehouseID != f.WarehouseID {
			continue
		}
		if f.Type != nil && t.Type != *f.Type {
			continue
		}
		if f.Status != nil && t.Status != *f.Status {
			continue
		}
		if f.AssigneeID != nil && (t.AssigneeID == nil || *t.AssigneeID != *f.AssigneeID) {
			continue
		}
		all = append(all, cloneTask(t))
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Priority != all[j].Priority {
			return all[i].Priority > all[j].Priority
		}
		return all[i].CreatedAt.Before(all[j].CreatedAt)
	})
	total := len(all)
	limit, offset := f.Limit, f.Offset
	if limit <= 0 {
		limit = 50
	}
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (r *taskRepo) CountByStatus(_ context.Context, tenantID, warehouseID uuid.UUID) (map[domain.TaskStatus]int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	out := map[domain.TaskStatus]int{}
	for _, t := range r.S.Tasks {
		if t.TenantID == tenantID && t.WarehouseID == warehouseID {
			out[t.Status]++
		}
	}
	return out, nil
}

func cloneTask(t domain.Task) domain.Task {
	cp := t
	if t.History != nil {
		cp.History = append([]domain.TaskHistoryEntry(nil), t.History...)
	}
	if t.AssigneeID != nil {
		id := *t.AssigneeID
		cp.AssigneeID = &id
	}
	if t.StationID != nil {
		id := *t.StationID
		cp.StationID = &id
	}
	return cp
}

// --- Pick ---

type pickRepo struct{ S *Store }

func (r *pickRepo) CreateSession(_ context.Context, s domain.PickSession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.PickSessions[s.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.PickSessions[s.ID] = clonePick(s)
	r.S.PickByTask[s.TaskID] = s.ID
	return nil
}

func (r *pickRepo) UpdateSession(_ context.Context, s domain.PickSession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.PickSessions[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.PickSessions[s.ID] = clonePick(s)
	return nil
}

func (r *pickRepo) GetSessionByID(_ context.Context, tenantID, id uuid.UUID) (domain.PickSession, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.PickSessions[id]
	if !ok || s.TenantID != tenantID {
		return domain.PickSession{}, domain.ErrNotFound
	}
	return clonePick(s), nil
}

func (r *pickRepo) GetSessionByTaskID(_ context.Context, tenantID, taskID uuid.UUID) (domain.PickSession, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.PickByTask[taskID]
	if !ok {
		return domain.PickSession{}, domain.ErrNotFound
	}
	s := r.S.PickSessions[id]
	if s.TenantID != tenantID {
		return domain.PickSession{}, domain.ErrNotFound
	}
	return clonePick(s), nil
}

func clonePick(s domain.PickSession) domain.PickSession {
	cp := s
	if s.Lines != nil {
		cp.Lines = append([]domain.PickLine(nil), s.Lines...)
	}
	return cp
}

// --- Pack ---

type packRepo struct{ S *Store }

func (r *packRepo) Create(_ context.Context, s domain.PackSession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.PackSessions[s.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.PackSessions[s.ID] = s
	r.S.PackByTask[s.TaskID] = s.ID
	return nil
}

func (r *packRepo) Update(_ context.Context, s domain.PackSession) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.PackSessions[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.PackSessions[s.ID] = s
	return nil
}

func (r *packRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.PackSession, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.PackSessions[id]
	if !ok || s.TenantID != tenantID {
		return domain.PackSession{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *packRepo) GetByTaskID(_ context.Context, tenantID, taskID uuid.UUID) (domain.PackSession, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.PackByTask[taskID]
	if !ok {
		return domain.PackSession{}, domain.ErrNotFound
	}
	s := r.S.PackSessions[id]
	if s.TenantID != tenantID {
		return domain.PackSession{}, domain.ErrNotFound
	}
	return s, nil
}

// --- Dispatch ---

type dispatchRepo struct{ S *Store }

func (r *dispatchRepo) Create(_ context.Context, u domain.DispatchUnit) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Dispatches[u.ID]; ok {
		return domain.ErrAlreadyExists
	}
	r.S.Dispatches[u.ID] = u
	r.S.DispatchByFulf[u.FulfillmentID] = u.ID
	return nil
}

func (r *dispatchRepo) Update(_ context.Context, u domain.DispatchUnit) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Dispatches[u.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Dispatches[u.ID] = u
	return nil
}

func (r *dispatchRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.DispatchUnit, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	u, ok := r.S.Dispatches[id]
	if !ok || u.TenantID != tenantID {
		return domain.DispatchUnit{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *dispatchRepo) GetByFulfillmentID(_ context.Context, tenantID, fulfillmentID uuid.UUID) (domain.DispatchUnit, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.DispatchByFulf[fulfillmentID]
	if !ok {
		return domain.DispatchUnit{}, domain.ErrNotFound
	}
	u := r.S.Dispatches[id]
	if u.TenantID != tenantID {
		return domain.DispatchUnit{}, domain.ErrNotFound
	}
	return u, nil
}

func (r *dispatchRepo) ListQueued(_ context.Context, tenantID, warehouseID uuid.UUID, limit, offset int) ([]domain.DispatchUnit, int, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var all []domain.DispatchUnit
	for _, u := range r.S.Dispatches {
		if u.TenantID == tenantID && u.WarehouseID == warehouseID &&
			(u.Status == domain.DispatchStatusQueued || u.Status == domain.DispatchStatusVerified) {
			all = append(all, u)
		}
	}
	sort.Slice(all, func(i, j int) bool { return all[i].CreatedAt.Before(all[j].CreatedAt) })
	total := len(all)
	if limit <= 0 {
		limit = 50
	}
	if offset > total {
		return nil, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

// --- Station ---

type stationRepo struct{ S *Store }

func (r *stationRepo) Create(_ context.Context, s domain.Station) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Stations[s.ID] = s
	return nil
}

func (r *stationRepo) Update(_ context.Context, s domain.Station) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Stations[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Stations[s.ID] = s
	return nil
}

func (r *stationRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Station, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Stations[id]
	if !ok || s.TenantID != tenantID {
		return domain.Station{}, domain.ErrNotFound
	}
	return s, nil
}

func (r *stationRepo) ListByWarehouse(_ context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Station, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Station
	for _, s := range r.S.Stations {
		if s.TenantID == tenantID && s.WarehouseID == warehouseID {
			out = append(out, s)
		}
	}
	return out, nil
}

// --- Workforce ---

type workforceRepo struct{ S *Store }

func (r *workforceRepo) CreateEmployee(_ context.Context, e domain.Employee) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Employees[e.ID] = e
	return nil
}

func (r *workforceRepo) GetEmployee(_ context.Context, tenantID, id uuid.UUID) (domain.Employee, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	e, ok := r.S.Employees[id]
	if !ok || e.TenantID != tenantID {
		return domain.Employee{}, domain.ErrNotFound
	}
	return e, nil
}

func (r *workforceRepo) ListEmployees(_ context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Employee, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Employee
	for _, e := range r.S.Employees {
		if e.TenantID == tenantID && e.WarehouseID == warehouseID {
			out = append(out, e)
		}
	}
	return out, nil
}

func (r *workforceRepo) CreateShift(_ context.Context, s domain.Shift) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	cp := cloneShift(s)
	r.S.Shifts[s.ID] = cp
	if s.Status == domain.ShiftStatusClockedIn || s.Status == domain.ShiftStatusOnBreak {
		r.S.ActiveShift[s.EmployeeID] = s.ID
	}
	return nil
}

func (r *workforceRepo) UpdateShift(_ context.Context, s domain.Shift) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Shifts[s.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Shifts[s.ID] = cloneShift(s)
	if s.Status == domain.ShiftStatusClockedOut {
		delete(r.S.ActiveShift, s.EmployeeID)
	} else {
		r.S.ActiveShift[s.EmployeeID] = s.ID
	}
	return nil
}

func (r *workforceRepo) GetActiveShift(_ context.Context, tenantID, employeeID uuid.UUID) (domain.Shift, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	id, ok := r.S.ActiveShift[employeeID]
	if !ok {
		return domain.Shift{}, domain.ErrNotFound
	}
	s := r.S.Shifts[id]
	if s.TenantID != tenantID {
		return domain.Shift{}, domain.ErrNotFound
	}
	return cloneShift(s), nil
}

func (r *workforceRepo) GetShift(_ context.Context, tenantID, id uuid.UUID) (domain.Shift, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	s, ok := r.S.Shifts[id]
	if !ok || s.TenantID != tenantID {
		return domain.Shift{}, domain.ErrNotFound
	}
	return cloneShift(s), nil
}

func cloneShift(s domain.Shift) domain.Shift {
	cp := s
	if s.Breaks != nil {
		cp.Breaks = append([]domain.BreakInterval(nil), s.Breaks...)
	}
	return cp
}

// --- Equipment ---

type equipmentRepo struct{ S *Store }

func (r *equipmentRepo) Create(_ context.Context, e domain.Equipment) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Equipment[e.ID] = e
	return nil
}

func (r *equipmentRepo) Update(_ context.Context, e domain.Equipment) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Equipment[e.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Equipment[e.ID] = e
	return nil
}

func (r *equipmentRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Equipment, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	e, ok := r.S.Equipment[id]
	if !ok || e.TenantID != tenantID {
		return domain.Equipment{}, domain.ErrNotFound
	}
	return e, nil
}

func (r *equipmentRepo) ListByWarehouse(_ context.Context, tenantID, warehouseID uuid.UUID) ([]domain.Equipment, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.Equipment
	for _, e := range r.S.Equipment {
		if e.TenantID == tenantID && e.WarehouseID == warehouseID {
			out = append(out, e)
		}
	}
	return out, nil
}

// --- QC ---

type qcRepo struct{ S *Store }

func (r *qcRepo) Create(_ context.Context, i domain.QCInspection) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Inspections[i.ID] = i
	return nil
}

func (r *qcRepo) Update(_ context.Context, i domain.QCInspection) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	if _, ok := r.S.Inspections[i.ID]; !ok {
		return domain.ErrNotFound
	}
	r.S.Inspections[i.ID] = i
	return nil
}

func (r *qcRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.QCInspection, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	i, ok := r.S.Inspections[id]
	if !ok || i.TenantID != tenantID {
		return domain.QCInspection{}, domain.ErrNotFound
	}
	return i, nil
}

func (r *qcRepo) ListByFulfillment(_ context.Context, tenantID, fulfillmentID uuid.UUID) ([]domain.QCInspection, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	var out []domain.QCInspection
	for _, i := range r.S.Inspections {
		if i.TenantID == tenantID && i.FulfillmentID == fulfillmentID {
			out = append(out, i)
		}
	}
	return out, nil
}

// --- Label ---

type labelRepo struct{ S *Store }

func (r *labelRepo) Create(_ context.Context, l domain.Label) error {
	r.S.mu.Lock()
	defer r.S.mu.Unlock()
	r.S.Labels[l.ID] = l
	return nil
}

func (r *labelRepo) GetByID(_ context.Context, tenantID, id uuid.UUID) (domain.Label, error) {
	r.S.mu.RLock()
	defer r.S.mu.RUnlock()
	l, ok := r.S.Labels[id]
	if !ok || l.TenantID != tenantID {
		return domain.Label{}, domain.ErrNotFound
	}
	return l, nil
}
