package sim

import (
	"slices"

	"goflow/internal/model"
)

type Engine struct {
	state model.Snapshot
}

func NewEngine() *Engine {
	return &Engine{model.Snapshot{
		Density:  3,
		Seed:     42,
		Mode:     model.ModeFixed,
		Vehicles: []model.Vehicle{},
		Intersections: []model.Intersection{
			{ID: 1, X: 25, Y: 25, Phase: model.PhaseNS},
			{ID: 2, X: 75, Y: 25, Phase: model.PhaseNS},
			{ID: 3, X: 25, Y: 75, Phase: model.PhaseNS},
			{ID: 4, X: 75, Y: 75, Phase: model.PhaseNS},
		},
	}}
}

func (e *Engine) Snapshot() model.Snapshot {
	snapshot := e.state
	snapshot.Intersections = slices.Clone(e.state.Intersections)
	snapshot.Vehicles = slices.Clone(e.state.Vehicles)
	return snapshot
}
