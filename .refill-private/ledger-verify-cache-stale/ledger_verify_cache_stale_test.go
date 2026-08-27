package ledger_verify_cache_stale

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"deepdeploy/internal/application"
	"deepdeploy/internal/persistence"
)

func TestLedgerCorruptionAfterVerificationIsDetected(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewService(store)
	if _, err = service.CreateTask("ledger-task", "任务", "海域", "2026-08-25T00:00:00Z/2026-08-26T00:00:00Z", "负责人"); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Events("ledger-task"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "events.jsonl")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	marker := `"digest":"`
	start := strings.Index(text, marker)
	if start < 0 || start+len(marker) >= len(raw) {
		t.Fatal("ledger has no digest")
	}
	idx := start + len(marker)
	replacement := byte('0')
	if raw[idx] == replacement {
		replacement = '1'
	}
	raw[idx] = replacement
	if err = os.WriteFile(path, raw, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = service.Events("ledger-task"); err == nil {
		t.Fatal("corrupt ledger was silently accepted")
	}
}
