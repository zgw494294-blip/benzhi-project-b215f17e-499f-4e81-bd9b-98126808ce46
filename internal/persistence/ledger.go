package persistence

import (
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
func (l *Ledger) Append(e *domain.Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	// Finalize the event with the next sequence number before persisting so the
	// in-memory mutation only becomes visible to callers once the line is durable.
	next := l.seq + 1
	finalized := *e
	finalized.Sequence = next
	finalized.SchemaVersion = CurrentSchemaVersion
	raw, _ := json.Marshal(finalized)
	sum := sha256.Sum256(raw)
	finalized.Digest = hex.EncodeToString(sum[:])
	raw, _ = json.Marshal(finalized)
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(raw, '\n')); err != nil {
		return err
	}
	// The append is durable; only now advance the ledger sequence and publish the
	// finalized event fields to the caller.
	l.seq = next
	*e = finalized
	return nil
}
// Truncate rolls back the most recent appended line. It is used to undo a
// successful ledger append when a later persistence step (e.g. snapshot write)
// fails, so failed saves leave neither in-memory nor durable state behind.
func (l *Ledger) Truncate() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, err := os.ReadFile(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	// Split into newline-delimited records and drop the final one. A trailing
	// newline after the last record does not produce an extra empty record, so the
	// line count matches the number of appended events.
	records := splitLines(b)
	if len(records) == 0 {
		return nil
	}
	kept := records[:len(records)-1]
	out := make([]byte, 0)
	for _, ln := range kept {
		out = append(out, ln...)
		out = append(out, '\n')
	}
	if err := os.WriteFile(l.path, out, 0600); err != nil {
		return err
	}
	if l.seq > 0 {
		l.seq--
	}
	return nil
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
		e.Digest = ""
		raw, _ := json.Marshal(e)
		sum := sha256.Sum256(raw)
		if given != "" && given != hex.EncodeToString(sum[:]) {
			return fmt.Errorf("ledger digest mismatch")
		}
	}
	return nil
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
