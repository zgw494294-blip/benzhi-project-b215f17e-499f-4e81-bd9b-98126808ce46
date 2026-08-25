package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

func PermitDigest(p *Permit) string {
	b, _ := json.Marshal(p)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
func (p *Permit) Valid(at time.Time) bool {
	return p != nil && p.RevokedAt == nil && !p.IssuedAt.After(at) && p.ConfigHash != ""
}
func (t *Task) RevokePermit(at time.Time) error {
	if t.Permit == nil || t.Status != StatusReleased {
		return ErrNotFound
	}
	if t.Permit.RevokedAt != nil {
		return ErrInvalidState
	}
	t.Permit.RevokedAt = &at
	t.Status = StatusApproved
	t.Version++
	t.UpdatedAt = at
	return nil
}
