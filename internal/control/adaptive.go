package control

import "goflow/internal/model"

type AdaptiveController struct {
	MinDuration float64
	MaxDuration float64
	QueueGap    int
}

func (ac *AdaptiveController) NextPhase(i model.Intersection) model.Phase {
	if i.PhaseElapsed < ac.MinDuration {
		return i.Phase
	}
	if i.Phase == model.PhaseNS &&
		i.PhaseElapsed >= ac.MaxDuration &&
		i.QueueEW > 0 {
		return model.PhaseEW
	}
	if i.Phase == model.PhaseNS && i.QueueEW >= i.QueueNS+ac.QueueGap {
		return model.PhaseEW
	}
	if i.Phase == model.PhaseEW &&
		i.PhaseElapsed >= ac.MaxDuration &&
		i.QueueNS > 0 {
		return model.PhaseNS
	}
	if i.Phase == model.PhaseEW && i.QueueNS >= i.QueueEW+ac.QueueGap {
		return model.PhaseNS
	}
	if i.Phase == model.PhaseNS &&
		i.QueueNS == 0 &&
		i.QueueEW > 0 {
		return model.PhaseEW
	}
	if i.Phase == model.PhaseEW &&
		i.QueueEW == 0 &&
		i.QueueNS > 0 {
		return model.PhaseNS
	}
	return i.Phase
}
