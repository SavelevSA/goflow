package sim

import (
	"context"
	"math"
	"reflect"
	"testing"
	"testing/synctest"
	"time"

	"goflow/internal/model"
)

func TestVehicleStopAtRedLight(t *testing.T) {
	engine := NewEngine()
	for i := 0; i < 26; i++ {
		engine.step()
	}
	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) == 0 || snapshot.Vehicles[0].ID != 1 {
		t.Fatal("the first vehicle should still be on the map")
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.X != 15 {
		t.Fatalf("got vehicle at X=%v, want X=15", vehicle.X)
	}
	if !vehicle.Waiting {
		t.Fatalf("got vehicle waiting=%v, want waiting=true", vehicle.Waiting)
	}
	if snapshot.Intersections[0].QueueEW != 1 {
		t.Fatalf("got intersection queueEW=%d, want queueEW=1", snapshot.Intersections[0].QueueEW)
	}
}

func TestVehicleStopsAtSecondIntersection(t *testing.T) {
	engine := NewEngine()
	engine.state.Intersections[0].Phase = model.PhaseEW
	for i := 0; i < 76; i++ {
		engine.step()
	}
	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) == 0 || snapshot.Vehicles[0].ID != 1 {
		t.Fatal("the first vehicle should still be on the map")
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.X != 65 {
		t.Fatalf("got vehicle at X=%v, want X=65", vehicle.X)
	}
	if snapshot.Intersections[1].QueueEW != 1 {
		t.Fatalf("got intersection queueEW=%d, want queueEW=1", snapshot.Intersections[1].QueueEW)
	}
	if snapshot.Intersections[0].QueueEW != 0 {
		t.Fatalf("got intersection queueEW=%d, want queueEW=0", snapshot.Intersections[0].QueueEW)
	}
}
func TestVerticalVehicleStopsAtRedLight(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles[0] = model.Vehicle{ID: 1, X: 22, Y: 15, Direction: model.South}
	engine.state.Intersections[0].Phase = model.PhaseEW
	engine.step()
	snapshot := engine.Snapshot()
	if snapshot.Vehicles[0].Y != 15 {
		t.Fatalf("got vehicle at Y=%v, want Y=15", snapshot.Vehicles[0].Y)
	}
	if !snapshot.Vehicles[0].Waiting {
		t.Fatalf("got vehicle waiting=%v, want waiting=true", snapshot.Vehicles[0].Waiting)
	}
	if snapshot.Statistics.TotalQueue != 1 {
		t.Fatalf("got total queue=%d, want 1", snapshot.Statistics.TotalQueue)
	}
	engine.state.Intersections[0].Phase = model.PhaseNS
	engine.step()
	snapshot = engine.Snapshot()
	if snapshot.Vehicles[0].Y != 16 {
		t.Fatalf("got vehicle at Y=%v, want Y=16", snapshot.Vehicles[0].Y)
	}
	if snapshot.Statistics.TotalQueue != 0 {
		t.Fatalf("got total queue=%d, want 0", snapshot.Statistics.TotalQueue)
	}
	engine.step()
	snapshot = engine.Snapshot()
	if snapshot.Vehicles[0].Y != 17 {
		t.Fatalf("got vehicle at Y=%v, want Y=17", snapshot.Vehicles[0].Y)
	}

}

