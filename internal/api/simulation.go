package api

import (
	"context"

	"goflow/internal/model"
)

// Simulation is the server's contract with the engine. Snapshot returns a
// detached copy. Commands return after the engine has applied the change.
type Simulation interface {
	Snapshot() model.Snapshot
	SetMode(model.Mode) error
	SetDensity(int) error
	SetPaused(bool) error
	Reset(seed int64) error
	Start(context.Context)
	Stop()
}
