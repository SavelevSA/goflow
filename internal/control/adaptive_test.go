package control

import (
	"goflow/internal/model"
	"testing"
)

func TestAdaptiveController(t *testing.T) {
	controller := AdaptiveController{
		MinDuration: 5,
		MaxDuration: 15,
		QueueGap:    2,
	}

	intersection := model.Intersection{
		Phase:        model.PhaseNS,
		PhaseElapsed: 6,
		QueueNS:      2,
		QueueEW:      5,
	}

	result := controller.NextPhase(intersection)
	if result != model.PhaseEW {
		t.Errorf("Expected %s, got %s", model.PhaseEW, result)
	}

}

func TestAdaptiveScenarios(t *testing.T) {
	controller := AdaptiveController{
		MinDuration: 5,
		MaxDuration: 15,
		QueueGap:    2,
	}

	testingCases := []struct {
		name     string
		input    model.Intersection
		expected model.Phase
	}{
		{name: "Large EW qeue",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: 6,
				QueueNS:      2,
				QueueEW:      100,
			},
			expected: model.PhaseEW,
		},
		{name: "Large NS queue",
			input: model.Intersection{
				Phase:        model.PhaseEW,
				PhaseElapsed: 6,
				QueueNS:      100,
				QueueEW:      2,
			},
			expected: model.PhaseNS,
		},
		{name: "Min duration not reached",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: controller.MinDuration - 1,
				QueueNS:      2,
				QueueEW:      100,
			},
			expected: model.PhaseNS,
		},
		{name: "Max duration exceeded",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: controller.MaxDuration + 1,
				QueueNS:      100,
				QueueEW:      2,
			},
			expected: model.PhaseEW,
		},
		{name: "Zero NS queue",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: 6,
				QueueNS:      0,
				QueueEW:      1,
			},
			expected: model.PhaseEW,
		},
		{name: "Zero EW queue",
			input: model.Intersection{
				Phase:        model.PhaseEW,
				PhaseElapsed: 10,
				QueueNS:      1,
				QueueEW:      0,
			},
			expected: model.PhaseNS,
		},
		{name: "Zero EW and NS queue",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: controller.MaxDuration + 1,
				QueueNS:      0,
				QueueEW:      0,
			},
			expected: model.PhaseNS,
		},
		{name: "Equal queues",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: controller.MaxDuration - 1,
				QueueNS:      5,
				QueueEW:      5,
			},
			expected: model.PhaseNS,
		},
		{name: "Gap not reached",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: controller.MinDuration + 1,
				QueueNS:      5,
				QueueEW:      6,
			},
			expected: model.PhaseNS,
		},
		{name: "Gap is reached",
			input: model.Intersection{
				Phase:        model.PhaseNS,
				PhaseElapsed: controller.MinDuration + 1,
				QueueNS:      5,
				QueueEW:      5 + controller.QueueGap,
			},
			expected: model.PhaseEW,
		},
	}

	for _, tc := range testingCases {
		//               анонимная функция
		t.Run(tc.name, func(t *testing.T) {
			result := controller.NextPhase(tc.input)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}