func TestVehicleCompletesRouteAfterLeavingMap(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles[0] = model.Vehicle{ID: 1, X: 99, Y: 28, Direction: model.East}
	engine.step()
	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("got vehicles=%d, want 1", len(snapshot.Vehicles))
	}
	if snapshot.Vehicles[0].X != 100 {
		t.Fatalf("got vehicle at X=%v, want X=100", snapshot.Vehicles[0].X)
	}
	if snapshot.Statistics.TotalCompleted != 0 {
		t.Fatalf("got completed vehicles=%d, want 0", snapshot.Statistics.TotalCompleted)
	}
	if snapshot.Statistics.TotalSpawned != 1 {
		t.Fatalf("got spawned vehicles=%d, want 1", snapshot.Statistics.TotalSpawned)
	}
	if snapshot.Statistics.ActiveVehicles != 1 {
		t.Fatalf("got active vehicles=%d, want 1", snapshot.Statistics.ActiveVehicles)
	}
	engine.step()
	snapshot = engine.Snapshot()
	if len(snapshot.Vehicles) != 0 {
		t.Fatalf("got %d vehicles, want 0", len(snapshot.Vehicles))
	}
	if snapshot.Statistics.TotalCompleted != 1 {
		t.Fatalf("got completed vehicles=%d, want 1", snapshot.Statistics.TotalCompleted)
	}
	if snapshot.Statistics.TotalSpawned != 1 {
		t.Fatalf("got spawned vehicles=%d, want 1", snapshot.Statistics.TotalSpawned)
	}
	if snapshot.Statistics.ActiveVehicles != 0 {
		t.Fatalf("got active vehicles=%d, want 0", snapshot.Statistics.ActiveVehicles)
	}
}

func TestVehicleSpawnsFromNextEntryPoint(t *testing.T) {
	engine := NewEngine()
	for range spawnIntervalTicks {
		engine.step()
	}
	snapshot := engine.Snapshot()
	if snapshot.Statistics.TotalSpawned != 1 {
		t.Fatalf("got total spawned=%d, want 1", snapshot.Statistics.TotalSpawned)
	}
	engine.step()
	snapshot = engine.Snapshot()
	if snapshot.Statistics.TotalSpawned != 2 {
		t.Fatalf("got total spawned=%d, want 2", snapshot.Statistics.TotalSpawned)
	}
	if len(snapshot.Vehicles) != 2 {
		t.Fatalf("got %d vehicles, want 2", len(snapshot.Vehicles))
	}
	if snapshot.Vehicles[1].ID != 2 {
		t.Fatalf("got vehicle ID=%d, want 2", snapshot.Vehicles[1].ID)
	}
	if snapshot.Vehicles[1].X != 100 {
		t.Fatalf("got vehicle X=%v, want 100", snapshot.Vehicles[1].X)
	}
	if snapshot.Vehicles[1].Y != 72 {
		t.Fatalf("got vehicle Y=%v, want 72", snapshot.Vehicles[1].Y)
	}
	if snapshot.Vehicles[1].Direction != model.West {
		t.Fatalf("got vehicle direction=%v, want west", snapshot.Vehicles[1].Direction)
	}
}

func TestQueueCountsVehiclesBehindRedLight(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
		{ID: 2, X: 8, Y: 28, Direction: model.East},
		{ID: 3, X: 1, Y: 28, Direction: model.East},
	}
	engine.step()
	snapshot := engine.Snapshot()
	if snapshot.Intersections[0].QueueEW != 3 {
		t.Fatalf("got queueEW=%d, want 3", snapshot.Intersections[0].QueueEW)
	}
	if snapshot.Intersections[0].QueueNS != 0 {
		t.Fatalf("got queueNS=%d, want 0", snapshot.Intersections[0].QueueNS)
	}

}

func TestEngineStartAndStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		engine := NewEngine()
		engine.Start(context.Background())
		defer engine.Stop()

		time.Sleep(500 * time.Millisecond)
		synctest.Wait()
		if tick := engine.Snapshot().Tick; tick != 4 {
			t.Fatalf("tick after 500 ms = %d, want 4", tick)
		}

		engine.Stop()
		stoppedTick := engine.Snapshot().Tick
		time.Sleep(500 * time.Millisecond)
		synctest.Wait()
		if tick := engine.Snapshot().Tick; tick != stoppedTick {
			t.Fatalf("tick after Stop = %d, want %d", tick, stoppedTick)
		}
	})
}

