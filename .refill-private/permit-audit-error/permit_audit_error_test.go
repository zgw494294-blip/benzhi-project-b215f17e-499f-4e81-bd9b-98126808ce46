package permit_audit_error_test

import (
	"errors"
	"testing"

	"deepdeploy/internal/application"
	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
)

var errAuditRead = errors.New("audit store unavailable")

type eventsFailingRepo struct {
	inner *persistence.MemoryStore
	fail  bool
}

func (r *eventsFailingRepo) Create(t *domain.Task) error               { return r.inner.Create(t) }
func (r *eventsFailingRepo) Get(id string) (*domain.Task, error)       { return r.inner.Get(id) }
func (r *eventsFailingRepo) Save(t *domain.Task, e domain.Event) error { return r.inner.Save(t, e) }
func (r *eventsFailingRepo) Events(id string) ([]domain.Event, error) {
	if r.fail {
		return nil, errAuditRead
	}
	return r.inner.Events(id)
}

func TestPermitPropagatesAuditReadError(t *testing.T) {
	repo := &eventsFailingRepo{inner: persistence.NewMemoryStore()}
	app := application.NewService(repo)
	if _, err := app.CreateTask("task-audit", "任务", "海域", "窗口", "工程师"); err != nil {
		t.Fatal(err)
	}
	repo.fail = true
	cfg := &domain.Config{
		ID:                 "cfg-audit",
		TaskID:             "task-audit",
		Sensors:            []domain.Sensor{{ID: "sensor-1", Model: "M", Bus: "bus-1", SampleRate: 10}},
		FirmwareSet:        map[string]string{"sensor-1": "1.0"},
		MountingParameters: map[string]float64{"depth_m": 100},
		EnvironmentLimits:  map[string]float64{"max_depth_m": 1000, "max_pressure_bar": 100},
		SubmittedBy:        "工程师",
	}
	result, err := app.Execute("task-audit", cfg, "安全负责人", "安全负责人")
	if !errors.Is(err, errAuditRead) {
		t.Fatalf("permit audit read error was not propagated: got %v", err)
	}
	if result.Permit != nil {
		t.Fatalf("permit was issued after audit read failure: %+v", result.Permit)
	}
	task, getErr := repo.inner.Get("task-audit")
	if getErr != nil {
		t.Fatal(getErr)
	}
	if task.Permit != nil || task.Status != domain.StatusApproved {
		t.Fatalf("task changed despite audit read failure: status=%s permit=%v", task.Status, task.Permit)
	}
}
