package persistence

import (
	"encoding/json"

	"deepdeploy/internal/domain"
)

func Replay(events []domain.Event) map[string]int {
	out := map[string]int{}
	for _, e := range events {
		out[e.Type]++
	}
	return out
}

// ReplayTask reconstructs a task from its event stream. It is used during
// restart recovery when a snapshot file is unavailable but the event ledger is
// intact. Events are applied in sequence through the domain methods so that the
// resulting task carries the same status, version and associations as before
// the restart.
func ReplayTask(events []domain.Event) (*domain.Task, error) {
	if len(events) == 0 {
		return nil, domain.ErrNotFound
	}
	var task *domain.Task
	for _, e := range events {
		switch e.Type {
		case "task_created":
			d := toStringMap(e.Data)
			id := firstNonEmpty(d["id"], e.TaskID)
			t, err := domain.NewTask(id, d["missionName"], d["seaArea"], d["deploymentWindow"], d["owner"], e.At)
			if err != nil {
				return nil, err
			}
			t.IdempotencyKey = d["idempotencyKey"]
			t.CreateFingerprint = d["createFingerprint"]
			task = t
		case "config_created":
			if task == nil {
				continue
			}
			cfg := decodeConfig(e.Data)
			if cfg == nil {
				continue
			}
			_ = task.AddConfig(cfg, task.Version, e.At)
		case "validated":
			if task == nil {
				continue
			}
			d := toStringMap(e.Data)
			configID := d["configID"]
			if configID == "" {
				continue
			}
			report := decodeReport(e.Data)
			findings := []domain.Finding{}
			if report != nil {
				findings = report.Report.Findings
			}
			_, _ = task.ApplyFindings(configID, findings, e.At)
			if task.ValidatedReports == nil {
				task.ValidatedReports = map[string]domain.ValidationRecord{}
			}
			if c, ok := task.Configs[configID]; ok {
				task.ValidatedReports[configID] = domain.ValidationRecord{ConfigHash: c.ContentHash, Findings: findings}
			}
		case "evidence_submitted":
			if task == nil {
				continue
			}
			d := toStringMap(e.Data)
			riskID := d["riskID"]
			if riskID == "" {
				continue
			}
			_ = task.AddEvidence(riskID, d["mitigation"], d["evidence"], task.Version, e.At)
		case "reviewed":
			if task == nil {
				continue
			}
			d := toStringMap(e.Data)
			riskID := d["riskID"]
			decision := d["decision"]
			if decision == "" {
				// Legacy events stored the decision as a bare string.
				if s, ok := e.Data.(string); ok {
					decision = s
				}
			}
			if riskID != "" {
				_ = task.ReviewRisk(riskID, decision, d["reviewer"], d["comment"], task.Version, e.At)
			} else {
				_ = task.Review(decision, d["reviewer"], d["comment"], task.Version, e.At)
			}
		case "permit_issued":
			if task == nil {
				continue
			}
			p := decodePermit(e.Data)
			if p == nil {
				continue
			}
			_, _ = task.FreezeAndPermit(p.ID, p.SerialNumber, p.IssuedBy, p.AuditDigest, task.Version, p.IssuedAt)
		case "permit_revoked":
			if task == nil {
				continue
			}
			_ = task.RevokePermit(e.At)
		}
	}
	if task == nil {
		return nil, domain.ErrNotFound
	}
	return task, nil
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func toStringMap(v any) map[string]string {
	out := map[string]string{}
	b, err := json.Marshal(v)
	if err != nil {
		return out
	}
	_ = json.Unmarshal(b, &out)
	return out
}

func decodeConfig(v any) *domain.Config {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var c domain.Config
	if json.Unmarshal(b, &c) != nil {
		return nil
	}
	return &c
}

type reportData struct {
	ConfigID string      `json:"configID"`
	Report   checkReport `json:"report"`
}

type checkReport struct {
	Findings []domain.Finding `json:"Findings"`
	Passed   bool             `json:"Passed"`
}

func decodeReport(v any) *reportData {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var r reportData
	if json.Unmarshal(b, &r) != nil {
		return nil
	}
	return &r
}

func decodePermit(v any) *domain.Permit {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var p domain.Permit
	if json.Unmarshal(b, &p) != nil {
		return nil
	}
	return &p
}