func TestEnginePauseAndResume(t *testing.T) {
	engine := NewEngine()

	if err := engine.SetPaused(true); err != nil {
		t.Fatalf("pause engine: %v", err)
	}
	engine.step()
	paused := engine.Snapshot()
	if !paused.Paused || paused.Tick != 0 {
		t.Fatalf("after paused step: paused=%t, tick=%d, want paused=true, tick=0", paused.Paused, paused.Tick)
	}

	if err := engine.SetPaused(false); err != nil {
		t.Fatalf("resume engine: %v", err)
	}
	engine.step()
	resumed := engine.Snapshot()
	if resumed.Paused || resumed.Tick != 1 {
		t.Fatalf("after resumed step: paused=%t, tick=%d, want paused=false, tick=1", resumed.Paused, resumed.Tick)
	}
}

func TestEngineSetDensity(t *testing.T) {
	engine := NewEngine()

	if err := engine.SetDensity(5); err != nil {
		t.Fatalf("SetDensity(5): %v", err)
	}
	if density := engine.Snapshot().Density; density != 5 {
		t.Fatalf("density after SetDensity(5) = %d, want 5", density)
	}

	for _, invalid := range []int{0, 6} {
		if err := engine.SetDensity(invalid); err == nil {
			t.Fatalf("SetDensity(%d) returned nil, want an error", invalid)
		}
		if density := engine.Snapshot().Density; density != 5 {
			t.Fatalf("density after SetDensity(%d) = %d, want 5", invalid, density)
		}
	}
}

func TestEngineSetMode(t *testing.T) {
	engine := NewEngine()

	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatalf("SetMode(adaptive): %v", err)
	}
	if mode := engine.Snapshot().Mode; mode != model.ModeAdaptive {
		t.Fatalf("mode after SetMode(adaptive) = %q, want adaptive", mode)
	}

	if err := engine.SetMode(model.Mode("unknown")); err == nil {
		t.Fatal("SetMode(unknown) returned nil, want an error")
	}
	if mode := engine.Snapshot().Mode; mode != model.ModeAdaptive {
		t.Fatalf("mode after SetMode(unknown) = %q, want adaptive", mode)
	}

	if err := engine.SetMode(model.ModeFixed); err != nil {
		t.Fatalf("SetMode(fixed): %v", err)
	}
	if mode := engine.Snapshot().Mode; mode != model.ModeFixed {
		t.Fatalf("mode after SetMode(fixed) = %q, want fixed", mode)
	}
}

func TestEngineReset(t *testing.T) {
	engine := NewEngine()
	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatal(err)
	}
	if err := engine.SetDensity(5); err != nil {
		t.Fatal(err)
	}
	engine.step()
	if err := engine.SetPaused(true); err != nil {
		t.Fatal(err)
	}

	engine.state.Intersections[0].Phase = model.PhaseEW
	engine.state.Intersections[0].PhaseElapsed = 8
	engine.state.Intersections[0].QueueEW = 2
	engine.state.Intersections[0].QueueNS = 1
	engine.state.Statistics.TotalCompleted = 1
	engine.state.Statistics.TotalQueue = 3
	engine.state.Statistics.AverageWaitingTime = 2.5
	engine.completedWaitingTime = 2.5

	const seed int64 = 123
	if err := engine.Reset(seed); err != nil {
		t.Fatalf("reset engine: %v", err)
	}

	snapshot := engine.Snapshot()
	if snapshot.Mode != model.ModeAdaptive || snapshot.Density != 5 || !snapshot.Paused {
		t.Fatalf("settings after reset: mode=%q, density=%d, paused=%t", snapshot.Mode, snapshot.Density, snapshot.Paused)
	}
	if snapshot.Seed != seed || snapshot.Tick != 0 || len(snapshot.Vehicles) != 0 {
		t.Fatalf("after reset: seed=%d, tick=%d, vehicles=%d", snapshot.Seed, snapshot.Tick, len(snapshot.Vehicles))
	}

	initial := NewEngine().Snapshot()
	if len(snapshot.Intersections) != len(initial.Intersections) {
		t.Fatalf("intersections after reset = %d, want %d", len(snapshot.Intersections), len(initial.Intersections))
	}
	for i, intersection := range snapshot.Intersections {
		if intersection != initial.Intersections[i] {
			t.Fatalf("intersection %d after reset: got %+v, want %+v", i, intersection, initial.Intersections[i])
		}
	}
	if snapshot.Statistics != (model.Statistics{}) || engine.completedWaitingTime != 0 {
		t.Fatalf("after reset: statistics=%+v, completed waiting time=%v", snapshot.Statistics, engine.completedWaitingTime)
	}
}

