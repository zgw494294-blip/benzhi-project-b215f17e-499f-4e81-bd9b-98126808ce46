package persistence

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"deepdeploy/internal/domain"
)

type Ledger struct {
	mu   sync.Mutex
	path string
	seq  int64
}

func NewLedger(path string) *Ledger { return &Ledger{path: path} }
func (l *Ledger) Append(e domain.Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq++
	e.Sequence = l.seq
	if e.SchemaVersion == 0 {
		e.SchemaVersion = CurrentSchemaVersion
	}
	e.Digest = ""
	raw, _ := json.Marshal(e)
	sum := sha256.Sum256(raw)
	e.Digest = hex.EncodeToString(sum[:])
	raw, _ = json.Marshal(e)
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(raw, '\n'))
	return err
}
func (l *Ledger) Verify() error {
	b, err := os.ReadFile(l.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, line := range splitLines(b) {
		if len(line) == 0 {
			continue
		}
		var e domain.Event
		if json.Unmarshal(line, &e) != nil {
			return fmt.Errorf("invalid ledger line")
		}
		given := e.Digest
		if given == "" {
			continue
		}
		// Recompute the digest from the exact line bytes with the digest field
		// blanked out. Re-marshaling after an Unmarshal round-trip would reorder
		// keys for struct-typed Data fields (they become map[string]any), so we
		// must verify against the bytes that were actually written to the file.
		redacted := blankDigest(line)
		sum := sha256.Sum256(redacted)
		if given != hex.EncodeToString(sum[:]) {
			return fmt.Errorf("ledger digest mismatch")
		}
	}
	return nil
}

// blankDigest replaces the "digest" field value in raw JSON bytes with an empty
// string, preserving everything else (including key/field ordering) so that the
// hash matches the bytes originally written by Append (which marshals with
// Digest == ""). It operates on the raw bytes to avoid reordering map keys that
// happens when re-marshaling after an Unmarshal round-trip.
func blankDigest(raw []byte) []byte {
	var e struct{ Digest string }
	if json.Unmarshal(raw, &e) != nil {
		return raw
	}
	if e.Digest == "" {
		return raw
	}
	needle := []byte(`"digest":"` + e.Digest + `"`)
	repl := []byte(`"digest":""`)
	if bytes.Contains(raw, needle) {
		return bytes.Replace(raw, needle, repl, 1)
	}
	// Fallback for pretty-printed JSON with spaces.
	needle = []byte(`"digest": "` + e.Digest + `"`)
	repl = []byte(`"digest": ""`)
	return bytes.Replace(raw, needle, repl, 1)
}
func splitLines(b []byte) [][]byte {
	var out [][]byte
	start := 0
	for i, c := range b {
		if c == '\n' {
			out = append(out, b[start:i])
			start = i + 1
		}
	}
	if start < len(b) {
		out = append(out, b[start:])
	}
	return out
}
