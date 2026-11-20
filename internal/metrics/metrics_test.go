package metrics

import (
	"testing"
)

func TestCalculate(t *testing.T) {
	spawned := 6
	completed := 4

	c := Counters{
		TotalSpawned:         spawned,
		TotalCompleted:       completed,
		ActiveVehicles:       spawned - completed,
		TotalQueue:           2,
		CompletedWaitingTime: 32,
	}
	result := Calculate(c)
	if result.AverageWaitingTime != 8 {
		t.Errorf("Expected average waiting time to be 8, got %f", result.AverageWaitingTime)
	}
	if result.TotalSpawned != 6 {
		t.Errorf("Expected total spawned to be 6, got %d", result.TotalSpawned)
	}
	if result.TotalCompleted != 4 {
		t.Errorf("Expected total completed to be 4, got %d", result.TotalCompleted)
	}
	if result.ActiveVehicles != 2 {
		t.Errorf("Expected active vehicles to be 2, got %d", result.ActiveVehicles)
	}
	if result.TotalQueue != 2 {
		t.Errorf("Expected total queue to be 2, got %d", result.TotalQueue)
	}
}

func TestCalculateZeroCompleted(t *testing.T) {
	spawned := 2
	completed := 0

	c := Counters{
		TotalSpawned:         spawned,
		TotalCompleted:       completed,
		ActiveVehicles:       spawned - completed,
		TotalQueue:           2,
		CompletedWaitingTime: 0,
	}
	result := Calculate(c)
	if result.AverageWaitingTime != 0 {
		t.Errorf("Expected average waiting time to be 0, got %f", result.AverageWaitingTime)
	}
}
