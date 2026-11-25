package sim

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"slices"
	"sync"
	"time"

	"goflow/internal/control"
	"goflow/internal/metrics"
	"goflow/internal/model"
)

const (
	tickSeconds        = 0.25
	tickInterval       = 125 * time.Millisecond
	spawnIntervalTicks = 10
	vehicleStep        = 1.0

	laneOffset           = 3.0
	intersectionHalfSize = 6.0
	vehicleHalfLength    = 3.0
	vehicleGap           = 1.0
	// minVehicleSpacing    = 2*vehicleHalfLength + vehicleGap

	carLength   = 6.0
	truckLength = 9.0
)

type crossingRecord struct {
	IntersectionID int
	Mode           model.Mode
}

type Engine struct {
	state                model.Snapshot
	completedWaitingTime float64
	mu                   sync.Mutex
	cancel               context.CancelFunc
	done                 chan struct{}
	crossings            map[int]crossingRecord
	waitingSeen          map[model.Mode]map[int]bool
	rng                  *rand.Rand
}

func NewEngine() *Engine {
	const defaultSeed int64 = 42
	return &Engine{state: model.Snapshot{
		Density:  3,
		Seed:     defaultSeed,
		Mode:     model.ModeFixed,
		Vehicles: []model.Vehicle{},
		Intersections: []model.Intersection{
			{ID: 1, X: 25, Y: 25, Phase: model.PhaseNS},
			{ID: 2, X: 75, Y: 25, Phase: model.PhaseNS},
			{ID: 3, X: 25, Y: 75, Phase: model.PhaseNS},
			{ID: 4, X: 75, Y: 75, Phase: model.PhaseNS},
		},
	},
		rng:       rand.New(rand.NewSource(defaultSeed)),
		crossings: make(map[int]crossingRecord),
		waitingSeen: map[model.Mode]map[int]bool{
			model.ModeFixed:    {},
			model.ModeAdaptive: {},
		},
	}
}

func (e *Engine) Snapshot() model.Snapshot {
	e.mu.Lock()
	defer e.mu.Unlock()
	snapshot := e.state
	snapshot.Intersections = slices.Clone(e.state.Intersections)
	snapshot.Vehicles = slices.Clone(e.state.Vehicles)
	return snapshot
}

