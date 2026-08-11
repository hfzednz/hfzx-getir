// Package warehouse provides an HTTP warehouse-service client.
package warehouse

import (
	"log/slog"

	"github.com/nexora/order-service/internal/adapters/httpclients"
	"github.com/nexora/order-service/internal/app/ports"
)

// NewClient returns an HTTP warehouse client for the given base URL.
func NewClient(baseURL string, log *slog.Logger) ports.WarehouseClient {
	if log != nil && baseURL != "" {
		log.Info("warehouse.client.http", "baseURL", baseURL)
	}
	return httpclients.WarehouseHTTP{Client: httpclients.New(baseURL)}
}