func TestVehicleMovesAfterFixedLightTurnsGreen(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
	}
	engine.state.Intersections[0].Phase = model.PhaseNS
	engine.state.Intersections[0].PhaseElapsed = 19.75

	engine.step()
	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("vehicles on switching tick = %d, want 1", len(snapshot.Vehicles))
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.X != 15 || !vehicle.Waiting {
		t.Fatalf("vehicle on switching tick: X=%v, waiting=%t, want X=15, waiting=true", vehicle.X, vehicle.Waiting)
	}
	intersection := snapshot.Intersections[0]
	if intersection.Phase != model.PhaseEW || intersection.PhaseElapsed != 0 {
		t.Fatalf("intersection on switching tick: phase=%q, elapsed=%v, want phase=%q, elapsed=0", intersection.Phase, intersection.PhaseElapsed, model.PhaseEW)
	}

	engine.step()
	snapshot = engine.Snapshot()
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("vehicles after green = %d, want 1", len(snapshot.Vehicles))
	}
	vehicle = snapshot.Vehicles[0]
	if vehicle.X != 16 || vehicle.Waiting {
		t.Fatalf("vehicle after green: X=%v, waiting=%t, want X=16, waiting=false", vehicle.X, vehicle.Waiting)
	}
}

func TestEngineChoosesPhaseByMode(t *testing.T) {
	for _, test := range []struct {
		mode      model.Mode
		wantPhase model.Phase
	}{
		{mode: model.ModeFixed, wantPhase: model.PhaseNS},
		{mode: model.ModeAdaptive, wantPhase: model.PhaseEW},
	} {
		t.Run(string(test.mode), func(t *testing.T) {
			engine := NewEngine()
			if err := engine.SetMode(test.mode); err != nil {
				t.Fatalf("set mode: %v", err)
			}
			engine.step()
			engine.state.Vehicles = []model.Vehicle{
				{ID: 1, X: 15, Y: 28, Direction: model.East},
				{ID: 2, X: 8, Y: 28, Direction: model.East},
				{ID: 3, X: 1, Y: 28, Direction: model.East},
			}
			engine.state.Intersections[0].Phase = model.PhaseNS
			engine.state.Intersections[0].PhaseElapsed = 5.75

			engine.step()
			intersection := engine.Snapshot().Intersections[0]
			if intersection.QueueEW != 3 || intersection.QueueNS != 0 {
				t.Fatalf("queues: EW=%d, NS=%d, want EW=3, NS=0", intersection.QueueEW, intersection.QueueNS)
			}
			if intersection.Phase != test.wantPhase {
				t.Fatalf("phase = %q, want %q", intersection.Phase, test.wantPhase)
			}
		})
	}
}

func TestEngineReportsAverageWaitingTimeAfterTrip(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles[0] = model.Vehicle{ID: 1, X: 65, Y: 28, Direction: model.East}
	engine.state.Intersections[1].Phase = model.PhaseNS
	engine.state.Intersections[1].PhaseElapsed = 19.75

	engine.step()
	waiting := engine.Snapshot()
	if len(waiting.Vehicles) != 1 {
		t.Fatalf("vehicles on red = %d, want 1", len(waiting.Vehicles))
	}
	vehicle := waiting.Vehicles[0]
	if vehicle.X != 65 || !vehicle.Waiting || vehicle.WaitingTime != 0.25 {
		t.Fatalf("vehicle on red: X=%v, waiting=%t, waiting time=%v, want X=65, waiting=true, waiting time=0.25", vehicle.X, vehicle.Waiting, vehicle.WaitingTime)
	}
	if waiting.Statistics.TotalCompleted != 0 || waiting.Statistics.AverageWaitingTime != 0 {
		t.Fatalf("before trip completion: completed=%d, average=%v, want completed=0, average=0", waiting.Statistics.TotalCompleted, waiting.Statistics.AverageWaitingTime)
	}

	for range 36 {
		engine.step()
	}
	finished := engine.Snapshot()
	if finished.Statistics.TotalCompleted != 1 || finished.Statistics.AverageWaitingTime != 0.25 {
		t.Fatalf("after trip completion: completed=%d, average=%v, want completed=1, average=0.25", finished.Statistics.TotalCompleted, finished.Statistics.AverageWaitingTime)
	}
}

