package application

import (
	"deepdeploy/internal/domain"
	"encoding/json"
)

type TaskInput struct {
	ID               string `json:"taskID"`
	MissionName      string `json:"missionName"`
	SeaArea          string `json:"seaArea"`
	DeploymentWindow string `json:"deploymentWindow"`
	Owner            string `json:"owner"`
}

func (t *TaskInput) UnmarshalJSON(b []byte) error {
	var v struct {
		ID               string `json:"id"`
		TaskID           string `json:"taskID"`
		MissionName      string `json:"missionName"`
		SeaArea          string `json:"seaArea"`
		DeploymentWindow string `json:"deploymentWindow"`
		Owner            string `json:"owner"`
	}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	t.ID = v.TaskID
	if t.ID == "" {
		t.ID = v.ID
	}
	t.MissionName = v.MissionName
	t.SeaArea = v.SeaArea
	t.DeploymentWindow = v.DeploymentWindow
	t.Owner = v.Owner
	return nil
}

type ConfigInput struct {
	ExpectedVersion    int                `json:"expectedVersion,omitempty"`
	Sensors            []domain.Sensor    `json:"sensors"`
	FirmwareSet        map[string]string  `json:"firmwareSet"`
	MountingParameters map[string]float64 `json:"mountingParameters"`
	EnvironmentLimits  map[string]float64 `json:"environmentLimits"`
	SubmittedBy        string             `json:"submittedBy"`
}
type EvidenceInput struct {
	ExpectedVersion int    `json:"expectedVersion,omitempty"`
	Mitigation      string `json:"mitigation"`
	Evidence        string `json:"evidence"`
}
type ReviewInput struct {
	ExpectedVersion int    `json:"expectedVersion,omitempty"`
	Decision        string `json:"decision"`
	Reviewer        string `json:"reviewer"`
	Comment         string `json:"comment"`
	RiskID          string `json:"riskID,omitempty"`
}
type PermitInput struct {
	ExpectedVersion int    `json:"expectedVersion,omitempty"`
	Operation       string `json:"operation,omitempty"`
	Action          string `json:"action,omitempty"`
	Revoke          bool   `json:"revoke,omitempty"`
	Issuer          string `json:"issuer,omitempty"`
	RevokedBy       string `json:"revokedBy,omitempty"`
	Reason          string `json:"reason,omitempty"`
}
