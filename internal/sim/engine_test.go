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
	engine.state.Intersections[0].Phase = model.PhaseNS
	engine.step()
	snapshot = engine.Snapshot()
	if snapshot.Vehicles[0].Y != 25 {
		t.Fatalf("got vehicle at Y=%v, want Y=25", snapshot.Vehicles[0].Y)
	}
	engine.step()
	snapshot = engine.Snapshot()
	if snapshot.Vehicles[0].Y != 26 {
		t.Fatalf("got vehicle at Y=%v, want Y=26", snapshot.Vehicles[0].Y)
	}

}
