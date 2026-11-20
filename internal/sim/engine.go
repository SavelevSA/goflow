package sim

import (
	"slices"

	"goflow/internal/model"
)

const tickSeconds = 0.25

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

func (e *Engine) step() {
	e.state.Tick++
	e.state.Intersections[0].QueueEW = 0
	for i := range e.state.Vehicles {
		e.state.Vehicles[i].Waiting = false
		if e.state.Vehicles[i].Direction == model.East {
			intersection := e.state.Intersections[0]
			if e.state.Vehicles[i].Y == intersection.Y && (e.state.Vehicles[i].X+1) == intersection.X && intersection.Phase == model.PhaseNS {
				e.state.Vehicles[i].Waiting = true
				e.state.Intersections[0].QueueEW++
				e.state.Vehicles[i].WaitingTime += tickSeconds
				continue
			}
			e.state.Vehicles[i].X++
		}
	}
	if e.state.Tick == 1 {
		e.state.Vehicles = append(e.state.Vehicles, model.Vehicle{ID: 1, X: 0, Y: 25, Direction: model.East})
	}
}
