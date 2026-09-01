package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpadapter "github.com/nexora/bff-admin/internal/adapters/http"
	"github.com/nexora/bff-admin/internal/adapters/httpclients"
	"github.com/nexora/bff-admin/internal/app"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8114"
	}
	cfg := httpclients.ConfigFromEnv()
	deps := &app.Deps{
		Orders:  httpclients.OrderClient{Client: httpclients.New(cfg.OrderURL)},
		LiveOps: httpclients.LiveOpsClient{Client: httpclients.New(cfg.LiveOpsURL)},
		Catalog: httpclients.CatalogClient{Client: httpclients.New(cfg.CatalogURL)},
		CRM:     httpclients.CRMClient{Client: httpclients.New(cfg.CrmURL)},
		Ledger:  httpclients.LedgerClient{Client: httpclients.NewLedger(cfg.LedgerURL)},
	}
	srv := httpadapter.NewServer(addr, deps)
	go func() {
		log.Println("listen", addr, "order", cfg.OrderURL, "catalog", cfg.CatalogURL, "crm", cfg.CrmURL, "ledger", cfg.LedgerURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	_ = srv.Close()
}
