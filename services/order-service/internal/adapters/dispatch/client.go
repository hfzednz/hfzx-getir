// Package dispatch provides an HTTP dispatch-service client.
package dispatch

import (
	"log/slog"

	"github.com/nexora/order-service/internal/adapters/httpclients"
	"github.com/nexora/order-service/internal/app/ports"
)

// NewClient returns an HTTP dispatch client for the given base URL.
func NewClient(baseURL string, log *slog.Logger) ports.DispatchClient {
	if log != nil && baseURL != "" {
		log.Info("dispatch.client.http", "baseURL", baseURL)
	}
	return httpclients.DispatchHTTP{Client: httpclients.New(baseURL)}
}
