package domain

import (
	"testing"
	"time"
)

func TestTaskLifecycle(t *testing.T) {
	now := time.Unix(1, 0)
	task, err := NewTask("t1", "任务", "海域", "窗口", "负责人", now)
	if err != nil {
		t.Fatal(err)
	}
	c := &Config{ID: "c1", TaskID: "t1", Sensors: []Sensor{{ID: "s", Model: "m", Bus: "b", SampleRate: 10}}, MountingParameters: map[string]float64{"depth_m": 10}}
	if err := task.AddConfig(c, 1, now); err != nil {
		t.Fatal(err)
	}
	if _, err := task.ApplyFindings("c1", nil, now); err != nil {
		t.Fatal(err)
	}
	if err := task.Review("approve", "安全", "", 3, now); err != nil {
		t.Fatal(err)
	}
	if _, err := task.FreezeAndPermit("p", "S", "安全", "d", 4, now); err != nil {
		t.Fatal(err)
	}
	if !c.Frozen || task.Status != StatusReleased {
		t.Fatalf("unexpected final state: %+v", task)
	}
}
