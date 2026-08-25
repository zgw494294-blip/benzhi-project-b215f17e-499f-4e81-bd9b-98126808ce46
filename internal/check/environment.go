package check

import "deepdeploy/internal/domain"

func Environment(c *domain.Config) []domain.Finding {
	var out []domain.Finding
	if c == nil {
		return out
	}
	depth := c.MountingParameters["depth_m"]
	max := c.EnvironmentLimits["max_depth_m"]
	if max > 0 && depth > max {
		out = append(out, domain.Finding{RuleCode: "ENV-003", Severity: "critical", Description: "布放深度超过设备边界", Mitigation: "调整深度或提高设备耐压等级"})
	}
	return out
}
