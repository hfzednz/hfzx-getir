package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpadapter "github.com/nexora/bff-customer/internal/adapters/http"
	"github.com/nexora/bff-customer/internal/adapters/httpclients"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8111"
	}
	cfg := httpclients.ConfigFromEnv()
	deps := httpclients.NewDeps(cfg)
	srv := httpadapter.NewServer(addr, deps)
	go func() {
		log.Info("http.listen", "addr", addr,
			"identity", cfg.IdentityURL, "catalog", cfg.CatalogURL, "cart", cfg.CartURL,
			"checkout", cfg.CheckoutURL, "order", cfg.OrderURL, "tracking", cfg.TrackingURL,
			"payment", cfg.PaymentURL,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("serve", "err", err)
			os.Exit(1)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	_ = srv.Close()
}
