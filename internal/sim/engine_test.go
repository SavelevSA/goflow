package sim

import (
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