func TestVehicleWaitsOnGreenUntilIntersectionClears(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()
	engine.state.Intersections[0].Phase = model.PhaseEW
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
		{ID: 2, X: 28, Y: 27, Direction: model.North},
	}

	engine.step()
	vehicles := engine.Snapshot().Vehicles
	if len(vehicles) != 2 {
		t.Fatalf("vehicle count after first tick = %d, want 2", len(vehicles))
	}
	if approaching := vehicles[0]; approaching.X != 15 || !approaching.Waiting {
		t.Fatalf("approaching vehicle: X=%v, waiting=%t, want X=15, waiting=true", approaching.X, approaching.Waiting)
	}
	if crossing := vehicles[1]; crossing.Y != 26 || crossing.Waiting {
		t.Fatalf("crossing vehicle: Y=%v, waiting=%t, want Y=26, waiting=false", crossing.Y, crossing.Waiting)
	}

	for range 16 {
		engine.step()
		vehicles = engine.Snapshot().Vehicles
		if len(vehicles) != 2 {
			t.Fatalf("vehicle count while clearing intersection = %d, want 2", len(vehicles))
		}
		if approaching := vehicles[0]; approaching.X > 15 && !approaching.Waiting {
			return
		}
	}
	t.Fatal("approaching vehicle did not resume after the intersection cleared")
}

func TestOppositeVehiclesCrossOnSeparateLanes(t *testing.T) {
	for _, scenario := range []struct {
		name  string
		westX float64
		steps int
	}{
		{name: "both at stop lines", westX: 35, steps: 4},
		{name: "westbound vehicle already inside", westX: 31, steps: 3},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			engine := NewEngine()
			engine.step()
			engine.state.Intersections[0].Phase = model.PhaseEW
			engine.state.Vehicles = []model.Vehicle{
				{ID: 1, X: 15, Y: 28, Direction: model.East},
				{ID: 2, X: scenario.westX, Y: 22, Direction: model.West},
			}

			for step := 1; step <= scenario.steps; step++ {
				engine.step()
				vehicles := engine.Snapshot().Vehicles
				if len(vehicles) != 2 {
					t.Fatalf("step %d: vehicle count = %d, want 2", step, len(vehicles))
				}
				east, west := vehicles[0], vehicles[1]
				if wantX := 15 + float64(step); east.X != wantX || east.Y != 28 || east.Waiting {
					t.Fatalf("step %d: eastbound vehicle at (%v, %v), waiting=%t, want (%v, 28), waiting=false", step, east.X, east.Y, east.Waiting, wantX)
				}
				if wantX := scenario.westX - float64(step); west.X != wantX || west.Y != 22 || west.Waiting {
					t.Fatalf("step %d: westbound vehicle at (%v, %v), waiting=%t, want (%v, 22), waiting=false", step, west.X, west.Y, west.Waiting, wantX)
				}
			}
		})
	}
}

