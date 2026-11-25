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

type Turn int

const (
	TurnLeft     Turn = -1
	TurnStraight Turn = 0
	TurnRight    Turn = 1
)

type Phase string

const (
	PhaseNS Phase = "ns"
	PhaseEW Phase = "ew"
)

type Vehicle struct {
	ID            int       `json:"id"`
	X             float64   `json:"x"`
	Y             float64   `json:"y"`
	Length        float64   `json:"length"`
	Direction     Direction `json:"direction"`
	Waiting       bool      `json:"waiting"`
	WaitingTime   float64   `json:"waitingTime"`
	NextTurn      Turn      `json:"-"`
	TurnedAt      int       `json:"-"`
	CompletedTurn Turn      `json:"-"`
	TurnAngle     float64   `json:"turnAngle,omitempty"`
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

type ModeStatistics struct {
	CompletedCrossings int     `json:"completedCrossings"`
	WaitingSeconds     float64 `json:"waitingSeconds"`
	ActiveTicks        int     `json:"activeTicks"`
	WaitingVehicles    int     `json:"waitingVehicles"`
	AverageWaitingTime float64 `json:"averageWaitingTime"`
}

type ModeStats struct {
	Fixed    ModeStatistics `json:"fixed"`
	Adaptive ModeStatistics `json:"adaptive"`
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
	ModeStats     ModeStats      `json:"modeStats"`
}
