package transport

import (
	"deepdeploy/internal/application"
	"deepdeploy/internal/check"
	"deepdeploy/internal/domain"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func (s *Server) tasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var in application.TaskInput
		if decode(r, &in) != nil {
			writeErr(w, domain.ErrInvalid)
			return
		}
		if application.ValidateTaskInput(in) != nil {
			writeErr(w, domain.ErrInvalid)
			return
		}
		t, e := s.app.CreateTaskWithKey(in.ID, in.MissionName, in.SeaArea, in.DeploymentWindow, in.Owner, idempotencyKey(r))
		if e != nil {
			writeErr(w, e)
			return
		}
		writeJSON(w, 201, t)
		return
	}
	writeErr(w, domain.ErrInvalid)
}
func (s *Server) taskSubresource(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		writeErr(w, domain.ErrInvalid)
		return
	}
	taskID := parts[2]
	if len(parts) == 3 && r.Method == "GET" {
		severity := r.URL.Query().Get("severity")
		openOnly := r.URL.Query().Get("openOnly") == "true"
		if severity != "" && severity != "low" && severity != "medium" && severity != "high" && severity != "critical" {
			writeErr(w, domain.ErrInvalid)
			return
		}
		t, e := s.app.SummaryFiltered(taskID, severity, openOnly)
		if e != nil {
			writeErr(w, e)
			return
		}
		writeJSON(w, 200, t)
		return
	}
	if len(parts) == 4 && parts[3] == "configs" && r.Method == "POST" {
		var in application.ConfigInput
		if decode(r, &in) != nil {
			writeErr(w, domain.ErrInvalid)
			return
		}
		c := &domain.Config{ID: fmt.Sprintf("cfg-%d", now().UnixNano()), TaskID: taskID, Sensors: in.Sensors, FirmwareSet: in.FirmwareSet, MountingParameters: in.MountingParameters, EnvironmentLimits: in.EnvironmentLimits, SubmittedBy: in.SubmittedBy, SubmittedAt: now()}
		for _, sensor := range in.Sensors {
			if strings.TrimSpace(sensor.ID) == "" || strings.TrimSpace(sensor.Bus) == "" || sensor.SampleRate <= 0 || strings.TrimSpace(in.FirmwareSet[sensor.ID]) == "" {
				writeErr(w, domain.ErrInvalid)
				return
			}
		}
		t, e := s.app.AddConfig(taskID, c, versionOr(r, in.ExpectedVersion), r.Header.Get("Idempotency-Key"))
		if e != nil {
			writeErr(w, e)
			return
		}
		cfg := t.Configs[t.ActiveConfigID]
		writeJSON(w, 201, map[string]any{"task": t, "taskID": t.ID, "status": t.Status, "version": t.Version, "configID": cfg.ID, "revision": cfg.Revision, "contentHash": cfg.ContentHash})
		return
	}
	if len(parts) >= 5 && parts[3] == "configs" {
		configID := parts[4]
		if len(parts) == 6 && parts[5] == "validate" && r.Method == "POST" {
			var in struct {
				ExpectedVersion int `json:"expectedVersion"`
			}
			_ = decode(r, &in)
			t, risks, e := s.app.Validate(taskID, configID, versionOr(r, in.ExpectedVersion))
			if e != nil {
				writeErr(w, e)
				return
			}
			report := check.Deterministic(t.Configs[configID])
			if cached, ok := t.ValidatedReports[configID]; ok && cached.ConfigHash == t.Configs[configID].ContentHash {
				report = check.Report{Findings: cached.Findings, Passed: len(cached.Findings) == 0}
			}
			writeJSON(w, 200, map[string]any{"task": t, "risks": risks, "report": validationReport(report)})
			return
		}
	}
	if len(parts) == 4 && parts[3] == "review" && r.Method == "POST" {
		var in application.ReviewInput
		if decode(r, &in) != nil {
			writeErr(w, domain.ErrInvalid)
			return
		}
		var t *domain.Task
		var e error
		if in.RiskID != "" {
			t, e = s.app.ReviewRisk(taskID, in.RiskID, in.Decision, in.Reviewer, in.Comment, versionOr(r, in.ExpectedVersion))
		} else {
			t, e = s.app.Review(taskID, in.Decision, in.Reviewer, in.Comment, versionOr(r, in.ExpectedVersion))
		}
		if e != nil {
			writeErr(w, e)
			return
		}
		writeJSON(w, 200, t)
		return
	}
	if len(parts) == 4 && parts[3] == "permit" && r.Method == "POST" {
		var in application.PermitInput
		if decode(r, &in) != nil {
			writeErr(w, domain.ErrInvalid)
			return
		}
		if in.Operation == "revoke" || in.Action == "revoke" || in.Revoke {
			p, e := s.app.RevokePermit(taskID, in.RevokedBy, in.Reason, versionOr(r, in.ExpectedVersion))
			if e != nil {
				writeErr(w, e)
				return
			}
			writeJSON(w, 200, p)
			return
		}
		p, e := s.app.Permit(taskID, in.Issuer, versionOr(r, in.ExpectedVersion))
		if e != nil {
			writeErr(w, e)
			return
		}
		writeJSON(w, 201, p)
		return
	}
	if len(parts) == 4 && parts[3] == "events" && r.Method == "GET" {
		e, err := s.app.Events(taskID)
		if err != nil {
			writeErr(w, err)
			return
		}
		writeJSON(w, 200, e)
		return
	}
	if len(parts) == 6 && parts[3] == "risks" && parts[5] == "evidence" && r.Method == "POST" {
		var in application.EvidenceInput
		if decode(r, &in) != nil {
			writeErr(w, domain.ErrInvalid)
			return
		}
		t, e := s.app.AddEvidence(taskID, parts[4], in.Mitigation, in.Evidence, versionOr(r, in.ExpectedVersion), r.Header.Get("Idempotency-Key"))
		if e != nil {
			writeErr(w, e)
			return
		}
		writeJSON(w, 200, t)
		return
	}
	writeErr(w, domain.ErrInvalid)
}

func validationReport(report check.Report) map[string]any {
	counts := map[string]int{"low": 0, "medium": 0, "high": 0, "critical": 0}
	highest := ""
	rank := map[string]int{"low": 1, "medium": 2, "high": 3, "critical": 4}
	best := 0
	for _, f := range report.Findings {
		counts[f.Severity]++
		if rank[f.Severity] > best {
			best = rank[f.Severity]
			highest = f.Severity
		}
	}
	findings := make([]map[string]any, 0, len(report.Findings))
	for _, f := range report.Findings {
		findings = append(findings, map[string]any{"ruleCode": f.RuleCode, "severity": f.Severity, "description": f.Description, "mitigation": f.Mitigation, "explanation": check.Explain(f.RuleCode)})
	}
	return map[string]any{"passed": report.Passed, "digest": report.Digest(), "reportDigest": report.Digest(), "highestSeverity": highest, "severityCounts": counts, "findings": findings}
}
func version(r *http.Request) int {
	var v int
	fmt.Sscanf(r.Header.Get("X-Expected-Version"), "%d", &v)
	return v
}
func versionOr(r *http.Request, body int) int {
	if v := version(r); v != 0 {
		return v
	}
	return body
}
func now() time.Time { return time.Now() }
