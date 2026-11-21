//go:build !js

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"goflow/internal/api"

	"goflow/internal/sim"
)

func serve() {
	ctx, interrupt := signal.NotifyContext(context.Background(), os.Interrupt)
	engine := sim.NewEngine()
	engine.Start(ctx)

	defer interrupt()
	defer engine.Stop()

	mux := http.NewServeMux()
	api.RegisterRoutes(mux, engine)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("{\"error\":\"endpoint not found\"}\n"))
	})
	mux.Handle("/", &app.Handler{
		Name:        "GoFlow",
		Description: "Симулятор городского дорожного движения",
		Styles:      []string{"/web/style.css"},
	})

	const address = "127.0.0.1:8080"
	log.Printf("GoFlow открыт на http://%s", address)

	srv := http.Server{
		Addr:    address,
		Handler: mux,
	}

	serverErr := make(chan error, 1)

	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		log.Printf("Ошибка HTTP-сервера: %v", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Ошибка остановки HTTP-сервера: %v", err)
		}
	}
}
