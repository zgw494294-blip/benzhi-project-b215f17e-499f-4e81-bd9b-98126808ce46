package ledger_load_error_swallow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
)

func TestCorruptLedgerFailsStartup(t *testing.T) {
	dir := t.TempDir()
	task, err := domain.NewTask("task-corrupt-ledger", "任务", "海域", "窗口", "工程师", fixedTime())
	if err != nil {
		t.Fatal(err)
	}
	if err := persistence.WriteSnapshot(dir, task); err != nil {
		t.Fatal(err)
	}
	event, err := json.Marshal(domain.Event{SchemaVersion: 1, Sequence: 1, Type: "task_created", TaskID: task.ID, At: fixedTime(), Data: map[string]string{"owner": task.Owner}})
	if err != nil {
		t.Fatal(err)
	}
	ledger := append(event, '\n')
	ledger = append(ledger, []byte("not-json\n")...)
	if err := os.WriteFile(filepath.Join(dir, "events.jsonl"), ledger, 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := persistence.NewFileStore(dir); err == nil {
		t.Fatal("corrupt ledger was silently accepted")
	}
}

func fixedTime() (t time.Time) {
	return time.Date(2026, 8, 25, 0, 0, 0, 0, time.UTC)
}
