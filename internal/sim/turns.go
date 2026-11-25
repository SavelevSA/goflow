package sim

import "goflow/internal/model"

func turnDirection(direction model.Direction, turn model.Turn) model.Direction {
	directions := [...]model.Direction{model.North, model.East, model.South, model.West}
	for i, current := range directions {
		if current == direction {
			return directions[(i+int(turn)+len(directions))%len(directions)]
		}
	}
	return direction
}

func turnPosition(vehicle model.Vehicle, intersection model.Intersection) model.Vehicle {
	vehicle.Direction = turnDirection(vehicle.Direction, vehicle.NextTurn)
	switch vehicle.Direction {
	case model.North:
		vehicle.X = intersection.X + laneOffset
	case model.South:
		vehicle.X = intersection.X - laneOffset
	case model.East:
		vehicle.Y = intersection.Y + laneOffset
	case model.West:
		vehicle.Y = intersection.Y - laneOffset
	}
	return vehicle
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

		turned := turnPosition(vehicle, intersection)
		distance := (turned.X-vehicle.X)*decision.dx + (turned.Y-vehicle.Y)*decision.dy
		if distance >= 0 && distance <= vehicleStep {
			decision.dx = (turned.X - vehicle.X) / vehicleStep
			decision.dy = (turned.Y - vehicle.Y) / vehicleStep
			decision.newDirection = turned.Direction
			decision.turnAt = intersection.ID
			return
		}
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
				!canEnterLane(turnPosition(vehicle, intersection), vehicles))) {
				decision.waitingAt = i
				continue
			}
			turning = turning || plannedTurn
		}
	}
}
