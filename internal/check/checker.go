package check

import (
	"deepdeploy/internal/domain"
	"sort"
)

func Run(c *domain.Config) Report {
	r := Validate(c)
	r.Findings = append(r.Findings, Compatibility(c)...)
	r.Findings = append(r.Findings, Environment(c)...)
	sort.SliceStable(r.Findings, func(i, j int) bool {
		if r.Findings[i].RuleCode == r.Findings[j].RuleCode {
			return r.Findings[i].Description < r.Findings[j].Description
		}
		return r.Findings[i].RuleCode < r.Findings[j].RuleCode
	})
	r.Passed = len(r.Findings) == 0
	return r
}
