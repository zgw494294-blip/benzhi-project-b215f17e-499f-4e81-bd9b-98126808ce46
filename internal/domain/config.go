package domain

import (
	"sort"
	"strings"
	"time"
)

func NewConfig(id, taskID, submittedBy string, sensors []Sensor, firmware map[string]string, mount, env map[string]float64, now time.Time) (*Config, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(taskID) == "" || strings.TrimSpace(submittedBy) == "" {
		return nil, ErrInvalid
	}
	if len(sensors) == 0 {
		return nil, ErrInvalid
	}
	copySensors := append([]Sensor(nil), sensors...)
	sort.Slice(copySensors, func(i, j int) bool { return copySensors[i].ID < copySensors[j].ID })
	return &Config{ID: id, TaskID: taskID, Sensors: copySensors, FirmwareSet: copyMap(firmware), MountingParameters: copyFloatMap(mount), EnvironmentLimits: copyFloatMap(env), SubmittedBy: submittedBy, SubmittedAt: now}, nil
}
func copyMap(in map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
func copyFloatMap(in map[string]float64) map[string]float64 {
	out := map[string]float64{}
	for k, v := range in {
		out[k] = v
	}
	return out
}
func (c *Config) Sensor(id string) (Sensor, bool) {
	for _, s := range c.Sensors {
		if s.ID == id {
			return s, true
		}
	}
	return Sensor{}, false
}
func (c *Config) IsComplete() bool {
	if len(c.Sensors) == 0 {
		return false
	}
	for _, s := range c.Sensors {
		if s.ID == "" || s.Model == "" || s.Bus == "" {
			return false
		}
	}
	return true
}
func (c *Config) Clone() *Config {
	out := *c
	out.Sensors = append([]Sensor(nil), c.Sensors...)
	out.FirmwareSet = copyMap(c.FirmwareSet)
	out.MountingParameters = copyFloatMap(c.MountingParameters)
	out.EnvironmentLimits = copyFloatMap(c.EnvironmentLimits)
	out.RiskIDs = append([]string(nil), c.RiskIDs...)
	out.Diff = cloneDiff(c.Diff)
	return &out
}

func cloneDiff(d *ConfigDiff) *ConfigDiff {
	if d == nil {
		return nil
	}
	cp := *d
	cp.SensorsAdded = append([]string(nil), d.SensorsAdded...)
	cp.SensorsRemoved = append([]string(nil), d.SensorsRemoved...)
	cp.FirmwareChanged = append([]string(nil), d.FirmwareChanged...)
	cp.ParametersChanged = append([]string(nil), d.ParametersChanged...)
	return &cp
}
