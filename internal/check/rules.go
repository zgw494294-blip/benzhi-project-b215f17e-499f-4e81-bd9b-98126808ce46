package check

import (
	"fmt"
	"sort"

	"deepdeploy/internal/domain"
)

type Report struct {
	Findings []domain.Finding
	Passed   bool
}

func Validate(c *domain.Config) Report {
	findings := []domain.Finding{}
	if c == nil || len(c.Sensors) == 0 {
		findings = append(findings, domain.Finding{RuleCode: "CFG-001", Severity: "critical", Description: "未登记传感器", Mitigation: "至少登记一个传感器"})
		return Report{Findings: findings}
	}
	if c.MountingParameters["depth_m"] <= 0 {
		findings = append(findings, domain.Finding{RuleCode: "ENV-001", Severity: "high", Description: "缺少有效布放深度", Mitigation: "设置正数 depth_m"})
	}
	if c.EnvironmentLimits["max_pressure_bar"] <= 0 {
		findings = append(findings, domain.Finding{RuleCode: "ENV-002", Severity: "medium", Description: "缺少最大压力边界", Mitigation: "设置 max_pressure_bar"})
	}
	buses := map[string]string{}
	for _, s := range c.Sensors {
		if s.ID == "" || s.Model == "" || s.Bus == "" {
			findings = append(findings, domain.Finding{RuleCode: "CFG-002", Severity: "high", Description: "传感器字段不完整", Mitigation: "补齐 ID、Model 和 Bus"})
		}
		if prev, ok := buses[s.Bus]; ok && prev != s.ID {
			findings = append(findings, domain.Finding{RuleCode: "BUS-001", Severity: "high", Description: fmt.Sprintf("总线 %s 存在冲突", s.Bus), Mitigation: "为传感器分配独立总线地址"})
		}
		buses[s.Bus] = s.ID
	}
	for id, fw := range c.FirmwareSet {
		if fw == "" {
			findings = append(findings, domain.Finding{RuleCode: "FW-001", Severity: "medium", Description: "固件版本为空: " + id, Mitigation: "登记固件版本"})
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].RuleCode < findings[j].RuleCode })
	return Report{Findings: findings, Passed: len(findings) == 0}
}
