package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

func NewTask(id, mission, sea, window, owner string, now time.Time) (*Task, error) {
	if strings.TrimSpace(id) == "" || strings.TrimSpace(mission) == "" || strings.TrimSpace(sea) == "" || strings.TrimSpace(window) == "" || strings.TrimSpace(owner) == "" {
		return nil, ErrInvalid
	}
	window = strings.TrimSpace(window)
	if strings.ContainsAny(id+mission+sea+window+owner, "\r\n\x00") {
		return nil, ErrInvalid
	}
	return &Task{ID: strings.TrimSpace(id), MissionName: strings.TrimSpace(mission), SeaArea: strings.TrimSpace(sea), DeploymentWindow: window, Owner: strings.TrimSpace(owner), Status: StatusDraft, Version: 1, CreatedAt: now, UpdatedAt: now, Configs: map[string]*Config{}, Risks: map[string]*Risk{}, ValidatedReports: map[string]ValidationRecord{}}, nil
}

func (t *Task) AddConfig(c *Config, expected int, now time.Time) error {
	if t.Version != expected {
		return ErrConflict
	}
	if t.Status == StatusReleased || t.Status == StatusApproved {
		return ErrInvalidState
	}
	if c == nil || len(c.Sensors) == 0 || c.TaskID != t.ID {
		return ErrInvalid
	}
	if _, exists := t.Configs[c.ID]; exists {
		return ErrConflict
	}
	seen := map[string]bool{}
	for _, s := range c.Sensors {
		if strings.TrimSpace(s.ID) == "" || strings.TrimSpace(s.Bus) == "" || s.SampleRate < 0 || seen[s.ID] {
			return ErrInvalid
		}
		seen[s.ID] = true
	}
	for _, values := range []map[string]float64{c.MountingParameters, c.EnvironmentLimits} {
		for _, v := range values {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return ErrInvalid
			}
		}
	}
	if d := c.MountingParameters["depth_m"]; d < 0 || (t.DeploymentWindow == "" && d != 0) {
		return ErrInvalid
	}
	providedHash := c.ContentHash
	for _, old := range t.Configs {
		if old.Revision >= c.Revision {
			c.Revision = old.Revision + 1
		}
	}
	if c.Revision <= 0 {
		c.Revision = len(t.Configs) + 1
	}
	c.ContentHash = HashConfig(c)
	if providedHash != "" && providedHash != c.ContentHash {
		return ErrConflict
	}
	if prev := t.Configs[t.ActiveConfigID]; prev != nil {
		prev.Active = false
		c.Diff = DiffConfig(prev, c)
		for _, risk := range t.Risks {
			if risk.ConfigID == prev.ID {
				risk.Active = false
			}
		}
	}
	c.Active = true
	t.Configs[c.ID] = c
	t.ActiveConfigID = c.ID
	t.Status = StatusDraft
	t.Version++
	t.UpdatedAt = now
	return nil
}

func DiffConfig(a, b *Config) *ConfigDiff {
	d := &ConfigDiff{}
	am, bm := map[string]bool{}, map[string]bool{}
	as := map[string]Sensor{}
	for _, s := range a.Sensors {
		am[s.ID] = true
		as[s.ID] = s
	}
	for _, s := range b.Sensors {
		bm[s.ID] = true
		if !am[s.ID] {
			d.SensorsAdded = append(d.SensorsAdded, s.ID)
		} else if as[s.ID] != s {
			d.ParametersChanged = append(d.ParametersChanged, "sensor:"+s.ID)
		}
	}
	for id := range am {
		if !bm[id] {
			d.SensorsRemoved = append(d.SensorsRemoved, id)
		}
	}
	for k, v := range b.FirmwareSet {
		if a.FirmwareSet[k] != v {
			d.FirmwareChanged = append(d.FirmwareChanged, k)
		}
	}
	for k := range a.FirmwareSet {
		if _, ok := b.FirmwareSet[k]; !ok {
			d.FirmwareChanged = append(d.FirmwareChanged, k)
		}
	}
	for k, v := range b.MountingParameters {
		if a.MountingParameters[k] != v {
			d.ParametersChanged = append(d.ParametersChanged, "mounting:"+k)
		}
	}
	for k := range a.MountingParameters {
		if _, ok := b.MountingParameters[k]; !ok {
			d.ParametersChanged = append(d.ParametersChanged, "mounting:"+k)
		}
	}
	for k, v := range b.EnvironmentLimits {
		if a.EnvironmentLimits[k] != v {
			d.ParametersChanged = append(d.ParametersChanged, "environment:"+k)
		}
	}
	for k := range a.EnvironmentLimits {
		if _, ok := b.EnvironmentLimits[k]; !ok {
			d.ParametersChanged = append(d.ParametersChanged, "environment:"+k)
		}
	}
	sort.Strings(d.SensorsAdded)
	sort.Strings(d.SensorsRemoved)
	sort.Strings(d.FirmwareChanged)
	sort.Strings(d.ParametersChanged)
	return d
}

