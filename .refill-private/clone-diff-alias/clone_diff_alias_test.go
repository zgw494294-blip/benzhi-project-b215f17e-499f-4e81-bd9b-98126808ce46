package clone_diff_alias

import (
	"testing"

	"deepdeploy/internal/application"
	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
)

func TestConfigIdempotencyCloneDoesNotShareDiffSlices(t *testing.T) {
	app := application.NewService(persistence.NewMemoryStore())
	if _, err := app.CreateTask("task", "任务", "海域", "窗口", "负责人"); err != nil {
		t.Fatal(err)
	}
	first := &domain.Config{ID: "cfg-1", TaskID: "task", Sensors: []domain.Sensor{{ID: "s1", Bus: "bus1"}}}
	if _, err := app.AddConfig("task", first, 1, ""); err != nil {
		t.Fatal(err)
	}
	second := &domain.Config{ID: "cfg-2", TaskID: "task", Sensors: []domain.Sensor{{ID: "s2", Bus: "bus2"}}}
	result, err := app.AddConfig("task", second, 2, "idem")
	if err != nil {
		t.Fatal(err)
	}
	result.Configs["cfg-2"].Diff.SensorsAdded[0] = "mutated"
	retryInput := &domain.Config{ID: "cfg-2", TaskID: "task", Sensors: []domain.Sensor{{ID: "s2", Bus: "bus2"}}}
	retry, err := app.AddConfig("task", retryInput, 999, "idem")
	if err != nil {
		t.Fatal(err)
	}
	if retry.Configs["cfg-2"].Diff.SensorsAdded[0] == "mutated" {
		t.Fatal("idempotency retry exposed caller mutation through shallow Diff clone")
	}
}
