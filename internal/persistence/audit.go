package persistence

import (
	"crypto/sha256"
	"deepdeploy/internal/domain"
	"encoding/hex"
	"encoding/json"
)

// AuditDigest computes a stable hash over a set of events. It canonicalizes each
// event by round-tripping it through json.Unmarshal into a generic structure so
// that struct-typed Data fields (which marshal in declaration order) and
// map-typed Data fields (which marshal in sorted key order) produce identical
// bytes. This ensures the digest matches across process restarts where events
// loaded from the ledger have Data as map[string]interface{} rather than the
// original concrete struct.
func AuditDigest(events []domain.Event) string {
	canonical := make([]map[string]any, 0, len(events))
	for _, e := range events {
		b, _ := json.Marshal(e)
		var m map[string]any
		_ = json.Unmarshal(b, &m)
		canonical = append(canonical, m)
	}
	b, _ := json.Marshal(canonical)
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}
