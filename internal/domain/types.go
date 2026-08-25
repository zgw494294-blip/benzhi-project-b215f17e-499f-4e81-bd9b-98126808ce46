package domain

import "time"

type TaskStatus string

const (
	StatusDraft         TaskStatus = "draft"
	StatusValidated     TaskStatus = "validated"
	StatusRemediation   TaskStatus = "remediation"
	StatusPendingReview TaskStatus = "pending_review"
	StatusApproved      TaskStatus = "approved"
	StatusReleased      TaskStatus = "released"
)

type Sensor struct {
	ID         string `json:"id"`
	Model      string `json:"model"`
	Bus        string `json:"bus"`
	SampleRate int    `json:"sampleRate"`
}
type Config struct {
	ID                 string             `json:"id"`
	TaskID             string             `json:"taskID"`
	Revision           int                `json:"revision"`
	Sensors            []Sensor           `json:"sensors"`
	FirmwareSet        map[string]string  `json:"firmwareSet"`
	MountingParameters map[string]float64 `json:"mountingParameters"`
	EnvironmentLimits  map[string]float64 `json:"environmentLimits"`
	SubmittedBy        string             `json:"submittedBy"`
	SubmittedAt        time.Time          `json:"submittedAt"`
	ContentHash        string             `json:"contentHash"`
	Frozen             bool               `json:"frozen"`
	Diff               *ConfigDiff        `json:"diff,omitempty"`
	Active             bool               `json:"active,omitempty"`
	RiskIDs            []string           `json:"riskIDs,omitempty"`
}
type ConfigDiff struct {
	SensorsAdded      []string `json:"sensorsAdded,omitempty"`
	SensorsRemoved    []string `json:"sensorsRemoved,omitempty"`
	FirmwareChanged   []string `json:"firmwareChanged,omitempty"`
	ParametersChanged []string `json:"parametersChanged,omitempty"`
}
type ValidationRecord struct {
	ConfigHash string    `json:"configHash"`
	Digest     string    `json:"digest"`
	Findings   []Finding `json:"findings"`
}
type Risk struct {
	ID             string    `json:"id"`
	TaskID         string    `json:"taskID"`
	ConfigID       string    `json:"configID"`
	RuleCode       string    `json:"ruleCode"`
	Severity       string    `json:"severity"`
	Description    string    `json:"description"`
	Mitigation     string    `json:"mitigation,omitempty"`
	Evidence       string    `json:"evidence,omitempty"`
	ReviewDecision string    `json:"reviewDecision,omitempty"`
	ReviewedBy     string    `json:"reviewedBy,omitempty"`
	ReviewComment  string    `json:"reviewComment,omitempty"`
	ReviewedAt     time.Time `json:"reviewedAt,omitempty"`
	Active         bool      `json:"active"`
}
type Permit struct {
	ID           string     `json:"id"`
	TaskID       string     `json:"taskID"`
	ConfigID     string     `json:"configID"`
	SerialNumber string     `json:"serialNumber"`
	IssuedBy     string     `json:"issuedBy"`
	ConfigHash   string     `json:"configHash"`
	AuditDigest  string     `json:"auditDigest"`
	IssuedAt     time.Time  `json:"issuedAt"`
	RevokedAt    *time.Time `json:"revokedAt,omitempty"`
	PermitDigest string     `json:"permitDigest,omitempty"`
}
type Task struct {
	ID                string                      `json:"id"`
	MissionName       string                      `json:"missionName"`
	SeaArea           string                      `json:"seaArea"`
	DeploymentWindow  string                      `json:"deploymentWindow"`
	Owner             string                      `json:"owner"`
	Status            TaskStatus                  `json:"status"`
	Version           int                         `json:"version"`
	ActiveConfigID    string                      `json:"activeConfigID,omitempty"`
	CreatedAt         time.Time                   `json:"createdAt"`
	UpdatedAt         time.Time                   `json:"updatedAt"`
	Configs           map[string]*Config          `json:"configs"`
	Risks             map[string]*Risk            `json:"risks"`
	Permit            *Permit                     `json:"permit,omitempty"`
	ValidatedReports  map[string]ValidationRecord `json:"validatedReports,omitempty"`
	PermitHistory     []*Permit                   `json:"permitHistory,omitempty"`
	IdempotencyKey    string                      `json:"idempotencyKey,omitempty"`
	CreateFingerprint string                      `json:"createFingerprint,omitempty"`
}
