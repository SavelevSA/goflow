package sim

import (
	"math"

	"goflow/internal/model"
)

func turnDirection(direction model.Direction, turn model.Turn) model.Direction {
	directions := [...]model.Direction{model.North, model.East, model.South, model.West}
	for i, current := range directions {
		if current == direction {
			return directions[(i+int(turn)+len(directions))%len(directions)]
		}
	}
	return direction
}

func turnRadius(turn model.Turn) float64 {
	if turn == model.TurnLeft {
		return 2 * laneOffset
	}
	return laneOffset
}

func turnPoint(vehicle model.Vehicle, intersection model.Intersection, angle float64) model.Vehicle {
	outgoing := turnDirection(vehicle.Direction, vehicle.NextTurn)
	dx, dy := directionStep(vehicle.Direction)
	outX, outY := directionStep(outgoing)
	radius := turnRadius(vehicle.NextTurn)

	// The pivot is where the incoming and outgoing lane centers meet.
	pivotX := intersection.X - (dy+outY)*laneOffset
	pivotY := intersection.Y + (dx+outX)*laneOffset
	if angle == 90 {
		vehicle.X = pivotX + outX*radius
		vehicle.Y = pivotY + outY*radius
		vehicle.Direction = outgoing
		vehicle.TurnAngle = 0
		return vehicle
	}

	sin, cos := math.Sincos(angle * math.Pi / 180)
	vehicle.X = pivotX + radius*(dx*(sin-1)+outX*(1-cos))
	vehicle.Y = pivotY + radius*(dy*(sin-1)+outY*(1-cos))
	vehicle.TurnAngle = float64(vehicle.NextTurn) * angle
	return vehicle
}

func advanceTurn(vehicle model.Vehicle, intersection model.Intersection, decision *movementDecision) {
	// The wider left turn takes twice as many ticks as the right turn.
	anglePerTick := 15 * laneOffset / turnRadius(vehicle.NextTurn)
	angle := min(90, math.Abs(vehicle.TurnAngle)+anglePerTick)
	turned := turnPoint(vehicle, intersection, angle)
	decision.dx = (turned.X - vehicle.X) / vehicleStep
	decision.dy = (turned.Y - vehicle.Y) / vehicleStep
	decision.turnAt = intersection.ID
	decision.turnAngle = turned.TurnAngle
	if angle == 90 {
		decision.newDirection = turned.Direction
	}
}

func planIntersectionMove(vehicle model.Vehicle, intersections []model.Intersection, decision *movementDecision) {
	next := vehicle
	next.X += decision.dx * vehicleStep
	next.Y += decision.dy * vehicleStep

	for i, intersection := range intersections {
		if !occupiesIntersection(vehicle, intersection) {
			if occupiesIntersection(next, intersection) {
				decision.enteringAt = i
			}
			continue
		}
		if vehicle.NextTurn == model.TurnStraight {
			continue
		}

		start := turnPoint(vehicle, intersection, 0)
		distance := (start.X-vehicle.X)*decision.dx + (start.Y-vehicle.Y)*decision.dy
		if distance < 0 || distance > vehicleStep {
			continue
		}
		if distance == 0 {
			advanceTurn(vehicle, intersection, decision)
		} else {
			decision.dx = (start.X - vehicle.X) / vehicleStep
			decision.dy = (start.Y - vehicle.Y) / vehicleStep
		}
		return
	}
}

// A turning vehicle reserves the intersection until its rear has cleared it.
func resolveIntersectionEntries(vehicles []model.Vehicle, intersections []model.Intersection, decisions []movementDecision) {
	for i, intersection := range intersections {
		occupied, turning, straightEntering := false, false, false
		for j, vehicle := range vehicles {
			plannedTurn := vehicle.NextTurn != model.TurnStraight
			if occupiesIntersection(vehicle, intersection) {
				occupied = true
				turning = turning || plannedTurn || vehicle.TurnedAt == intersection.ID
			}
			if decisions[j].enteringAt == i && !plannedTurn {
				straightEntering = true
			}
		}

		for j, vehicle := range vehicles {
			decision := &decisions[j]
			if decision.enteringAt != i {
				continue
			}
			plannedTurn := vehicle.NextTurn != model.TurnStraight
			if turning || (plannedTurn && (occupied || straightEntering ||
				!canEnterLane(turnPoint(vehicle, intersection, 90), vehicles))) {
				decision.waitingAt = i
				continue
			}
			turning = turning || plannedTurn
		}
	}
}
