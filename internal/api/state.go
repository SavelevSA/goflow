package api

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"goflow/internal/model"
)

type previewSimulation struct {
	density atomic.Int32
	paused  atomic.Bool
	mode    atomic.Value
	seed    atomic.Int64
}

// Создать снимок
func (sim *previewSimulation) Snapshot() model.Snapshot {
	snapshot := model.Snapshot{
		Mode: sim.mode.Load().(model.Mode), Density: int(sim.density.Load()),
		Paused: sim.paused.Load(), Seed: sim.seed.Load(),
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

func (sim *previewSimulation) SetPaused(paused bool) error {
	sim.paused.Store(paused)
	return nil
}

func (sim *previewSimulation) SetMode(mode model.Mode) error {
	sim.mode.Store(mode)
	return nil
}

func (sim *previewSimulation) Reset(seed int64) error {
	sim.seed.Store(seed)
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
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid request body"})
			return
		}

		if body.Density < 1 || body.Density > 5 {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "density is not between 1 and 5"})
			return
		}

		if sim.SetDensity(body.Density) != nil {
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "could not set density"})
			return
		}

		snapshot := sim.Snapshot()
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}

// POST-запрос переключения паузы симуляции
func registerPaused(mux *http.ServeMux, sim interface {
	Snapshot() model.Snapshot
	SetPaused(bool) error
}) {
	mux.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		var body struct {
			Paused *bool `json:"paused"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil || body.Paused == nil {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "expected a boolean paused field"})
			return
		}

		if sim.SetPaused(*body.Paused) != nil {
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "could not set pause"})
			return
		}

		snapshot := sim.Snapshot()
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}

// POST-запрос переключения режима работы
func registerMode(mux *http.ServeMux, sim interface {
	Snapshot() model.Snapshot
	SetMode(model.Mode) error
}) {
	mux.HandleFunc("/api/mode", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		var body struct {
			Mode model.Mode `json:"mode"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil || (body.Mode != model.ModeFixed && body.Mode != model.ModeAdaptive) {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "expected fixed or adaptive mode"})
			return
		}

		if err := sim.SetMode(body.Mode); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "could not set mode"})
			return
		}

		snapshot := sim.Snapshot()
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}

func registerReset(mux *http.ServeMux, sim interface {
	Snapshot() model.Snapshot
	Reset(int64) error
}) {
	mux.HandleFunc("/api/reset", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			return
		}

		var body struct {
			Seed *int64 `json:"seed"`
		}
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil || body.Seed == nil {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "expected int value"})
			return
		}

		if sim.Reset(*body.Seed) != nil {
			w.WriteHeader(500)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "could not reset"})
			return
		}

		snapshot := sim.Snapshot()
		_ = json.NewEncoder(w).Encode(snapshot)
	})
}

func RegisterPreview(mux *http.ServeMux) {
	sim := &previewSimulation{}
	sim.density.Store(3)
	sim.mode.Store(model.ModeFixed)
	sim.seed.Store(42)
	RegisterRoutes(mux, sim)
}

func RegisterRoutes(mux *http.ServeMux, sim interface {
	Snapshot() model.Snapshot
	SetDensity(int) error
	SetPaused(bool) error
	SetMode(model.Mode) error
	Reset(int64) error
}) {
	registerState(mux, sim)
	registerDensity(mux, sim)
	registerPaused(mux, sim)
	registerMode(mux, sim)
	registerReset(mux, sim)
}
