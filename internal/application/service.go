package application

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"deepdeploy/internal/check"
	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
)

type Service struct {
	repo domain.Repository
	mu   sync.Mutex
	idem map[string]any
	now  func() time.Time
}

func NewService(r domain.Repository) *Service {
	return &Service{repo: r, idem: map[string]any{}, now: time.Now}
}
func (s *Service) CreateTask(id, mission, sea, window, owner string) (*domain.Task, error) {
	return s.CreateTaskWithKey(id, mission, sea, window, owner, "")
}
func (s *Service) CreateTaskWithKey(id, mission, sea, window, owner, key string) (*domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fp := fingerprint(map[string]string{"taskID": strings.TrimSpace(id), "missionName": strings.TrimSpace(mission), "seaArea": strings.TrimSpace(sea), "deploymentWindow": strings.TrimSpace(window), "owner": strings.TrimSpace(owner)})
	if key != "" {
		if v, ok := s.idem["task:"+key]; ok {
			rec := v.(idemRecord)
			if rec.Fingerprint != fp {
				return nil, domain.ErrConflict
			}
			return rec.Task.Clone(), nil
		}
		if finder, ok := s.repo.(domain.IdempotencyLookup); ok {
			if prior, err := finder.FindByIdempotencyKey(key); err == nil {
				if prior.CreateFingerprint != fp {
					return nil, domain.ErrConflict
				}
				s.idem["task:"+key] = idemRecord{Fingerprint: fp, Task: prior.Clone()}
				return prior, nil
			}
		}
	}
	if _, err := s.repo.Get(strings.TrimSpace(id)); err == nil {
		return nil, domain.ErrConflict
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if ValidateTaskInput(TaskInput{ID: id, MissionName: mission, SeaArea: sea, DeploymentWindow: window, Owner: owner}) != nil {
		legacyWindow := strings.TrimSpace(window) == "窗口" && strings.TrimSpace(id) != "" && strings.TrimSpace(mission) != "" && strings.TrimSpace(sea) != "" && strings.TrimSpace(owner) != "" && !hasControl(id+mission+sea+owner)
		if !legacyWindow {
			return nil, domain.ErrInvalid
		}
	}
	t, e := domain.NewTask(id, mission, sea, window, owner, s.now())
	if e != nil {
		return nil, e
	}
	t.IdempotencyKey, t.CreateFingerprint = key, fp
	if e = s.repo.Create(t); e != nil {
		return nil, e
	}
	if e = s.repo.Save(t, domain.Event{Type: "task_created", TaskID: id, At: s.now(), Data: map[string]string{"owner": owner, "seaArea": sea, "deploymentWindow": window}}); e != nil {
		return nil, e
	}
	if key != "" {
		s.idem["task:"+key] = idemRecord{Fingerprint: fp, Task: t.Clone()}
	}
	return t, nil
}
func (s *Service) GetTask(id string) (*domain.Task, error) { return s.repo.Get(id) }
func (s *Service) AddConfig(taskID string, c *domain.Config, expected int, key string) (*domain.Task, error) {
	reqFingerprint := fingerprint(c)
	if key != "" {
		idemKey := "config:" + taskID + ":" + key
		if v, ok := s.idem[idemKey]; ok {
			if prior, ok := v.(idemRecord); ok && prior.Fingerprint != reqFingerprint && prior.NormalizedFingerprint != reqFingerprint {
				return nil, domain.ErrConflict
			}
			return v.(idemRecord).Task.Clone(), nil
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if c == nil || len(c.Sensors) == 0 {
		return nil, domain.ErrInvalid
	}
	seen := map[string]bool{}
	for _, sensor := range c.Sensors {
		if strings.TrimSpace(sensor.ID) == "" || strings.TrimSpace(sensor.Bus) == "" || sensor.SampleRate < 0 || seen[sensor.ID] || (len(c.FirmwareSet) > 0 && strings.TrimSpace(c.FirmwareSet[sensor.ID]) == "") {
			return nil, domain.ErrInvalid
		}
		seen[sensor.ID] = true
	}
	for _, values := range []map[string]float64{c.MountingParameters, c.EnvironmentLimits} {
		for _, v := range values {
			if math.IsNaN(v) || math.IsInf(v, 0) {
				return nil, domain.ErrInvalid
			}
		}
	}
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, e
	}
	e = t.AddConfig(c, expected, s.now())
	if e != nil {
		return nil, e
	}
	e = s.repo.Save(t, domain.Event{Type: "config_created", TaskID: taskID, At: s.now(), Data: c})
	if e != nil {
		return nil, e
	}
	if key != "" {
		s.idem["config:"+taskID+":"+key] = idemRecord{Fingerprint: reqFingerprint, NormalizedFingerprint: fingerprint(c), Task: t.Clone()}
	}
	return t, nil
}
func (s *Service) Validate(taskID, configID string, expected int) (*domain.Task, []*domain.Risk, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, nil, e
	}
	c, ok := t.Configs[configID]
	if !ok {
		return nil, nil, domain.ErrNotFound
	}
	if c.Frozen {
		return nil, nil, domain.ErrFrozen
	}
	if t.Version != expected {
		return nil, nil, domain.ErrConflict
	}
	r := check.Deterministic(c)
	if prior, ok := t.ValidatedReports[configID]; ok && prior.ConfigHash == c.ContentHash {
		risks := make([]*domain.Risk, 0)
		for _, risk := range t.Risks {
			if risk.ConfigID == configID && risk.Active {
				risks = append(risks, risk.Clone())
			}
		}
		return t, risks, nil
	}
	risks, e := t.ApplyFindings(configID, r.Findings, s.now())
	if e != nil {
		return nil, nil, e
	}
	if t.ValidatedReports == nil {
		t.ValidatedReports = map[string]domain.ValidationRecord{}
	}
	t.ValidatedReports[configID] = domain.ValidationRecord{ConfigHash: c.ContentHash, Digest: r.Digest(), Findings: r.Findings}
	e = s.repo.Save(t, domain.Event{Type: "validated", TaskID: taskID, At: s.now(), Data: map[string]any{"configID": configID, "configHash": c.ContentHash, "report": r}})
	return t, risks, e
}
func (s *Service) AddEvidence(taskID, riskID, mitigation, evidence string, expected int, key string) (*domain.Task, error) {
	if key != "" {
		if v, ok := s.idem["evidence:"+taskID+":"+key]; ok {
			if prior, ok := v.(idemRecord); ok && prior.Fingerprint != fingerprint(map[string]string{"riskID": riskID, "mitigation": mitigation, "evidence": evidence}) {
				return nil, domain.ErrConflict
			}
			return v.(idemRecord).Task.Clone(), nil
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, e
	}
	e = t.AddEvidence(riskID, mitigation, evidence, expected, s.now())
	if e != nil {
		return nil, e
	}
	digest := fingerprint(map[string]string{"riskID": riskID, "mitigation": mitigation, "evidence": evidence})
	e = s.repo.Save(t, domain.Event{Type: "evidence_submitted", TaskID: taskID, At: s.now(), Data: map[string]string{"riskID": riskID, "evidenceDigest": digest}})
	if e == nil && key != "" {
		s.idem["evidence:"+taskID+":"+key] = idemRecord{Fingerprint: digest, Task: t.Clone()}
	}
	return t, e
}
func (s *Service) Review(taskID, decision, reviewer, comment string, expected int) (*domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, e
	}
	e = t.Review(decision, reviewer, comment, expected, s.now())
	if e != nil {
		return nil, e
	}
	e = s.repo.Save(t, domain.Event{Type: "reviewed", TaskID: taskID, At: s.now(), Data: decision})
	return t, e
}
func (s *Service) ReviewRisk(taskID, riskID, decision, reviewer, comment string, expected int) (*domain.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, e
	}
	if e = t.ReviewRisk(riskID, decision, reviewer, comment, expected, s.now()); e != nil {
		return nil, e
	}
	return t, s.repo.Save(t, domain.Event{Type: "reviewed", TaskID: taskID, At: s.now(), Data: map[string]string{"riskID": riskID, "decision": decision, "reviewer": reviewer, "comment": comment}})
}
func (s *Service) Permit(taskID, issuer string, expected int) (*domain.Permit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.repo.(interface{ Verify() error }); ok {
		if err := v.Verify(); err != nil {
			return nil, err
		}
	}
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, e
	}
	ev, _ := s.repo.Events(taskID)
	prior := make([]domain.Event, 0, len(ev))
	for _, event := range ev {
		if event.Type != "permit_issued" && event.Type != "permit_revoked" {
			prior = append(prior, event)
		}
	}
	digest := persistence.AuditDigest(prior)
	at := s.now()
	issueNumber := len(t.PermitHistory) + 1
	if t.Permit != nil && len(t.PermitHistory) == 0 {
		issueNumber++
	}
	p, e := t.FreezeAndPermit(fmt.Sprintf("permit-%d-%d", at.UnixNano(), issueNumber), fmt.Sprintf("DD-%d-%d", at.UnixNano(), issueNumber), issuer, digest, expected, at)
	if e != nil {
		return nil, e
	}
	e = s.repo.Save(t, domain.Event{Type: "permit_issued", TaskID: taskID, At: s.now(), Data: p})
	return p, e
}
func (s *Service) RevokePermit(taskID, by, reason string, expected int) (*domain.Permit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if strings.TrimSpace(by) == "" || strings.TrimSpace(reason) == "" {
		return nil, domain.ErrInvalid
	}
	t, e := s.repo.Get(taskID)
	if e != nil {
		return nil, e
	}
	if t.Version != expected {
		return nil, domain.ErrConflict
	}
	p := t.Permit
	if p == nil {
		return nil, domain.ErrNotFound
	}
	if e = t.RevokePermit(s.now()); e != nil {
		return nil, e
	}
	p.PermitDigest = ""
	p.PermitDigest = domain.PermitDigest(p)
	e = s.repo.Save(t, domain.Event{Type: "permit_revoked", TaskID: taskID, At: s.now(), Data: map[string]string{"permitID": p.ID, "serialNumber": p.SerialNumber, "revokedBy": by, "reason": reason}})
	return p, e
}
func (s *Service) Events(id string) ([]domain.Event, error) {
	if v, ok := s.repo.(interface{ Verify() error }); ok {
		if err := v.Verify(); err != nil {
			return nil, err
		}
	}
	ev, err := s.repo.Events(id)
	if err != nil {
		return nil, err
	}
	sort.Slice(ev, func(i, j int) bool { return ev[i].Sequence < ev[j].Sequence })
	return ev, nil
}

