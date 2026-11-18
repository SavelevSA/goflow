//go:build !js

package main

import (
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v11/pkg/app"

	"goflow/internal/api"
)

func serve() {
	mux := http.NewServeMux()
	api.RegisterPreview(mux)
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
	log.Fatal(http.ListenAndServe(address, mux))
}
