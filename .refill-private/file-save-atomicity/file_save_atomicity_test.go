package persistence

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
)

func TestFileStoreSaveDoesNotPublishWhenLedgerAppendFails(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	task, err := domain.NewTask("task", "任务", "海域", "窗口", "负责人", time.Unix(1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Create(task); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "events.jsonl"), 0700); err != nil {
		t.Fatal(err)
	}
	cfg := &domain.Config{ID: "cfg", TaskID: task.ID, Sensors: []domain.Sensor{{ID: "s", Bus: "bus"}}}
	task.Version = 2
	if err := store.Save(task, domain.Event{Type: "config_created", TaskID: task.ID}); err == nil {
		t.Fatal("expected ledger append failure")
	}
	got, err := store.Get(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != 1 || got.Configs[cfg.ID] != nil {
		t.Fatalf("failed save leaked into repository: version=%d configs=%v", got.Version, got.Configs)
	}
	events, err := store.Events(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("failed save leaked event: %d", len(events))
	}
}
