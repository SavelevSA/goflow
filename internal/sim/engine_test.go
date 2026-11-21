package sim

import (
	"context"
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
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("got %d vehicles, want 1", len(snapshot.Vehicles))
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.X != 24 {
		t.Fatalf("got vehicle at X=%v, want X=24", vehicle.X)
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
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("got %d vehicles, want 1", len(snapshot.Vehicles))
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.X != 74 {
		t.Fatalf("got vehicle at X=%v, want X=74", vehicle.X)
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
	engine.state.Vehicles[0] = model.Vehicle{ID: 1, X: 25, Y: 24, Direction: model.South}
	engine.state.Intersections[0].Phase = model.PhaseEW
	engine.step()
	snapshot := engine.Snapshot()
	if snapshot.Vehicles[0].Y != 24 {
		t.Fatalf("got vehicle at Y=%v, want Y=24", snapshot.Vehicles[0].Y)
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
	if snapshot.Vehicles[0].Y != 25 {
		t.Fatalf("got vehicle at Y=%v, want Y=25", snapshot.Vehicles[0].Y)
	}
	if snapshot.Statistics.TotalQueue != 0 {
		t.Fatalf("got total queue=%d, want 0", snapshot.Statistics.TotalQueue)
	}
	engine.step()
	snapshot = engine.Snapshot()
	if snapshot.Vehicles[0].Y != 26 {
		t.Fatalf("got vehicle at Y=%v, want Y=26", snapshot.Vehicles[0].Y)
	}

}

func TestVehicleCompletesRouteAfterLeavingMap(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles[0] = model.Vehicle{ID: 1, X: 99, Y: 25, Direction: model.East}
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
	for i := 0; i < 80; i++ {
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
	if snapshot.Vehicles[1].Y != 75 {
		t.Fatalf("got vehicle Y=%v, want 75", snapshot.Vehicles[1].Y)
	}
	if snapshot.Vehicles[1].Direction != model.West {
		t.Fatalf("got vehicle direction=%v, want west", snapshot.Vehicles[1].Direction)
	}
}

func TestQueueCountsVehiclesBehindRedLight(t *testing.T) {
	engine := NewEngine()
	engine.step()
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 24, Y: 25, Direction: model.East},
		{ID: 2, X: 23, Y: 25, Direction: model.East},
		{ID: 3, X: 22, Y: 25, Direction: model.East},
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
		if tick := engine.Snapshot().Tick; tick == 0 {
			t.Fatal("engine did not tick after Start")
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
		{ID: 1, X: 24, Y: 25, Direction: model.East},
	}
	engine.state.Intersections[0].Phase = model.PhaseNS
	engine.state.Intersections[0].PhaseElapsed = 19.75

	engine.step()
	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("vehicles on switching tick = %d, want 1", len(snapshot.Vehicles))
	}
	vehicle := snapshot.Vehicles[0]
	if vehicle.X != 24 || !vehicle.Waiting {
		t.Fatalf("vehicle on switching tick: X=%v, waiting=%t, want X=24, waiting=true", vehicle.X, vehicle.Waiting)
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
	if vehicle.X != 25 || vehicle.Waiting {
		t.Fatalf("vehicle after green: X=%v, waiting=%t, want X=25, waiting=false", vehicle.X, vehicle.Waiting)
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
				{ID: 1, X: 24, Y: 25, Direction: model.East},
				{ID: 2, X: 23, Y: 25, Direction: model.East},
				{ID: 3, X: 22, Y: 25, Direction: model.East},
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
	engine.state.Vehicles[0] = model.Vehicle{ID: 1, X: 74, Y: 25, Direction: model.East}
	engine.state.Intersections[1].Phase = model.PhaseNS
	engine.state.Intersections[1].PhaseElapsed = 19.75

	engine.step()
	waiting := engine.Snapshot()
	if len(waiting.Vehicles) != 1 {
		t.Fatalf("vehicles on red = %d, want 1", len(waiting.Vehicles))
	}
	vehicle := waiting.Vehicles[0]
	if vehicle.X != 74 || !vehicle.Waiting || vehicle.WaitingTime != 0.25 {
		t.Fatalf("vehicle on red: X=%v, waiting=%t, waiting time=%v, want X=74, waiting=true, waiting time=0.25", vehicle.X, vehicle.Waiting, vehicle.WaitingTime)
	}
	if waiting.Statistics.TotalCompleted != 0 || waiting.Statistics.AverageWaitingTime != 0 {
		t.Fatalf("before trip completion: completed=%d, average=%v, want completed=0, average=0", waiting.Statistics.TotalCompleted, waiting.Statistics.AverageWaitingTime)
	}

	for range 27 {
		engine.step()
	}
	finished := engine.Snapshot()
	if finished.Statistics.TotalCompleted != 1 || finished.Statistics.AverageWaitingTime != 0.25 {
		t.Fatalf("after trip completion: completed=%d, average=%v, want completed=1, average=0.25", finished.Statistics.TotalCompleted, finished.Statistics.AverageWaitingTime)
	}
}
