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
	for J := range e.state.Intersections {
		e.state.Intersections[J].QueueEW = 0
		e.state.Intersections[J].QueueNS = 0
	}
	for i := range e.state.Vehicles {
		e.state.Vehicles[i].Waiting = false
		var dx float64
		var dy float64
		switch e.state.Vehicles[i].Direction {
		case model.East:
			dx = 1
		case model.West:
			dx = -1
		case model.North:
			dy = -1
		case model.South:
			dy = 1
		default:
			continue
		}
		for J := range e.state.Intersections {
			intersection := e.state.Intersections[J]
			if dx != 0 && e.state.Vehicles[i].Y == intersection.Y && (e.state.Vehicles[i].X+dx) == intersection.X && intersection.Phase == model.PhaseNS {
				e.state.Vehicles[i].Waiting = true
				e.state.Intersections[J].QueueEW++
				e.state.Vehicles[i].WaitingTime += tickSeconds
				break
			}
			if dy != 0 && e.state.Vehicles[i].X == intersection.X && (e.state.Vehicles[i].Y+dy) == intersection.Y && intersection.Phase == model.PhaseEW {
				e.state.Vehicles[i].Waiting = true
				e.state.Intersections[J].QueueNS++
				e.state.Vehicles[i].WaitingTime += tickSeconds
				break
			}
		}
		if !e.state.Vehicles[i].Waiting {
			e.state.Vehicles[i].X += dx
			e.state.Vehicles[i].Y += dy
		}
	}

	if e.state.Tick == 1 {
		e.state.Vehicles = append(e.state.Vehicles, model.Vehicle{ID: 1, X: 0, Y: 25, Direction: model.East})
	}
}
