package sim

import (
	"goflow/internal/model"
	"testing"
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