func (e *Engine) step() {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.state.Paused {
		return
	}
	e.state.Tick++

	//статистика по режиму
	switch e.state.Mode {
	case model.ModeFixed:
		e.state.ModeStats.Fixed.ActiveTicks++
	case model.ModeAdaptive:
		e.state.ModeStats.Adaptive.ActiveTicks++
	}

	for i := range e.state.Intersections {
		e.state.Intersections[i].QueueEW = 0
		e.state.Intersections[i].QueueNS = 0
	}
	vehiclesBeforeMove := slices.Clone(e.state.Vehicles)
	decisions := make([]movementDecision, len(vehiclesBeforeMove))

	lanes := make(map[laneKey][]int)
	for i, vehicle := range vehiclesBeforeMove {
		lane := laneOf(vehicle)
		lanes[lane] = append(lanes[lane], i)
	}

	intersections := slices.Clone(e.state.Intersections)
	var workers sync.WaitGroup

	for _, indices := range lanes {
		workers.Add(1)

		go func(indices []int) {
			defer workers.Done()

			for _, i := range indices {
				decisions[i] = decideMovement(
					vehiclesBeforeMove[i],
					vehiclesBeforeMove,
					intersections,
				)
			}
		}(indices)
	}

	workers.Wait()
	resolveIntersectionEntries(vehiclesBeforeMove, intersections, decisions)
	for i, decision := range decisions {
		vehicle := &e.state.Vehicles[i]
		vehicle.Waiting = decision.waitingAt != -1 || decision.blockedBy != -1
		if vehicle.Waiting {
			vehicle.WaitingTime += tickSeconds
			//статистика по режиму
			stats := &e.state.ModeStats.Fixed

			if e.state.Mode == model.ModeAdaptive {
				stats = &e.state.ModeStats.Adaptive
			}

			stats.WaitingSeconds += tickSeconds

			if !e.waitingSeen[e.state.Mode][vehicle.ID] {
				e.waitingSeen[e.state.Mode][vehicle.ID] = true
				stats.WaitingVehicles++
			}

			stats.AverageWaitingTime =
				stats.WaitingSeconds / float64(stats.WaitingVehicles)
		}

		if !vehicle.Waiting {
			vehicle.X += decision.dx * vehicleStep
			vehicle.Y += decision.dy * vehicleStep
			if decision.turnAt != 0 {
				vehicle.TurnedAt = decision.turnAt
				vehicle.TurnAngle = decision.turnAngle
			}
			if decision.newDirection != "" {
				vehicle.Direction = decision.newDirection
				vehicle.NextTurn = model.TurnStraight
			}
		}
	}
	// Учёт начала и завершения пересечений
	for _, vehicle := range e.state.Vehicles {
		record, exists := e.crossings[vehicle.ID]

		if exists {
			for _, intersection := range e.state.Intersections {
				if intersection.ID != record.IntersectionID {
					continue
				}

				if !occupiesIntersection(vehicle, intersection) {
					switch record.Mode {
					case model.ModeFixed:
						e.state.ModeStats.Fixed.CompletedCrossings++
					case model.ModeAdaptive:
						e.state.ModeStats.Adaptive.CompletedCrossings++
					}

					delete(e.crossings, vehicle.ID)
				}
				break
			}
			continue
		}

		for _, intersection := range e.state.Intersections {
			if occupiesIntersection(vehicle, intersection) {
				e.crossings[vehicle.ID] = crossingRecord{
					IntersectionID: intersection.ID,
					Mode:           e.state.Mode,
				}
				break
			}
		}
	}

	for i := range e.state.Vehicles {
		if !e.state.Vehicles[i].Waiting {
			continue
		}
		current := i
		for current != -1 && decisions[current].waitingAt == -1 {
			current = decisions[current].blockedBy
		}
		if current == -1 {
			continue
		}

		intersection := &e.state.Intersections[decisions[current].waitingAt]
		switch e.state.Vehicles[i].Direction {
		case model.East, model.West:
			intersection.QueueEW++
		case model.North, model.South:
			intersection.QueueNS++
		}
	}
	for i := len(e.state.Vehicles); i > 0; i-- {
		if e.state.Vehicles[i-1].X > 100 || e.state.Vehicles[i-1].Y > 100 || e.state.Vehicles[i-1].X < 0 || e.state.Vehicles[i-1].Y < 0 {
			e.completedWaitingTime += e.state.Vehicles[i-1].WaitingTime
			e.state.Vehicles = slices.Delete(e.state.Vehicles, i-1, i)
			e.state.Statistics.TotalCompleted++
		}
	}

	e.spawnVehicle()

	totalQueue := 0
	for _, intersection := range e.state.Intersections {
		totalQueue += intersection.QueueEW + intersection.QueueNS
	}

	counters := metrics.Counters{
		TotalSpawned:         e.state.Statistics.TotalSpawned,
		TotalCompleted:       e.state.Statistics.TotalCompleted,
		ActiveVehicles:       len(e.state.Vehicles),
		TotalQueue:           totalQueue,
		CompletedWaitingTime: e.completedWaitingTime,
	}
	e.state.Statistics = metrics.Calculate(counters)

	var controller control.Controller
	if e.state.Mode == model.ModeAdaptive {
		controller = &control.AdaptiveController{MinDuration: 5, MaxDuration: 15, QueueGap: 2}
	} else {
		controller = &control.FixedController{Duration: 20}
	}
	for i := range e.state.Intersections {
		e.state.Intersections[i].PhaseElapsed += tickSeconds
		nextPhase := controller.NextPhase(e.state.Intersections[i])
		if nextPhase != e.state.Intersections[i].Phase {
			e.state.Intersections[i].Phase = nextPhase
			e.state.Intersections[i].PhaseElapsed = 0
		}
	}
}

type movementDecision struct {
	dx, dy       float64
	waitingAt    int
	blockedBy    int
	enteringAt   int
	turnAt       int
	turnAngle    float64
	newDirection model.Direction
}

func directionStep(direction model.Direction) (dx, dy float64) {
	switch direction {
	case model.East:
		return 1, 0
	case model.West:
		return -1, 0
	case model.North:
		return 0, -1
	case model.South:
		return 0, 1
	}
	return 0, 0
}

