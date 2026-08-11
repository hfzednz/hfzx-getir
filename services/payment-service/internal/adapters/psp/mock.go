// Package psp provides MockPSP and Failover router adapters.
package psp

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/google/uuid"
	"github.com/nexora/payment-service/internal/app/ports"
	"github.com/nexora/payment-service/internal/domain"
)

// MockPSP is an in-process PSP that can be forced to fail.
type MockPSP struct {
	name       string
	failAuth   atomic.Bool
	failCap    atomic.Bool
	failVoid   atomic.Bool
	failRef    atomic.Bool
	authCalls  atomic.Int64
	capCalls   atomic.Int64
	mu         sync.Mutex
	auths      map[string]ports.AuthorizeResult
}

// NewMock returns a named MockPSP that succeeds by default.
func NewMock(name string) *MockPSP {
	return &MockPSP{name: name, auths: make(map[string]ports.AuthorizeResult)}
}

func (m *MockPSP) Name() string { return m.name }

// SetFailAuthorize toggles authorize failures.
func (m *MockPSP) SetFailAuthorize(v bool) { m.failAuth.Store(v) }

// SetFailCapture toggles capture failures.
func (m *MockPSP) SetFailCapture(v bool) { m.failCap.Store(v) }

// AuthCalls returns authorize invocation count.
func (m *MockPSP) AuthCalls() int64 { return m.authCalls.Load() }

func (m *MockPSP) Authorize(_ context.Context, req ports.AuthorizeRequest) (ports.AuthorizeResult, error) {
	m.authCalls.Add(1)
	if m.failAuth.Load() {
		return ports.AuthorizeResult{Success: false, ErrorCode: "provider_down", ErrorMessage: m.name + " authorize failed"}, nil
	}
	ref := "auth_" + m.name + "_" + uuid.NewString()[:8]
	res := ports.AuthorizeResult{ProviderRef: ref, Success: true}
	m.mu.Lock()
	m.auths[req.IdempotencyKey] = res
	m.mu.Unlock()
	return res, nil
}

func (m *MockPSP) Capture(_ context.Context, req ports.CaptureRequest) (ports.CaptureResult, error) {
	m.capCalls.Add(1)
	if m.failCap.Load() {
		return ports.CaptureResult{Success: false, ErrorCode: "provider_down", ErrorMessage: m.name + " capture failed"}, nil
	}
	return ports.CaptureResult{ProviderRef: "cap_" + req.ProviderRef, Success: true}, nil
}

func (m *MockPSP) Void(_ context.Context, req ports.VoidRequest) (ports.VoidResult, error) {
	if m.failVoid.Load() {
		return ports.VoidResult{Success: false, ErrorCode: "provider_down", ErrorMessage: m.name + " void failed"}, nil
	}
	return ports.VoidResult{ProviderRef: "void_" + req.ProviderRef, Success: true}, nil
}

func (m *MockPSP) Refund(_ context.Context, req ports.RefundRequest) (ports.RefundResult, error) {
	if m.failRef.Load() {
		return ports.RefundResult{Success: false, ErrorCode: "provider_down", ErrorMessage: m.name + " refund failed"}, nil
	}
	return ports.RefundResult{ProviderRef: "ref_" + req.ProviderRef, Success: true}, nil
}

var _ ports.PSPClient = (*MockPSP)(nil)

// Failover tries providers in order until one succeeds.
type Failover struct {
	providers []ports.PSPClient
	lastUsed  atomic.Value // string
}

// NewFailover builds a router over ordered PSP clients.
func NewFailover(providers ...ports.PSPClient) *Failover {
	f := &Failover{providers: providers}
	f.lastUsed.Store("")
	return f
}

func (f *Failover) Name() string {
	if v, ok := f.lastUsed.Load().(string); ok && v != "" {
		return v
	}
	if len(f.providers) > 0 {
		return f.providers[0].Name()
	}
	return "failover"
}

// LastUsed returns the provider name that last succeeded.
func (f *Failover) LastUsed() string {
	if v, ok := f.lastUsed.Load().(string); ok {
		return v
	}
	return ""
}

func (f *Failover) Authorize(ctx context.Context, req ports.AuthorizeRequest) (ports.AuthorizeResult, error) {
	if len(f.providers) == 0 {
		return ports.AuthorizeResult{}, fmt.Errorf("%w: no providers", domain.ErrNoProviderRoute)
	}
	var last ports.AuthorizeResult
	var lastErr error
	for i, p := range f.providers {
		res, err := p.Authorize(ctx, req)
		if err == nil && res.Success {
			f.lastUsed.Store(p.Name())
			return res, nil
		}
		last = res
		lastErr = err
		_ = i
	}
	if lastErr != nil {
		return last, lastErr
	}
	return last, fmt.Errorf("%w: all providers failed", domain.ErrPSPFailed)
}

func (f *Failover) Capture(ctx context.Context, req ports.CaptureRequest) (ports.CaptureResult, error) {
	p := f.pick()
	res, err := p.Capture(ctx, req)
	if err == nil && res.Success {
		f.lastUsed.Store(p.Name())
	}
	return res, err
}

func (f *Failover) Void(ctx context.Context, req ports.VoidRequest) (ports.VoidResult, error) {
	p := f.pick()
	res, err := p.Void(ctx, req)
	if err == nil && res.Success {
		f.lastUsed.Store(p.Name())
	}
	return res, err
}

func (f *Failover) Refund(ctx context.Context, req ports.RefundRequest) (ports.RefundResult, error) {
	p := f.pick()
	res, err := p.Refund(ctx, req)
	if err == nil && res.Success {
		f.lastUsed.Store(p.Name())
	}
	return res, err
}

func (f *Failover) pick() ports.PSPClient {
	name, _ := f.lastUsed.Load().(string)
	for _, p := range f.providers {
		if p.Name() == name {
			return p
		}
	}
	return f.providers[0]
}

var _ ports.PSPClient = (*Failover)(nil)
