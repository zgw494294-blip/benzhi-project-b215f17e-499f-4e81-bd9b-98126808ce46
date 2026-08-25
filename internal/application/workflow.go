package application

import (
	"deepdeploy/internal/domain"
	"fmt"
	"time"
)

type WorkflowResult struct {
	Task   *domain.Task
	Risks  []*domain.Risk
	Permit *domain.Permit
}

func (s *Service) Execute(taskID string, cfg *domain.Config, reviewer, issuer string) (WorkflowResult, error) {
	var out WorkflowResult
	t, e := s.repo.Get(taskID)
	if e != nil {
		return out, e
	}
	v := t.Version
	t, e = s.AddConfig(taskID, cfg, v, "workflow-"+cfg.ID)
	if e != nil {
		return out, e
	}
	v = t.Version
	t, risks, e := s.Validate(taskID, cfg.ID, v)
	if e != nil {
		return out, e
	}
	out.Task = t
	out.Risks = risks
	if len(risks) > 0 {
		for _, r := range risks {
			if _, e = s.AddEvidence(taskID, r.ID, r.Mitigation, "workflow-evidence", t.Version, "workflow-evidence-"+r.ID); e != nil {
				return out, e
			}
			t, _ = s.repo.Get(taskID)
		}
		v = t.Version
	} else {
		v = t.Version
	}
	if len(t.Risks) > 0 {
		for _, r := range t.Risks {
			t, e = s.ReviewRisk(taskID, r.ID, "approve", reviewer, "自动流程", t.Version)
			if e != nil {
				return out, e
			}
		}
	} else {
		t, e = s.Review(taskID, "approve", reviewer, "自动流程", v)
	}
	if e != nil {
		return out, e
	}
	p, e := s.Permit(taskID, issuer, t.Version)
	out.Task = t
	out.Permit = p
	return out, e
}
func (s *Service) WaitForReview(taskID string, timeout time.Duration) (*domain.Task, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		t, e := s.repo.Get(taskID)
		if e != nil {
			return nil, e
		}
		if t.Status == domain.StatusPendingReview || t.Status == domain.StatusValidated {
			return t, nil
		}
		time.Sleep(5 * time.Millisecond)
	}
	return nil, fmt.Errorf("review timeout")
}
