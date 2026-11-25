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

func planIntersectionMove(vehicle model.Vehicle, traffic []model.Vehicle, intersections []model.Intersection, decision *movementDecision) {
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
			for j, other := range traffic {
				if other.TurnAngle != 0 && other.TurnedAt == intersection.ID && other.Direction == vehicle.Direction {
					decision.blockedBy = j
					return
				}
			}
			if !canEnterLane(turnPoint(vehicle, intersection, 90), traffic) {
				decision.waitingAt = i
				return
			}
			advanceTurn(vehicle, intersection, decision)
		} else {
			decision.dx = (start.X - vehicle.X) / vehicleStep
			decision.dy = (start.Y - vehicle.Y) / vehicleStep
		}
		return
	}
}

func intersectionManeuver(vehicle model.Vehicle, intersectionID int) (model.Direction, model.Turn) {
	if vehicle.TurnedAt == intersectionID && vehicle.CompletedTurn != model.TurnStraight {
		return turnDirection(vehicle.Direction, -vehicle.CompletedTurn), vehicle.CompletedTurn
	}
	return vehicle.Direction, vehicle.NextTurn
}

func approachLane(vehicle model.Vehicle, intersections []model.Intersection) (laneKey, model.Direction) {
	if vehicle.TurnedAt != 0 {
		for _, intersection := range intersections {
			if intersection.ID != vehicle.TurnedAt || !occupiesIntersection(vehicle, intersection) {
				continue
			}
			direction, _ := intersectionManeuver(vehicle, intersection.ID)
			dx, dy := directionStep(direction)
			if dx != 0 {
				return laneKey{horizontal: true, coordinate: intersection.Y + dx*laneOffset}, direction
			}
			return laneKey{coordinate: intersection.X - dy*laneOffset}, direction
		}
	}
	return laneOf(vehicle), vehicle.Direction
}

func intersectionPathsConflict(first, second model.Vehicle, intersectionID int) bool {
	firstDirection, firstTurn := intersectionManeuver(first, intersectionID)
	secondDirection, secondTurn := intersectionManeuver(second, intersectionID)
	dx, dy := directionStep(firstDirection)
	otherDX, otherDY := directionStep(secondDirection)
	approach := dx*otherDX + dy*otherDY
	if approach == 1 {
		return false
	}

	if firstTurn == model.TurnStraight && secondTurn == model.TurnStraight {
		return approach == 0
	}

	return approach != -1 || firstTurn == model.TurnLeft || secondTurn == model.TurnLeft
}

func resolveIntersectionEntries(vehicles []model.Vehicle, intersections []model.Intersection, decisions []movementDecision) {
	for i, intersection := range intersections {
		var occupants []model.Vehicle
		for _, vehicle := range vehicles {
			if occupiesIntersection(vehicle, intersection) {
				occupants = append(occupants, vehicle)
			}
		}

		// Admit through traffic and right turns before left turns.
		for _, turn := range [...]model.Turn{model.TurnStraight, model.TurnRight, model.TurnLeft} {
			for j, vehicle := range vehicles {
				decision := &decisions[j]
				if decision.enteringAt != i || vehicle.NextTurn != turn {
					continue
				}
				if turn != model.TurnStraight && !canEnterLane(turnPoint(vehicle, intersection, 90), vehicles) {
					decision.waitingAt = i
					continue
				}
				for _, other := range occupants {
					if intersectionPathsConflict(vehicle, other, intersection.ID) {
						decision.waitingAt = i
						break
					}
				}
				if decision.waitingAt == -1 {
					occupants = append(occupants, vehicle)
				}
			}
		}
	}
}
