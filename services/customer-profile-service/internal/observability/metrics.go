// Package observability provides minimal metrics counter stubs.
package observability

import "sync/atomic"

// Counters is a process-local metrics snapshot (no exporter wired yet).
type Counters struct {
	HTTPRequests   atomic.Int64
	HTTPErrors     atomic.Int64
	ProfilesCreated atomic.Int64
	EventsPublished atomic.Int64
	CacheHits      atomic.Int64
	CacheMisses    atomic.Int64
}

// Default is the shared process counters.
var Default Counters

// Snapshot returns a point-in-time view of Default counters.
func Snapshot() map[string]int64 {
	return map[string]int64{
		"httpRequests":    Default.HTTPRequests.Load(),
		"httpErrors":      Default.HTTPErrors.Load(),
		"profilesCreated": Default.ProfilesCreated.Load(),
		"eventsPublished": Default.EventsPublished.Load(),
		"cacheHits":       Default.CacheHits.Load(),
		"cacheMisses":     Default.CacheMisses.Load(),
	}
}
