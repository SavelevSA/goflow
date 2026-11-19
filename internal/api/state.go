package api

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"goflow/internal/model"
)

type previewSimulation struct {
	density atomic.Int32
}

// Создать снимок
func (sim *previewSimulation) Snapshot() model.Snapshot {
	snapshot := model.Snapshot{
		Mode: model.ModeFixed, Density: int(sim.density.Load()), Seed: 42,
		Vehicles: []model.Vehicle{},
		Intersections: []model.Intersection{
			{ID: 1, X: 25, Y: 25, Phase: model.PhaseNS},
			{ID: 2, X: 75, Y: 25, Phase: model.PhaseNS},
			{ID: 3, X: 25, Y: 75, Phase: model.PhaseNS},
			{ID: 4, X: 75, Y: 75, Phase: model.PhaseNS},
		},
	}
	return snapshot
}

func (sim *previewSimulation) SetDensity(density int) error {
	sim.density.Store(int32(density))
	return nil
}

// GET-запрос снимка
func registerState(mux *http.ServeMux, sim interface{ Snapshot() model.Snapshot }) {
	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}
		snapshot := sim.Snapshot()
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}

// POST-запрос на изменение плотности потока
func registerDensity(mux *http.ServeMux, sim interface {
	Snapshot() model.Snapshot
	SetDensity(int) error
}) {
	mux.HandleFunc("/api/density", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		var body struct {
			Density int `json:"density"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil {
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "could not set density"})
			return
		}

		if body.Density < 1 || body.Density > 5 {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "density is not between 1 and 5"})
			return
		}

		if sim.SetDensity(body.Density) != nil {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "wrong body"})
			return
		}

		snapshot := sim.Snapshot()
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}

func RegisterPreview(mux *http.ServeMux) {
	sim := &previewSimulation{}
	sim.density.Store(3)
	registerState(mux, sim)
	registerDensity(mux, sim)
}
