package restart_replay_loss

import (
	"testing"

	"deepdeploy/internal/application"
	"deepdeploy/internal/persistence"
)

func TestRestartReplaysLedger(t *testing.T) {
	dir := t.TempDir()
	store, err := persistence.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	service := application.NewService(store)
	if _, err := service.CreateTask("task-replay", "深潜任务", "海域A", "窗口", "负责人"); err != nil {
		t.Fatal(err)
	}
	if err := store.Remove("task-replay"); err != nil {
		t.Fatal(err)
	}

	restarted, err := persistence.NewFileStore(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := restarted.Get("task-replay"); err != nil {
		t.Fatalf("restart replay lost task: %v", err)
	}
}
