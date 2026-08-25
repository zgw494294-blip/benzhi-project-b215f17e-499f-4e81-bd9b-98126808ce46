package application

import "deepdeploy/internal/domain"

func RiskSummary(t *domain.Task) map[string]int {
	out := map[string]int{}
	for _, r := range t.Risks {
		out[r.Severity]++
	}
	return out
}
