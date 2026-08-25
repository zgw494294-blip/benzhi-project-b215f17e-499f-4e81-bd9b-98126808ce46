package check

import (
	"deepdeploy/internal/domain"
	"sort"
	"strings"
)

func Normalize(c *domain.Config) *domain.Config {
	if c == nil {
		return nil
	}
	out := c.Clone()
	for i := range out.Sensors {
		out.Sensors[i].ID = strings.TrimSpace(out.Sensors[i].ID)
		out.Sensors[i].Model = strings.TrimSpace(out.Sensors[i].Model)
		out.Sensors[i].Bus = strings.TrimSpace(out.Sensors[i].Bus)
	}
	sort.Slice(out.Sensors, func(i, j int) bool { return out.Sensors[i].ID < out.Sensors[j].ID })
	return out
}
func RuleCatalog() map[string]string {
	return map[string]string{"CFG-001": "传感器完整性", "CFG-002": "字段完整性", "BUS-001": "总线兼容性", "COMP-001": "采样率兼容性", "ENV-001": "布放深度", "ENV-002": "压力边界", "ENV-003": "深度边界", "FW-001": "固件版本"}
}
func Explain(code string) string { return RuleCatalog()[code] }
