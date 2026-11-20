package api

import (
	"encoding/json"
	"net/http"

	"goflow/internal/model"
)

// RegisterPreview exposes a fixed snapshot so the UI can start integration
// while the engine is being developed.
func RegisterPreview(mux *http.ServeMux) {
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		snapshot := model.Snapshot{
			Mode: model.ModeFixed, Density: 3, Seed: 42,
			Vehicles: []model.Vehicle{
				{
					ID:        1,
					X:         10,
					Y:         25,
					Direction: model.East,
				},
			},
			Intersections: []model.Intersection{
				{ID: 1, X: 25, Y: 25, Phase: model.PhaseNS},
				{ID: 2, X: 75, Y: 25, Phase: model.PhaseNS},
				{ID: 3, X: 25, Y: 75, Phase: model.PhaseNS},
				{ID: 4, X: 75, Y: 75, Phase: model.PhaseNS},
			},
		}
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}