func decideMovement(vehicle model.Vehicle, traffic []model.Vehicle, intersections []model.Intersection) movementDecision {
	decision := movementDecision{waitingAt: -1, blockedBy: -1, enteringAt: -1}
	decision.dx, decision.dy = directionStep(vehicle.Direction)
	if decision.dx == 0 && decision.dy == 0 {
		return decision
	}
	if vehicle.TurnAngle != 0 {
		for _, intersection := range intersections {
			if intersection.ID == vehicle.TurnedAt {
				advanceTurn(vehicle, intersection, &decision)
				return decision
			}
		}
	}

	for i, intersection := range intersections {
		if mustWaitAtIntersection(vehicle, intersection, decision.dx, decision.dy, traffic) {
			decision.waitingAt = i
			return decision
		}
	}

	for i, other := range traffic {
		if other.TurnAngle != 0 && other.Direction == vehicle.Direction {
			for _, intersection := range intersections {
				if intersection.ID == other.TurnedAt {
					// Keep a following car behind the start of the occupied turn.
					other = turnPoint(other, intersection, 0)
					break
				}
			}
		}
		distance := distanceAhead(vehicle, other, decision.dx, decision.dy)
		if distance > 0 && distance-vehicleStep < safeDistance(vehicle, other) {
			decision.blockedBy = i
			return decision
		}
	}

	planIntersectionMove(vehicle, intersections, &decision)
	return decision
}

func occupiesIntersection(
	vehicle model.Vehicle,
	intersection model.Intersection,
) bool {
	if vehicle.TurnAngle != 0 && vehicle.TurnedAt == intersection.ID {
		return true
	}
	clearance := intersectionHalfSize + halfLength(vehicle)

	switch vehicle.Direction {
	case model.East:
		return vehicle.Y == intersection.Y+laneOffset &&
			math.Abs(vehicle.X-intersection.X) < clearance

	case model.West:
		return vehicle.Y == intersection.Y-laneOffset &&
			math.Abs(vehicle.X-intersection.X) < clearance

	case model.North:
		return vehicle.X == intersection.X+laneOffset &&
			math.Abs(vehicle.Y-intersection.Y) < clearance

	case model.South:
		return vehicle.X == intersection.X-laneOffset &&
			math.Abs(vehicle.Y-intersection.Y) < clearance
	}

	return false
}

func distanceAhead(vehicle, other model.Vehicle, dx, dy float64) float64 {
	if vehicle.Direction == other.Direction {
		if dx != 0 && vehicle.Y == other.Y {
			return (other.X - vehicle.X) * dx
		}
		if dy != 0 && vehicle.X == other.X {
			return (other.Y - vehicle.Y) * dy
		}
	}
	return 0
}

func halfLength(vehicle model.Vehicle) float64 {
	if vehicle.Length == 0 {
		return vehicleHalfLength
	}
	return vehicle.Length / 2
}

func safeDistance(first, second model.Vehicle) float64 {
	return halfLength(first) + halfLength(second) + vehicleGap
}

func mustWaitAtIntersection(vehicle model.Vehicle, intersection model.Intersection, dx, dy float64, traffic []model.Vehicle) bool {
	var distance float64
	var greenPhase model.Phase
	if dx != 0 {
		if vehicle.Y != intersection.Y+dx*laneOffset {
			return false
		}
		distance = (intersection.X-vehicle.X)*dx - intersectionHalfSize - halfLength(vehicle)
		greenPhase = model.PhaseEW
	} else {
		if vehicle.X != intersection.X-dy*laneOffset {
			return false
		}
		distance = (intersection.Y-vehicle.Y)*dy - intersectionHalfSize - halfLength(vehicle)
		greenPhase = model.PhaseNS
	}
	if distance < 0 || distance > vehicleStep {
		return false
	}
	if intersection.Phase != greenPhase {
		return true
	}

	for _, other := range traffic {
		switch other.Direction {
		case model.East, model.West:
			if dy != 0 && occupiesIntersection(other, intersection) {
				return true
			}
		case model.North, model.South:
			if dx != 0 && occupiesIntersection(other, intersection) {
				return true
			}
		}
	}
	return false
}

