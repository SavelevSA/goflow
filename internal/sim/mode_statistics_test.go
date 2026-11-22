package sim

import (
	"goflow/internal/model"
	"testing"
)

func TestModeSwitchCrossingVehicle(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1

	engine.step()

	engine.state.Intersections[0].Phase = model.PhaseEW

	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 23, Y: 28, Direction: model.East},
		{ID: 2, X: 28, Y: 35, Direction: model.North},
	}

	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatalf("Failed to set mode: %v", err)
	}

	record, exists := engine.crossings[1]

	if !exists {
		t.Error("Crossing record for vehicle 1 not found")
	} else if record.Mode != model.ModeFixed {
		t.Errorf("Expected fixed mode, got %s", record.Mode)
	}

	_, exists = engine.crossings[2]
	if exists {
		t.Error("Crossing record for vehicle 2 should not exist")
	}

	engine.step()

	snapshot := engine.Snapshot()
	if len(snapshot.Vehicles) != 2 {
		t.Fatalf("Expected 2 vehicles, got %v", len(snapshot.Vehicles))
	}

	east := snapshot.Vehicles[0]
	north := snapshot.Vehicles[1]

	if east.X != 24 || east.Y != 28 || east.Waiting {
		t.Error("East vehicle has incorrect position or waiting state")
	}
	if north.X != 28 || north.Y != 35 || !north.Waiting {
		t.Error("North vehicle has incorrect position or waiting state")
	}
}

func TestModeStatsActiveTicks(t *testing.T) {
	engine := NewEngine()

	for range 4 {
		engine.step()
	}

	// переключение
	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatalf("Failed to set adaptive mode: %v", err)
	}

	for range 3 {
		engine.step()
	}

	if err := engine.SetMode(model.ModeFixed); err != nil {
		t.Fatalf("Failed to set fixed mode: %v", err)
	}

	for range 2 {
		engine.step()
	}

	snapshot := engine.Snapshot()
	if snapshot.ModeStats.Fixed.ActiveTicks != 6 {
		t.Errorf("Expected 6 active ticks for fixed mode, got %d", snapshot.ModeStats.Fixed.ActiveTicks)
	}
	if snapshot.ModeStats.Adaptive.ActiveTicks != 3 {
		t.Errorf("Expected 3 active ticks for adaptive mode, got %d", snapshot.ModeStats.Adaptive.ActiveTicks)
	}

}

func TestModeStatsWaitingSeconds(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()

	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
	}

	engine.state.Intersections[0].Phase = model.PhaseNS

	for range 2 {
		engine.step()
	}

	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatalf("Failed to set adaptive mode: %v", err)
	}

	for range 3 {
		engine.step()
	}

	if err := engine.SetMode(model.ModeFixed); err != nil {
		t.Fatalf("Failed to set fixed mode: %v", err)
	}

	engine.step()

	snapshot := engine.Snapshot()
	if snapshot.ModeStats.Fixed.WaitingSeconds != 0.75 {
		t.Errorf("Expected 0.75 waiting second for fixed mode, got %.2f", snapshot.ModeStats.Fixed.WaitingSeconds)
	}
	if snapshot.ModeStats.Adaptive.WaitingSeconds != 0.75 {
		t.Errorf("Expected 0.75 waiting seconds for adaptive mode, got %.2f", snapshot.ModeStats.Adaptive.WaitingSeconds)
	}
	if snapshot.ModeStats.Fixed.WaitingVehicles != 1 {
		t.Errorf("Expected 1 waiting vehicle for fixed mode, got %d", snapshot.ModeStats.Fixed.WaitingVehicles)
	}
	if snapshot.ModeStats.Adaptive.WaitingVehicles != 1 {
		t.Errorf("Expected 1 waiting vehicle for adaptive mode, got %d", snapshot.ModeStats.Adaptive.WaitingVehicles)
	}
	if snapshot.ModeStats.Fixed.AverageWaitingTime != 0.75 {
		t.Errorf("Expected 0.75 average waiting time for fixed mode, got %.2f", snapshot.ModeStats.Fixed.AverageWaitingTime)
	}
	if snapshot.ModeStats.Adaptive.AverageWaitingTime != 0.75 {
		t.Errorf("Expected 0.75 average waiting time for adaptive mode, got %.2f", snapshot.ModeStats.Adaptive.AverageWaitingTime)
	}
}

func TestModeStatsCrossingOwnership(t *testing.T) {
	engine := NewEngine()

	engine.state.Density = 1
	engine.step()

	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 23, Y: 28, Direction: model.East},
	}

	engine.state.Intersections[0].Phase = model.PhaseEW

	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatalf("Failed to set adaptive mode: %v", err)
	}

	for range 12 {
		engine.step()
	}

	snapshot := engine.Snapshot()

	if len(snapshot.Vehicles) != 1 {
		t.Fatalf("Expected 1 vehicle, got %d", len(snapshot.Vehicles))
	}

	if snapshot.Vehicles[0].X != 35 {
		t.Errorf("Expected vehicle X=35, got %v", snapshot.Vehicles[0].X)
	}

	if snapshot.Mode != model.ModeAdaptive {
		t.Errorf("Expected mode to be adaptive, got %s", snapshot.Mode)
	}

	if snapshot.ModeStats.Fixed.CompletedCrossings != 1 {
		t.Errorf("Expected 1 completed crossings for fixed mode, got %d", snapshot.ModeStats.Fixed.CompletedCrossings)
	}
	if snapshot.ModeStats.Adaptive.CompletedCrossings != 0 {
		t.Errorf("Expected 0 completed crossings for adaptive mode, got %d", snapshot.ModeStats.Adaptive.CompletedCrossings)
	}
}

