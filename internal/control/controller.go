package control

import "goflow/internal/model"

type Controller interface {
	NextPhase(model.Intersection) model.Phase
}

type FixedController struct {
	Duration float64
}

// * указатель
func (fc *FixedController) NextPhase(intersection model.Intersection) model.Phase {
	if intersection.PhaseElapsed < fc.Duration {
		return intersection.Phase
	}

	if intersection.Phase == model.PhaseNS {
		return model.PhaseEW
	}
	return model.PhaseNS

}
