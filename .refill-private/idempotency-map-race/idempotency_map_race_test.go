package idempotency_map_race

import (
	"sync"
	"testing"

	"deepdeploy/internal/application"
	"deepdeploy/internal/domain"
	"deepdeploy/internal/persistence"
)

func TestConcurrentConfigIdempotencyAccessIsRaceFree(t *testing.T) {
	s := application.NewService(persistence.NewMemoryStore())
	if _, err := s.CreateTask("task", "任务", "海域", "窗口", "负责人"); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			cfg := &domain.Config{ID: "cfg-" + string(rune('a'+i)), TaskID: "task", Sensors: []domain.Sensor{{ID: "s", Bus: "bus"}}}
			_, _ = s.AddConfig("task", cfg, 1, "key-"+string(rune('a'+i)))
		}(i)
	}
	close(start)
	wg.Wait()
}
