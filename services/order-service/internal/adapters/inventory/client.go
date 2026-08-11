// Package inventory provides an HTTP inventory-service client.
package inventory

import (
	"log/slog"

	"github.com/nexora/order-service/internal/adapters/httpclients"
	"github.com/nexora/order-service/internal/app/ports"
)

// NewClient returns an HTTP inventory client for the given base URL.
func NewClient(baseURL string, log *slog.Logger) ports.InventoryClient {
	if log != nil && baseURL != "" {
		log.Info("inventory.client.http", "baseURL", baseURL)
	}
	return httpclients.InventoryHTTP{Client: httpclients.New(baseURL)}
}
