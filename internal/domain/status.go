package domain

func (s TaskStatus) Terminal() bool { return s == StatusReleased }
func (s TaskStatus) String() string { return string(s) }
func AllowedTransition(from, to TaskStatus) bool {
	switch from {
	case StatusDraft:
		return to == StatusValidated || to == StatusRemediation
	case StatusValidated:
		return to == StatusApproved || to == StatusRemediation
	case StatusRemediation:
		return to == StatusPendingReview || to == StatusApproved
	case StatusPendingReview:
		return to == StatusApproved || to == StatusRemediation
	case StatusApproved:
		return to == StatusReleased
	}
	return false
}
