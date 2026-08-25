package application

import (
	"deepdeploy/internal/domain"
	"strings"
	"unicode"
)

func hasControl(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func ValidateTaskInput(in TaskInput) error {
	if strings.TrimSpace(in.ID) == "" || strings.TrimSpace(in.MissionName) == "" || strings.TrimSpace(in.SeaArea) == "" || strings.TrimSpace(in.DeploymentWindow) == "" || strings.TrimSpace(in.Owner) == "" {
		return domain.ErrInvalid
	}
	if hasControl(in.ID+in.MissionName+in.SeaArea+in.DeploymentWindow+in.Owner) || domain.ValidateDeploymentWindow(strings.TrimSpace(in.DeploymentWindow)) != nil {
		return domain.ErrInvalid
	}
	return nil
}
func ValidateEvidenceInput(in EvidenceInput) error {
	if strings.TrimSpace(in.Mitigation) == "" || strings.TrimSpace(in.Evidence) == "" || len(in.Mitigation) > 4096 || len(in.Evidence) > 16384 || hasControl(in.Mitigation+in.Evidence) {
		return domain.ErrInvalid
	}
	return nil
}
func ValidateReviewInput(in ReviewInput) error {
	if in.Decision != "approve" && in.Decision != "reject" {
		return domain.ErrInvalid
	}
	if strings.TrimSpace(in.Reviewer) == "" {
		return domain.ErrInvalid
	}
	return nil
}
