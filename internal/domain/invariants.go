package domain

import "fmt"

func (t *Task) ValidateInvariants() error {
	if t.ID == "" || t.Version < 1 {
		return fmt.Errorf("invalid task identity")
	}
	if t.ActiveConfigID != "" {
		if _, ok := t.Configs[t.ActiveConfigID]; !ok {
			return fmt.Errorf("active config missing")
		}
	}
	if t.Status == StatusReleased && t.Permit == nil {
		return fmt.Errorf("released task requires permit")
	}
	for id, c := range t.Configs {
		if id != c.ID || c.TaskID != t.ID {
			return fmt.Errorf("config ownership mismatch")
		}
		if c.ContentHash != "" && HashConfig(c) != c.ContentHash {
			return fmt.Errorf("config hash mismatch")
		}
	}
	for id, r := range t.Risks {
		if id != r.ID || r.TaskID != t.ID {
			return fmt.Errorf("risk ownership mismatch")
		}
	}
	return nil
}
func (t *Task) ConfigHistory() []*Config {
	out := make([]*Config, 0, len(t.Configs))
	for _, c := range t.Configs {
		out = append(out, c.Clone())
	}
	return out
}
func (t *Task) RiskCount() int {
	n := 0
	for _, r := range t.Risks {
		if r.Active && r.ConfigID == t.ActiveConfigID {
			n++
		}
	}
	return n
}
func (t *Task) HasPermit() bool { return t.Permit != nil && t.Permit.RevokedAt == nil }
func (t *Task) CanEdit() bool   { return t.Status != StatusReleased && t.Status != StatusApproved }
func (t *Task) CanReview() bool {
	return t.Status == StatusPendingReview || t.Status == StatusValidated || t.Status == StatusRemediation
}
func (t *Task) CanIssue() bool {
	return t.Status == StatusApproved && (t.Permit == nil || t.Permit.RevokedAt != nil)
}
