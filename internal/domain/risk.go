package domain

import "time"

func (r *Risk) HasEvidence() bool { return r != nil && r.Evidence != "" && r.Mitigation != "" }
func (r *Risk) IsClosed() bool {
	return r.HasEvidence() && (r.ReviewDecision == "approve" || r.ReviewDecision == "reject")
}
func (r *Risk) Reviewable() bool { return r != nil && r.HasEvidence() }
func (t *Task) OpenRisks() []*Risk {
	out := []*Risk{}
	for _, r := range t.Risks {
		if !r.HasEvidence() {
			out = append(out, r)
		}
	}
	return out
}
func (t *Task) ReviewedAt() time.Time {
	var latest time.Time
	for _, r := range t.Risks {
		if r.ReviewedAt.After(latest) {
			latest = r.ReviewedAt
		}
	}
	return latest
}
