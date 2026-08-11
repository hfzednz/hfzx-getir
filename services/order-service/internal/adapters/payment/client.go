// Package payment provides an HTTP payment-service client.
package payment

import (
	"log/slog"

	"github.com/nexora/order-service/internal/adapters/httpclients"
	"github.com/nexora/order-service/internal/app/ports"
)

// NewClient returns an HTTP payment client for the given base URL.
func NewClient(baseURL string, log *slog.Logger) ports.PaymentClient {
	if log != nil && baseURL != "" {
		log.Info("payment.client.http", "baseURL", baseURL)
	}
	return httpclients.PaymentHTTP{Client: httpclients.New(baseURL)}
}
