package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	httpadapter "github.com/nexora/bff-courier/internal/adapters/http"
	"github.com/nexora/bff-courier/internal/app"
)

func main() {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8112"
	}
	d := &app.Deps{}
	srv := httpadapter.NewServer(addr, d)
	go func() {
		log.Println("listen", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	_ = srv.Close()
}
