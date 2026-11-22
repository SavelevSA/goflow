package sim

import (
	"goflow/internal/model"
	"testing"
)

func TestModeComparison(t *testing.T) {
	tests := []struct {
		name    string
		density int
		ticks   int
		seed    int64
	}{
		{name: "High density", density: 5, ticks: 2000, seed: 2},
		{name: "Low density", density: 2, ticks: 2000, seed: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			density := test.density
			ticks := test.ticks
			seed := test.seed

			//               cрез элементов типа
			for _, mode := range []model.Mode{model.ModeFixed, model.ModeAdaptive} {
				engine := NewEngine()

				if err := engine.SetMode(mode); err != nil {
					t.Fatalf("Failed to set mode: %v", err)
				}
				if err := engine.SetDensity(density); err != nil {
					t.Fatalf("Failed to set density: %v", err)
				}

				if err := engine.Reset(seed); err != nil {
					t.Fatalf("Failed to reset engine: %v", err)
				}

				for range ticks {
					engine.step()
				}

				snapshot1 := engine.Snapshot()

				if snapshot1.Statistics.TotalCompleted == 0 {
					t.Errorf("No vehicles completed for mode %s", mode)
				}
				if snapshot1.Tick != ticks {
					t.Errorf("Expected tick %d, got %d for mode %s", ticks, snapshot1.Tick, mode)
				}
				t.Logf("Mode: %s, Total Completed: %d, Average Waiting Time: %.1f seconds, queue: %d, Active Vehicles: %d",
					mode,
					snapshot1.Statistics.TotalCompleted,
					snapshot1.Statistics.AverageWaitingTime,
					snapshot1.Statistics.TotalQueue,
					snapshot1.Statistics.ActiveVehicles,
				)

				//cбрасываем сим
				if err := engine.Reset(seed); err != nil {
					t.Fatalf("Failed to reset engine: %v", err)
				}

				for range ticks {
					engine.step()
				}
				snapshot2 := engine.Snapshot()

				if snapshot2.Statistics.TotalCompleted != snapshot1.Statistics.TotalCompleted {
					t.Errorf("Total completed mismatch for mode %s: %d vs %d", mode, snapshot2.Statistics.TotalCompleted, snapshot1.Statistics.TotalCompleted)
				}
				if snapshot2.Statistics.AverageWaitingTime != snapshot1.Statistics.AverageWaitingTime {
					t.Errorf("Average waiting time mismatch for mode %s: %.2f vs %.2f sec", mode, snapshot2.Statistics.AverageWaitingTime, snapshot1.Statistics.AverageWaitingTime)
				}
			}
		})
	}
}
