package check

import (
	"crypto/sha256"
	"deepdeploy/internal/domain"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type RuleResult struct {
	Code, Severity, Description string
	Passed                      bool
}

func (r Report) Codes() []string {
	out := []string{}
	for _, f := range r.Findings {
		out = append(out, f.RuleCode)
	}
	sort.Strings(out)
	return out
}
func (r Report) Digest() string {
	b, _ := json.Marshal(r)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func (r Report) HighestSeverity() string {
	rank := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	best := ""
	n := 0
	for _, f := range r.Findings {
		if rank[f.Severity] > n {
			n = rank[f.Severity]
			best = f.Severity
		}
	}
	return best
}
func FindingsBySeverity(fs []domain.Finding, severity string) []domain.Finding {
	out := []domain.Finding{}
	for _, f := range fs {
		if f.Severity == severity {
			out = append(out, f)
		}
	}
	return out
}