func TestOccupiesIntersection(t *testing.T) {
	intersection := model.Intersection{
		ID: 1,
		X:  25,
		Y:  25,
	}

	north := model.Vehicle{
		ID:        1,
		X:         28,
		Y:         23,
		Direction: model.North,
	}

	south := model.Vehicle{
		ID:        2,
		X:         22,
		Y:         23,
		Direction: model.South,
	}

	if !occupiesIntersection(north, intersection) {
		t.Error("North vehicle should occupy intersection")
	}
	if !occupiesIntersection(south, intersection) {
		t.Error("South vehicle should occupy intersection")
	}
}

func TestModeStatsNormalCrossing(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()

	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
	}

	engine.state.Intersections[0].Phase = model.PhaseEW

	for range 20 {
		engine.step()
	}

	snapshot := engine.Snapshot()

	if snapshot.ModeStats.Fixed.CompletedCrossings != 1 {
		t.Errorf("Expected 1 completed crossing for fixed mode, got %d", snapshot.ModeStats.Fixed.CompletedCrossings)
	}
	if snapshot.ModeStats.Adaptive.CompletedCrossings != 0 {
		t.Errorf("Expected 0 completed crossings for adaptive mode, got %d", snapshot.ModeStats.Adaptive.CompletedCrossings)
	}
}

func TestModeStatsTwoWaitingVehicles(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()

	engine.state.Intersections[0].Phase = model.PhaseNS
	engine.state.Intersections[2].Phase = model.PhaseNS

	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
	}

	// Первая
	for range 2 {
		engine.step()
	}

	// Добавляем вторую
	engine.state.Vehicles = append(engine.state.Vehicles,
		model.Vehicle{ID: 2, X: 15, Y: 78, Direction: model.East},
	)

	// Обе машины ждут
	for range 2 {
		engine.step()
	}

	snapshot := engine.Snapshot()

	if snapshot.ModeStats.Fixed.WaitingVehicles != 2 {
		t.Errorf("Expected 2 waiting vehicles for fixed mode, got %d", snapshot.ModeStats.Fixed.WaitingVehicles)
	}
	if snapshot.ModeStats.Fixed.WaitingSeconds != 1.5 {
		t.Errorf("Expected 1.5 waiting seconds for fixed mode, got %.2f",
			snapshot.ModeStats.Fixed.WaitingSeconds)
	}

	if snapshot.ModeStats.Fixed.AverageWaitingTime != 0.75 {
		t.Errorf("Expected 0.75 average waiting time for fixed mode, got %.2f",
			snapshot.ModeStats.Fixed.AverageWaitingTime)
	}

}

func TestModeStatsReset(t *testing.T) {
	engine := NewEngine()
	engine.state.Density = 1
	engine.step()

	// Создаём две машины: ожидающую и пересекающую перекрёсток
	engine.state.Vehicles = []model.Vehicle{
		{ID: 1, X: 15, Y: 28, Direction: model.East},
		{ID: 2, X: 23, Y: 28, Direction: model.East},
	}

	engine.state.Intersections[0].Phase = model.PhaseNS

	for range 2 {
		engine.step()
	}

	if err := engine.SetMode(model.ModeAdaptive); err != nil {
		t.Fatalf("Failed to set adaptive mode: %v", err)
	}

	for range 2 {
		engine.step()
	}

	// Проверяем, что до сброса статистика действительно накопилась
	before := engine.Snapshot()

	if before.ModeStats.Fixed.ActiveTicks == 0 ||
		before.ModeStats.Adaptive.ActiveTicks == 0 {
		t.Fatal("Both modes must have accumulated statistics before reset")
	}

	if len(engine.crossings) == 0 {
		t.Fatal("Expected active crossing records before reset")
	}

	if len(engine.waitingSeen[model.ModeFixed]) == 0 {
		t.Fatal("Expected waiting history for fixed mode before reset")
	}

	//  сбрасываем
	if err := engine.Reset(123); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	snapshot := engine.Snapshot()

	if snapshot.ModeStats != (model.ModeStats{}) {
		t.Errorf("Expected empty mode statistics, got %+v",
			snapshot.ModeStats)
	}

	if snapshot.Mode != model.ModeAdaptive {
		t.Errorf("Expected adaptive mode after reset, got %s",
			snapshot.Mode)
	}

	if snapshot.Seed != 123 {
		t.Errorf("Expected seed 123 after reset, got %d",
			snapshot.Seed)
	}

	if len(engine.crossings) != 0 {
		t.Errorf("Expected empty crossings after reset, got %d",
			len(engine.crossings))
	}

	if len(engine.waitingSeen[model.ModeFixed]) != 0 {
		t.Errorf("Expected empty fixed waiting history, got %d",
			len(engine.waitingSeen[model.ModeFixed]))
	}

	if len(engine.waitingSeen[model.ModeAdaptive]) != 0 {
		t.Errorf("Expected empty adaptive waiting history, got %d",
			len(engine.waitingSeen[model.ModeAdaptive]))
	}

	if snapshot.Tick != 0 {
		t.Errorf("Expected tick 0 after reset, got %d",
			snapshot.Tick)
	}
}
