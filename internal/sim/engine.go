package sim

import (
	"slices"

	"goflow/internal/model"
)

const tickSeconds = 0.25
const spawnIntervalTicks = 80

type Engine struct {
	state                model.Snapshot
	completedWaitingTime float64
}

func NewEngine() *Engine {
	return &Engine{state: model.Snapshot{
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
	vehiclesBeforeMove := slices.Clone(e.state.Vehicles)
	waitingAt := make([]int, len(e.state.Vehicles))
	blockedBy := make([]int, len(e.state.Vehicles))
	for i := range len(e.state.Vehicles) {
		waitingAt[i] = -1
		blockedBy[i] = -1
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
				waitingAt[i] = J
				e.state.Vehicles[i].WaitingTime += tickSeconds
				break
			}
			if dy != 0 && e.state.Vehicles[i].X == intersection.X && (e.state.Vehicles[i].Y+dy) == intersection.Y && intersection.Phase == model.PhaseEW {
				e.state.Vehicles[i].Waiting = true
				waitingAt[i] = J
				e.state.Vehicles[i].WaitingTime += tickSeconds
				break
			}
		}

		if !e.state.Vehicles[i].Waiting {
			nextX := e.state.Vehicles[i].X + dx
			nextY := e.state.Vehicles[i].Y + dy
			for otherIndex, other := range vehiclesBeforeMove {
				if otherIndex != i && other.X == nextX && other.Y == nextY {
					e.state.Vehicles[i].Waiting = true
					e.state.Vehicles[i].WaitingTime += tickSeconds
					if other.Direction == e.state.Vehicles[i].Direction {
						blockedBy[i] = otherIndex
					}
					break
				}
			}
		}

		if !e.state.Vehicles[i].Waiting {
			e.state.Vehicles[i].X += dx
			e.state.Vehicles[i].Y += dy
		}
	}
	for i := range e.state.Vehicles {
		if !e.state.Vehicles[i].Waiting {
			continue
		}
		current := i
		for current != -1 && waitingAt[current] == -1 {
			current = blockedBy[current]
		}
		if current == -1 {
			continue
		}
		switch e.state.Vehicles[i].Direction {
		case model.East, model.West:
			e.state.Intersections[waitingAt[current]].QueueEW++

		case model.North, model.South:
			e.state.Intersections[waitingAt[current]].QueueNS++
		}
	}
	for i := len(e.state.Vehicles); i > 0; i-- {
		if e.state.Vehicles[i-1].X > 100 || e.state.Vehicles[i-1].Y > 100 || e.state.Vehicles[i-1].X < 0 || e.state.Vehicles[i-1].Y < 0 {
			e.completedWaitingTime += e.state.Vehicles[i-1].WaitingTime
			e.state.Vehicles = slices.Delete(e.state.Vehicles, i-1, i)
			e.state.Statistics.TotalCompleted++
		}
	}

	spawnInterval := spawnIntervalTicks * 3 / e.state.Density
	if (e.state.Tick-1)%spawnInterval == 0 {
		spawnPoints := [4]model.Vehicle{
			{X: 0, Y: 25, Direction: model.East},
			{X: 100, Y: 75, Direction: model.West},
			{X: 25, Y: 0, Direction: model.South},
			{X: 75, Y: 100, Direction: model.North},
		}
		index := e.state.Statistics.TotalSpawned % len(spawnPoints)
		vehicle := spawnPoints[index]
		e.state.Statistics.TotalSpawned++
		vehicle.ID = e.state.Statistics.TotalSpawned
		e.state.Vehicles = append(e.state.Vehicles, vehicle)
	}
	e.state.Statistics.ActiveVehicles = len(e.state.Vehicles)
	totalQueue := 0
	for _, intersection := range e.state.Intersections {
		totalQueue += intersection.QueueEW + intersection.QueueNS
	}
	e.state.Statistics.TotalQueue = totalQueue

}