func HashConfig(c *Config) string {
	// Canonicalize map ordering so retries and process restarts produce one hash.
	type canonical struct {
		TaskID      string   `json:"taskID"`
		Revision    int      `json:"revision"`
		Sensors     []Sensor `json:"sensors"`
		Firmware    []string `json:"firmware"`
		Mounting    []string `json:"mounting"`
		Environment []string `json:"environment"`
	}
	keys := func(m map[string]float64) []string {
		ks := make([]string, 0, len(m))
		for k := range m {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		out := make([]string, 0, len(ks))
		for _, k := range ks {
			out = append(out, k+"="+formatFloat(m[k]))
		}
		return out
	}
	fwKeys := make([]string, 0, len(c.FirmwareSet))
	for k := range c.FirmwareSet {
		fwKeys = append(fwKeys, k)
	}
	sort.Strings(fwKeys)
	fw := make([]string, 0, len(fwKeys))
	for _, k := range fwKeys {
		fw = append(fw, k+"="+c.FirmwareSet[k])
	}
	b, _ := json.Marshal(canonical{TaskID: c.TaskID, Revision: c.Revision, Sensors: c.Sensors, Firmware: fw, Mounting: keys(c.MountingParameters), Environment: keys(c.EnvironmentLimits)})
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

func formatFloat(v float64) string { return strconv.FormatFloat(v, 'g', -1, 64) }

func (t *Task) ApplyFindings(configID string, findings []Finding, now time.Time) ([]*Risk, error) {
	c, ok := t.Configs[configID]
	if !ok {
		return nil, ErrNotFound
	}
	if c.Frozen {
		return nil, ErrFrozen
	}
	for _, risk := range t.Risks {
		if risk.ConfigID == configID {
			risk.Active = false
		}
	}
	var out []*Risk
	c.RiskIDs = nil
	for i, f := range findings {
		id := fmt.Sprintf("%s-risk-%d", configID, i+1)
		r := &Risk{ID: id, TaskID: t.ID, ConfigID: configID, RuleCode: f.RuleCode, Severity: f.Severity, Description: f.Description, Mitigation: f.Mitigation, Active: true}
		t.Risks[id] = r
		c.RiskIDs = append(c.RiskIDs, id)
		out = append(out, r)
	}
	if len(out) == 0 {
		t.Status = StatusValidated
	} else {
		t.Status = StatusRemediation
	}
	t.Version++
	t.UpdatedAt = now
	return out, nil
}

func (t *Task) AddEvidence(riskID, mitigation, evidence string, expected int, now time.Time) error {
	if t.Version != expected {
		return ErrConflict
	}
	r, ok := t.Risks[riskID]
	if !ok || !r.Active {
		return ErrNotFound
	}
	if invalidText(mitigation, 1, 4096) || invalidText(evidence, 1, 16384) {
		return ErrInvalid
	}
	if c, ok := t.Configs[r.ConfigID]; ok && c.Frozen {
		return ErrFrozen
	}
	if r.ReviewDecision != "" && r.ReviewDecision != "reject" {
		return ErrInvalidState
	}
	if t.Status == StatusApproved || t.Status == StatusReleased {
		return ErrInvalidState
	}
	if r.ReviewDecision == "reject" {
		r.ReviewDecision, r.ReviewedBy = "", ""
		r.ReviewComment = ""
		r.ReviewedAt = time.Time{}
	}
	r.Mitigation = mitigation
	r.Evidence = evidence
	all := false
	for _, x := range t.Risks {
		if !x.Active {
			continue
		}
		all = true
		if x.Evidence == "" {
			all = false
			break
		}
	}
	if all {
		t.Status = StatusPendingReview
	}
	t.Version++
	t.UpdatedAt = now
	return nil
}

func invalidText(v string, min, max int) bool {
	if len(strings.TrimSpace(v)) < min || len(v) > max {
		return true
	}
	for _, r := range v {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func (t *Task) Review(decision, reviewer, comment string, expected int, now time.Time) error {
	if t.Version != expected {
		return ErrConflict
	}
	if t.Status != StatusPendingReview && t.Status != StatusValidated && t.Status != StatusRemediation {
		return ErrInvalidState
	}
	if decision != "approve" && decision != "reject" {
		return ErrInvalid
	}
	if strings.TrimSpace(reviewer) == "" || (decision == "reject" && strings.TrimSpace(comment) == "") {
		return ErrInvalid
	}
	if decision == "approve" {
		for _, r := range t.Risks {
			if !r.Active {
				continue
			}
			if !r.HasEvidence() || r.ReviewDecision != "approve" {
				return ErrInvalidState
			}
		}
	}
	for _, r := range t.Risks {
		if !r.Active {
			continue
		}
		r.ReviewDecision = decision
		r.ReviewedBy = reviewer
		r.ReviewedAt = now
		r.ReviewComment = comment
		if comment != "" {
			r.Mitigation += " | " + comment
		}
	}
	if decision == "approve" {
		t.Status = StatusApproved
	} else {
		t.Status = StatusRemediation
	}
	t.Version++
	t.UpdatedAt = now
	return nil
}

func (t *Task) ReviewRisk(riskID, decision, reviewer, comment string, expected int, now time.Time) error {
	if t.Version != expected {
		return ErrConflict
	}
	r, ok := t.Risks[riskID]
	if !ok || !r.Active {
		return ErrNotFound
	}
	if !r.HasEvidence() || strings.TrimSpace(reviewer) == "" || (decision == "reject" && strings.TrimSpace(comment) == "") {
		return ErrInvalid
	}
	if decision != "approve" && decision != "reject" {
		return ErrInvalid
	}
	r.ReviewDecision, r.ReviewedBy, r.ReviewedAt = decision, reviewer, now
	r.ReviewComment = comment
	if decision == "reject" {
		r.ReviewDecision = ""
		r.ReviewedBy = ""
		r.ReviewedAt = time.Time{}
		t.Status = StatusRemediation
	} else {
		all := true
		for _, x := range t.Risks {
			if !x.Active {
				continue
			}
			if !x.HasEvidence() || x.ReviewDecision != "approve" {
				all = false
			}
		}
		if all {
			t.Status = StatusApproved
		} else {
			t.Status = StatusPendingReview
		}
	}
	t.Version++
	t.UpdatedAt = now
	return nil
}

func (t *Task) FreezeAndPermit(permitID, serial, issuer, digest string, expected int, now time.Time) (*Permit, error) {
	if t.Version != expected {
		return nil, ErrConflict
	}
	if t.Status != StatusApproved || strings.TrimSpace(issuer) == "" || strings.TrimSpace(serial) == "" {
		return nil, ErrInvalidState
	}
	if t.Permit != nil && t.Permit.RevokedAt == nil {
		return nil, ErrInvalidState
	}
	c, ok := t.Configs[t.ActiveConfigID]
	if !ok {
		return nil, ErrNotFound
	}
	if HashConfig(c) != c.ContentHash {
		return nil, ErrConflict
	}
	c.Frozen = true
	p := &Permit{ID: permitID, TaskID: t.ID, ConfigID: c.ID, SerialNumber: serial, IssuedBy: issuer, IssuedAt: now, ConfigHash: c.ContentHash, AuditDigest: digest}
	p.PermitDigest = PermitDigest(p)
	t.Permit = p
	t.PermitHistory = append(t.PermitHistory, p)
	t.Status = StatusReleased
	t.Version++
	t.UpdatedAt = now
	return p, nil
}

type Finding struct{ RuleCode, Severity, Description, Mitigation string }
