package application

import (
	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
	"testing"
)

func TestServiceIdempotency(t *testing.T) {
	s := NewService(persistence.NewMemoryStore())
	if _, err := s.CreateTask("t", "任务", "海域", "窗口", "人"); err != nil {
		t.Fatal(err)
	}
	c := &domain.Config{ID: "c", TaskID: "t", Sensors: []domain.Sensor{{ID: "s", Model: "m", Bus: "b"}}}
	a, err := s.AddConfig("t", c, 1, "k")
	if err != nil {
		t.Fatal(err)
	}
	b, err := s.AddConfig("t", c, 99, "k")
	if err != nil {
		t.Fatal(err)
	}
	if a.Version != b.Version {
		t.Fatal("idempotency did not return original result")
	}
}
