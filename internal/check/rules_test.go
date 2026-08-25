package check

import (
	"deepdeploy/internal/domain"
	"testing"
)

func TestValidateFindsBoundaryIssues(t *testing.T) {
	c := &domain.Config{Sensors: []domain.Sensor{{ID: "s", Model: "m", Bus: "b", SampleRate: 10}}, MountingParameters: map[string]float64{"depth_m": 20}, EnvironmentLimits: map[string]float64{"max_depth_m": 10}}
	r := Run(c)
	if r.Passed || len(r.Findings) == 0 {
		t.Fatal("expected deterministic finding")
	}
}