func TestVehicleClearsIntersectionAfterPhaseChange(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()
	engine.state.Intersections[0].Phase = model.PhaseEW
	engine.state.Intersections[0].PhaseElapsed = 19.75
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 23, Y: 28, Direction: model.East},
		{ID: 2, X: 28, Y: 35, Direction: model.North},
	}

	engine.step()
	snapshot := engine.Snapshot()
	if phase := snapshot.Intersections[0].Phase; phase != model.PhaseNS {
		t.Fatalf("phase after switching tick = %q, want %q", phase, model.PhaseNS)
	}
	if len(snapshot.Vehicles) != 2 {
		t.Fatalf("vehicle count after switching tick = %d, want 2", len(snapshot.Vehicles))
	}
	if east, north := snapshot.Vehicles[0], snapshot.Vehicles[1]; east.X != 24 || east.Waiting || north.Y != 35 || !north.Waiting {
		t.Fatalf("on switching tick: east X=%v waiting=%t, north Y=%v waiting=%t, want east X=24 moving and north Y=35 waiting", east.X, east.Waiting, north.Y, north.Waiting)
	}

	for wantEastX := 25.0; wantEastX <= 34; wantEastX++ {
		engine.step()
		vehicles := engine.Snapshot().Vehicles
		if len(vehicles) != 2 {
			t.Fatalf("vehicle count while clearing intersection = %d, want 2", len(vehicles))
		}
		east, north := vehicles[0], vehicles[1]
		if east.X != wantEastX || east.Waiting {
			t.Fatalf("east after red: X=%v waiting=%t, want X=%v moving", east.X, east.Waiting, wantEastX)
		}
		if north.Y != 35 || !north.Waiting {
			t.Fatalf("north while east reaches X=%v: Y=%v waiting=%t, want Y=35 waiting", east.X, north.Y, north.Waiting)
		}
	}

	engine.step()
	vehicles := engine.Snapshot().Vehicles
	if len(vehicles) != 2 {
		t.Fatalf("vehicle count after clearing intersection = %d, want 2", len(vehicles))
	}
	if north := vehicles[1]; north.Y != 34 || north.Waiting {
		t.Fatalf("north after intersection clears: Y=%v waiting=%t, want Y=34 moving", north.Y, north.Waiting)
	}
}

func TestModeComparisonAndResetReproducibility(t *testing.T) {
	const (
		seed    = 42
		density = 3
		ticks   = 2000
	)

	run := func(engine *Engine) model.Snapshot {
		t.Helper()
		if err := engine.Reset(seed); err != nil {
			t.Fatalf("reset simulation: %v", err)
		}
		for range ticks {
			engine.step()
		}
		return engine.Snapshot()
	}

	for _, mode := range []model.Mode{model.ModeFixed, model.ModeAdaptive} {
		engine := NewEngine()
		if err := engine.SetMode(mode); err != nil {
			t.Fatalf("set mode %q: %v", mode, err)
		}
		if err := engine.SetDensity(density); err != nil {
			t.Fatalf("set density %d: %v", density, err)
		}

		snapshot := run(engine)
		if snapshot.Mode != mode || snapshot.Seed != seed || snapshot.Density != density || snapshot.Tick != ticks {
			t.Fatalf("run parameters: mode=%q seed=%d density=%d ticks=%d, want mode=%q seed=%d density=%d ticks=%d",
				snapshot.Mode, snapshot.Seed, snapshot.Density, snapshot.Tick, mode, seed, density, ticks)
		}
		stats := snapshot.Statistics
		if stats.TotalCompleted == 0 {
			t.Fatalf("%s: no completed trips after %d ticks", mode, ticks)
		}
		t.Logf("%s: completed=%d, average wait=%.4f s", mode, stats.TotalCompleted, stats.AverageWaitingTime)

		repeated := run(engine)
		if !reflect.DeepEqual(repeated, snapshot) {
			t.Fatalf("%s: snapshot changed after reset with the same settings", mode)
		}
	}
}