type laneKey struct {
	horizontal bool
	coordinate float64
}

func laneOf(vehicle model.Vehicle) laneKey {
	switch vehicle.Direction {
	case model.East, model.West:
		return laneKey{horizontal: true, coordinate: vehicle.Y}
	default:
		return laneKey{horizontal: false, coordinate: vehicle.X}
	}
}

func canEnterLane(candidate model.Vehicle, traffic []model.Vehicle) bool {
	lane := laneOf(candidate)

	for _, other := range traffic {
		if laneOf(other) != lane {
			continue
		}

		distance := math.Abs(other.X-candidate.X) + math.Abs(other.Y-candidate.Y)
		if distance < safeDistance(candidate, other) {
			return false
		}
	}

	return true
}

func (e *Engine) spawnVehicle() {
	interval := spawnIntervalTicks * 3 / e.state.Density
	if (e.state.Tick-1)%interval != 0 {
		return
	}
	length := carLength
	if e.rng.Float64() < 0.2 {
		length = truckLength
	}

	var available []model.Vehicle
	for _, road := range [...]float64{25, 75} {
		for _, entry := range [...]model.Vehicle{
			{X: 0, Y: road + laneOffset, Direction: model.East},
			{X: 100, Y: road - laneOffset, Direction: model.West},
			{X: road - laneOffset, Y: 0, Direction: model.South},
			{X: road + laneOffset, Y: 100, Direction: model.North},
		} {
			entry.Length = length
			if canEnterLane(entry, e.state.Vehicles) {
				available = append(available, entry)
			}
		}
	}

	if len(available) == 0 {
		return
	}
	vehicle := available[e.rng.Intn(len(available))]
	e.state.Statistics.TotalSpawned++
	vehicle.ID = e.state.Statistics.TotalSpawned
	switch vehicle.ID % 5 {
	case 2:
		vehicle.NextTurn = model.TurnRight
	case 4:
		vehicle.NextTurn = model.TurnLeft
	}
	e.state.Vehicles = append(e.state.Vehicles, vehicle)
}

func (e *Engine) run(ctx context.Context) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			e.step()
		case <-ctx.Done():
			return
		}
	}
}

func (e *Engine) Start(ctx context.Context) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.done != nil {
		return
	}
	runCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	e.cancel = cancel
	e.done = done

	go func() {
		e.run(runCtx)
		cancel()
		e.mu.Lock()
		e.cancel = nil
		e.done = nil
		close(done)
		e.mu.Unlock()
	}()
}

func (e *Engine) Stop() {
	e.mu.Lock()
	cancel := e.cancel
	done := e.done
	e.mu.Unlock()
	if done == nil {
		return
	}
	cancel()
	<-done
}

func (e *Engine) SetPaused(paused bool) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.Paused = paused
	return nil
}

func (e *Engine) SetDensity(density int) error {
	if density < 1 || density > 5 {
		return fmt.Errorf("density must be between 1 and 5, got %d", density)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.state.Density = density
	return nil
}

func (e *Engine) SetMode(mode model.Mode) error {
	if mode != model.ModeFixed && mode != model.ModeAdaptive {
		return fmt.Errorf("invalid mode %q: expected fixed or adaptive", mode)
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	if mode != e.state.Mode {
		for _, vehicle := range e.state.Vehicles {
			for _, intersection := range e.state.Intersections {

				if !occupiesIntersection(vehicle, intersection) {
					continue
				}

				_, exists := e.crossings[vehicle.ID]

				if !exists {
					record := crossingRecord{
						IntersectionID: intersection.ID,
						Mode:           e.state.Mode,
					}

					e.crossings[vehicle.ID] = record
				}
			}
		}
	}

	e.state.Mode = mode
	return nil

}

func (e *Engine) Reset(seed int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	fresh := NewEngine().state
	fresh.Mode = e.state.Mode
	fresh.Density = e.state.Density
	fresh.Paused = e.state.Paused
	fresh.Seed = seed
	e.state = fresh
	e.rng.Seed(seed)
	e.completedWaitingTime = 0
	e.crossings = make(map[int]crossingRecord)
	e.waitingSeen = map[model.Mode]map[int]bool{
		model.ModeFixed:    {},
		model.ModeAdaptive: {},
	}
	return nil
}
