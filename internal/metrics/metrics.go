package metrics

import (
	"goflow/internal/model"
)

type Counters struct {
	TotalSpawned         int
	TotalCompleted       int
	ActiveVehicles       int
	TotalQueue           int
	CompletedWaitingTime float64 // Total waiting time of completed vehicles в секундах
}

func Calculate(c Counters) model.Statistics {
	var stats model.Statistics
	stats.TotalSpawned = c.TotalSpawned
	stats.TotalCompleted = c.TotalCompleted
	stats.ActiveVehicles = c.ActiveVehicles
	stats.TotalQueue = c.TotalQueue
	if c.TotalCompleted > 0 {
		averageWaitingTime := c.CompletedWaitingTime / float64(c.TotalCompleted)
		stats.AverageWaitingTime = averageWaitingTime
	}
	return stats
}