func TestVehicleWaitsOnGreenForStoppedVehicle(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()
	engine.state.Intersections[0].Phase = model.PhaseEW
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
		{ID: 2, X: 28, Y: 26, Direction: model.North},
		{ID: 3, X: 28, Y: 19, Direction: model.North},
	}

	engine.step()
	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) != 3 {
		t.Fatalf("vehicle count after first tick = %d, want 3", len(snapshot.Vehicles))
	}
	if snapshot.Intersections[0].Phase != model.PhaseEW {
		t.Fatal("eastbound vehicle must have a green light")
	}
	if stopped := snapshot.Vehicles[1]; stopped.Y != 26 || !stopped.Waiting {
		t.Fatalf("vehicle inside intersection: Y=%v waiting=%t, want Y=26 waiting=true", stopped.Y, stopped.Waiting)
	}
	if approaching := snapshot.Vehicles[0]; approaching.X != 15 || !approaching.Waiting {
		t.Fatalf("approaching vehicle on green: X=%v waiting=%t, want X=15 waiting=true", approaching.X, approaching.Waiting)
	}

	for range 16 {
		engine.step()
		snapshot = engine.Snapshot()
		if len(snapshot.Vehicles) != 3 {
			t.Fatalf("vehicle count while clearing intersection = %d, want 3", len(snapshot.Vehicles))
		}
		if approaching := snapshot.Vehicles[0]; approaching.X > 15 && !approaching.Waiting {
			stopped := snapshot.Vehicles[1]
			if stopped.Y+vehicleHalfLength > snapshot.Intersections[0].Y-intersectionHalfSize {
				t.Fatalf("eastbound vehicle moved before the northbound vehicle cleared the intersection: north Y=%v", stopped.Y)
			}
			return
		}
	}
	t.Fatal("eastbound vehicle did not resume after the intersection cleared")
}

func TestSpawnWaitsForSpaceAtEntry(t *testing.T) {
	engine := NewEngine()
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 100, Y: 72, Direction: model.West},
	}
	engine.state.Statistics.TotalSpawned = 1

	engine.step()
	snapshot := engine.Snapshot()
	if snapshot.Statistics.TotalSpawned != 1 || len(snapshot.Vehicles) != 1 {
		t.Fatalf("occupied entry: spawned=%d, active=%d, want 1 of each",
			snapshot.Statistics.TotalSpawned, len(snapshot.Vehicles))
	}

	for range spawnIntervalTicks {
		engine.step()
	}
	snapshot = engine.Snapshot()
	if snapshot.Statistics.TotalSpawned != 2 || len(snapshot.Vehicles) != 2 {
		t.Fatalf("clear entry: spawned=%d, active=%d, want 2 of each",
			snapshot.Statistics.TotalSpawned, len(snapshot.Vehicles))
	}
	if vehicle := snapshot.Vehicles[1]; vehicle.ID != 2 || vehicle.X != 100 || vehicle.Y != 72 {
		t.Fatalf("new vehicle = %+v, want ID=2 at (100, 72)", vehicle)
	}
}

func TestFixedModeBuildsQueueAtHighDensity(t *testing.T) {
	engine := NewEngine()
	if err := engine.SetDensity(5); err != nil {
		t.Fatal(err)
	}

	largestQueue := 0
	for range 80 {
		engine.step()
		largestQueue = max(largestQueue, engine.Snapshot().Statistics.TotalQueue)
	}
	if largestQueue < 2 {
		t.Fatalf("largest total queue in the first 80 ticks = %d, want at least 2", largestQueue)
	}
	t.Logf("largest total queue in the first 80 ticks = %d", largestQueue)
}

func TestVehicleMarkersStaySeparatedAtHighDensity(t *testing.T) {
	for _, mode := range []model.Mode{model.ModeFixed, model.ModeAdaptive} {
		t.Run(string(mode), func(t *testing.T) {
			engine := NewEngine()
			engine.state.Mode = mode
			engine.state.Density = 5
			for range 500 {
				engine.step()
				snapshot := engine.Snapshot()
				for i, vehicle := range snapshot.Vehicles {
					for _, other := range snapshot.Vehicles[i+1:] {
						if math.Abs(vehicle.X-other.X) < 6 && math.Abs(vehicle.Y-other.Y) < 4 {
							t.Fatalf("tick %d: vehicles %d and %d overlap at (%v, %v) and (%v, %v)",
								snapshot.Tick, vehicle.ID, other.ID, vehicle.X, vehicle.Y, other.X, other.Y)
						}
					}
				}
			}
		})
	}
}
