package domain

import (
	"encoding/json"
	"time"
)

type TaskSummary struct {
	ID, MissionName, SeaArea, Owner string
	Status                          TaskStatus
	Version                         int
	ConfigID                        string
	RiskCount                       int
	PermitSerial                    string
	UpdatedAt                       time.Time
}

func (t *Task) Summary() TaskSummary {
	p := ""
	if t.Permit != nil {
		p = t.Permit.SerialNumber
	}
	return TaskSummary{ID: t.ID, MissionName: t.MissionName, SeaArea: t.SeaArea, Owner: t.Owner, Status: t.Status, Version: t.Version, ConfigID: t.ActiveConfigID, RiskCount: len(t.Risks), PermitSerial: p, UpdatedAt: t.UpdatedAt}
}
func MarshalSummary(t *Task) ([]byte, error) { return json.Marshal(t.Summary()) }
