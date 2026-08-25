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
	mu           sync.Mutex
	path         string
	seq          int64
	verified     bool
	verifiedSize int64
}

func NewLedger(path string) *Ledger { return &Ledger{path: path} }
func (l *Ledger) Append(e domain.Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.seq++
	e.Sequence = l.seq
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
	if err == nil {
		l.verified = false
	}
	return err
}
func (l *Ledger) Verify() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	b, err := os.ReadFile(l.path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if l.verified {
		if info, statErr := os.Stat(l.path); statErr == nil && info.Size() == l.verifiedSize {
			return nil
		}
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
	l.verified = true
	l.verifiedSize = int64(len(b))
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
