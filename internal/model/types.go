package model

type Mode string

const (
	ModeFixed    Mode = "fixed"
	ModeAdaptive Mode = "adaptive"
)

type Direction string

const (
	North Direction = "north"
	South Direction = "south"
	East  Direction = "east"
	West  Direction = "west"
)

type Phase string

const (
	PhaseNS Phase = "ns"
	PhaseEW Phase = "ew"
)

type Vehicle struct {
	ID          int       `json:"id"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Direction   Direction `json:"direction"`
	Waiting     bool      `json:"waiting"`
	WaitingTime float64   `json:"waitingTime"`
}

type Intersection struct {
	ID           int     `json:"id"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	Phase        Phase   `json:"phase"`
	QueueNS      int     `json:"queueNS"`
	QueueEW      int     `json:"queueEW"`
	PhaseElapsed float64 `json:"phaseElapsed"`
}

type Statistics struct {
	TotalSpawned       int     `json:"totalSpawned"`
	TotalCompleted     int     `json:"totalCompleted"`
	ActiveVehicles     int     `json:"activeVehicles"`
	TotalQueue         int     `json:"totalQueue"`
	AverageWaitingTime float64 `json:"averageWaitingTime"`
}

type Snapshot struct {
	Tick          int            `json:"tick"`
	Mode          Mode           `json:"mode"`
	Density       int            `json:"density"`
	Paused        bool           `json:"paused"`
	Seed          int64          `json:"seed"`
	Vehicles      []Vehicle      `json:"vehicles"`
	Intersections []Intersection `json:"intersections"`
	Statistics    Statistics     `json:"statistics"`
}
