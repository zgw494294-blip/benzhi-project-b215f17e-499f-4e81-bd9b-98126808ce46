package persistence

import (
	"crypto/sha256"
	"deepdeploy/internal/domain"
	"encoding/hex"
	"encoding/json"
)

func AuditDigest(events []domain.Event) string {
	b, _ := json.Marshal(events)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