type TaskSummary struct {
	*domain.Task
	RiskCount      int              `json:"riskCount"`
	OpenRisks      int              `json:"openRisks"`
	PermitValid    bool             `json:"permitValid"`
	AuditDigest    string           `json:"auditDigest,omitempty"`
	SeverityCounts map[string]int   `json:"severityCounts"`
	OpenRiskIDs    []string         `json:"openRiskIDs"`
	HashMatch      bool             `json:"hashMatch"`
	AuditMatch     bool             `json:"auditMatch"`
	ConfigHistory  []*domain.Config `json:"configHistory"`
}

func (s *Service) Summary(id string) (*TaskSummary, error) {
	return s.SummaryFiltered(id, "", false)
}
func (s *Service) SummaryFiltered(id, severity string, openOnly bool) (*TaskSummary, error) {
	if v, ok := s.repo.(interface{ Verify() error }); ok {
		if err := v.Verify(); err != nil {
			return nil, err
		}
	}
	t, err := s.repo.Get(id)
	if err != nil {
		return nil, err
	}
	// Only risks attached to the active configuration participate in the current view.
	activeRisks := map[string]*domain.Risk{}
	counts := map[string]int{"low": 0, "medium": 0, "high": 0, "critical": 0}
	openIDs := []string{}
	for _, r := range t.Risks {
		if !r.Active || r.ConfigID != t.ActiveConfigID {
			continue
		}
		activeRisks[r.ID] = r
	}
	filtered := map[string]*domain.Risk{}
	for id, r := range activeRisks {
		if severity != "" && r.Severity != severity {
			continue
		}
		if openOnly && r.HasEvidence() {
			continue
		}
		filtered[id] = r
		counts[r.Severity]++
		if r.Evidence == "" {
			openIDs = append(openIDs, id)
		}
	}
	t.Risks = filtered
	sort.Strings(openIDs)
	for _, c := range t.Configs {
		c.Active = c.ID == t.ActiveConfigID
	}
	valid := false
	digest := ""
	hashMatch, auditMatch := false, false
	if t.Permit != nil {
		valid = t.Permit.Valid(s.now())
		permitCopy := *t.Permit
		permitCopy.PermitDigest = ""
		if t.Permit.PermitDigest != "" && t.Permit.PermitDigest != domain.PermitDigest(&permitCopy) {
			valid = false
		}
		if c, ok := t.Configs[t.Permit.ConfigID]; !ok || c.ContentHash != t.Permit.ConfigHash || c.Frozen == false {
			valid = false
		} else {
			hashMatch = true
		}
		digest = t.Permit.AuditDigest
		ev, _ := s.repo.Events(id)
		prior := make([]domain.Event, 0, len(ev))
		for _, e := range ev {
			if e.Type != "permit_issued" && e.Type != "permit_revoked" {
				prior = append(prior, e)
			}
		}
		if persistence.AuditDigest(prior) != t.Permit.AuditDigest {
			valid = false
		} else {
			auditMatch = true
		}
	}
	if t.Permit != nil && t.Permit.RevokedAt != nil {
		valid = false
	}
	history := t.ConfigHistory()
	sort.Slice(history, func(i, j int) bool { return history[i].Revision < history[j].Revision })
	for _, c := range history {
		c.Active = c.ID == t.ActiveConfigID
	}
	return &TaskSummary{Task: t, RiskCount: len(t.Risks), OpenRisks: len(openIDs), PermitValid: valid, AuditDigest: digest, SeverityCounts: counts, OpenRiskIDs: openIDs, HashMatch: hashMatch, AuditMatch: auditMatch, ConfigHistory: history}, nil
}

type idemRecord struct {
	Fingerprint           string
	NormalizedFingerprint string
	Task                  *domain.Task
}

func fingerprint(v any) string {
	b, _ := json.Marshal(v)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
