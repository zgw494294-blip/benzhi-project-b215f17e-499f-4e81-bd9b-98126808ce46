package check

import "deepdeploy/internal/domain"

func Deterministic(c *domain.Config) Report { n := Normalize(c); return Run(n) }
func HasCritical(r Report) bool {
	for _, f := range r.Findings {
		if f.Severity == "critical" {
			return true
		}
	}
	return false
}
