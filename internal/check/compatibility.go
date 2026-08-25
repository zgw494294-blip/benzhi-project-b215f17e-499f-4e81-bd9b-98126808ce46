package check

import "deepdeploy/internal/domain"

func Compatibility(c *domain.Config) []domain.Finding {
	var out []domain.Finding
	if c == nil {
		return out
	}
	for _, s := range c.Sensors {
		if s.SampleRate < 1 || s.SampleRate > 1000 {
			out = append(out, domain.Finding{RuleCode: "COMP-001", Severity: "medium", Description: "采样率超出支持范围", Mitigation: "将采样率设置在 1-1000Hz"})
		}
	}
	return out
}
