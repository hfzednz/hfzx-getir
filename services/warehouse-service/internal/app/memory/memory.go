// Package memory provides in-memory port implementations for tests and dev mode.
package memory

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/warehouse-service/internal/app/ports"
	"github.com/nexora/warehouse-service/internal/domain"
)

// Store is a shared in-memory database for all repositories.
type Store struct {
	mu sync.RWMutex

	Fulfillments   map[uuid.UUID]domain.FulfillmentOrder
	FulfillByExt   map[string]uuid.UUID // tenant:externalOrderID
	Tasks          map[uuid.UUID]domain.Task
	PickSessions   map[uuid.UUID]domain.PickSession
	PickByTask     map[uuid.UUID]uuid.UUID
	PackSessions   map[uuid.UUID]domain.PackSession
	PackByTask     map[uuid.UUID]uuid.UUID
	Dispatches     map[uuid.UUID]domain.DispatchUnit
	DispatchByFulf map[uuid.UUID]uuid.UUID
	Stations       map[uuid.UUID]domain.Station
	Employees      map[uuid.UUID]domain.Employee
	Shifts         map[uuid.UUID]domain.Shift
	ActiveShift    map[uuid.UUID]uuid.UUID // employeeID -> shiftID
	Equipment      map[uuid.UUID]domain.Equipment
	Inspections    map[uuid.UUID]domain.QCInspection
	Labels         map[uuid.UUID]domain.Label

	Events           []publishedEvent
	SoftReserveCalls []ports.SoftReserveRequest
	ConsumeCalls     []ports.ConsumeRequest
	ConfirmCalls     []ports.ConfirmHardRequest
	ReleaseCalls     []ports.ReleaseRequest
}

type publishedEvent struct {
	Topic   string
	Key     string
	Payload any
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{
		Fulfillments:   make(map[uuid.UUID]domain.FulfillmentOrder),
		FulfillByExt:   make(map[string]uuid.UUID),
		Tasks:          make(map[uuid.UUID]domain.Task),
		PickSessions:   make(map[uuid.UUID]domain.PickSession),
		PickByTask:     make(map[uuid.UUID]uuid.UUID),
		PackSessions:   make(map[uuid.UUID]domain.PackSession),
		PackByTask:     make(map[uuid.UUID]uuid.UUID),
		Dispatches:     make(map[uuid.UUID]domain.DispatchUnit),
		DispatchByFulf: make(map[uuid.UUID]uuid.UUID),
		Stations:       make(map[uuid.UUID]domain.Station),
		Employees:      make(map[uuid.UUID]domain.Employee),
		Shifts:         make(map[uuid.UUID]domain.Shift),
		ActiveShift:    make(map[uuid.UUID]uuid.UUID),
		Equipment:      make(map[uuid.UUID]domain.Equipment),
		Inspections:    make(map[uuid.UUID]domain.QCInspection),
		Labels:         make(map[uuid.UUID]domain.Label),
	}
}

func tenantKey(tenantID uuid.UUID, parts ...string) string {
	b := strings.Builder{}
	b.WriteString(tenantID.String())
	for _, p := range parts {
		b.WriteByte(':')
		b.WriteString(p)
	}
	return b.String()
}

// Clock is a fixed/mutable clock for tests.
type Clock struct{ T time.Time }

func (c *Clock) Now() time.Time {
	if c.T.IsZero() {
		return time.Now().UTC()
	}
	return c.T.UTC()
}

// IDGen generates random UUIDs.
type IDGen struct{}

func (IDGen) New() uuid.UUID { return uuid.New() }

// EventPublisher records published events.
type EventPublisher struct{ S *Store }

func (p *EventPublisher) Publish(_ context.Context, topic, key string, payload any) error {
	p.S.mu.Lock()
	defer p.S.mu.Unlock()
	p.S.Events = append(p.S.Events, publishedEvent{Topic: topic, Key: key, Payload: payload})
	return nil
}

// InventoryClient is an in-memory stub of inventory-service SoftReserve/Confirm/Release/Consume.
type InventoryClient struct{ S *Store }

func (c *InventoryClient) SoftReserve(_ context.Context, req ports.SoftReserveRequest) (ports.SoftReserveResult, error) {
	c.S.mu.Lock()
	defer c.S.mu.Unlock()
	c.S.SoftReserveCalls = append(c.S.SoftReserveCalls, req)
	id := uuid.New()
	return ports.SoftReserveResult{ReservationID: id, Status: "active"}, nil
}

func (c *InventoryClient) ConfirmHard(_ context.Context, req ports.ConfirmHardRequest) error {
	c.S.mu.Lock()
	defer c.S.mu.Unlock()
	c.S.ConfirmCalls = append(c.S.ConfirmCalls, req)
	return nil
}

func (c *InventoryClient) Release(_ context.Context, req ports.ReleaseRequest) error {
	c.S.mu.Lock()
	defer c.S.mu.Unlock()
	c.S.ReleaseCalls = append(c.S.ReleaseCalls, req)
	return nil
}

func (c *InventoryClient) Consume(_ context.Context, req ports.ConsumeRequest) error {
	c.S.mu.Lock()
	defer c.S.mu.Unlock()
	c.S.ConsumeCalls = append(c.S.ConsumeCalls, req)
	return nil
}

// SoftReserveCallCount returns how many SoftReserve calls were made.
func (c *InventoryClient) SoftReserveCallCount() int {
	c.S.mu.RLock()
	defer c.S.mu.RUnlock()
	return len(c.S.SoftReserveCalls)
}

// SoftReserveCalls returns a copy of SoftReserve requests recorded by the stub.
func (c *InventoryClient) SoftReserveCalls() []ports.SoftReserveRequest {
	c.S.mu.RLock()
	defer c.S.mu.RUnlock()
	return append([]ports.SoftReserveRequest(nil), c.S.SoftReserveCalls...)
}

// ConsumeCallCount returns how many Consume calls were made.
func (c *InventoryClient) ConsumeCallCount() int {
	c.S.mu.RLock()
	defer c.S.mu.RUnlock()
	return len(c.S.ConsumeCalls)
}

// RouteOptimizer sorts lines by location code (AI stub).
type RouteOptimizer struct{}

func (RouteOptimizer) OptimizePickRoute(_ context.Context, _ uuid.UUID, lines []ports.RouteLine) ([]ports.RouteLine, error) {
	out := append([]ports.RouteLine(nil), lines...)
	sort.SliceStable(out, func(i, j int) bool {
		return strings.Compare(out[i].LocationCode, out[j].LocationCode) < 0
	})
	for i := range out {
		out[i].Sequence = i + 1
	}
	return out, nil
}

var (
	_ ports.EventPublisher  = (*EventPublisher)(nil)
	_ ports.InventoryClient = (*InventoryClient)(nil)
	_ ports.RouteOptimizer  = (RouteOptimizer{})
	_ ports.Clock           = (*Clock)(nil)
	_ ports.IDGen           = (IDGen{})
)
