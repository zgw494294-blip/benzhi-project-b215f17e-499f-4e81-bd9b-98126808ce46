package persistence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"

	"deepdeploy/internal/domain"
)

type MemoryStore struct {
	mu     sync.RWMutex
	tasks  map[string]*domain.Task
	events map[string][]domain.Event
	seq    int64
}

func (s *MemoryStore) Verify() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, list := range s.events {
		for _, e := range list {
			if e.Digest == "" {
				continue
			}
			given := e.Digest
			e.Digest = ""
			raw, _ := json.Marshal(e)
			sum := sha256.Sum256(raw)
			if given != hex.EncodeToString(sum[:]) {
				return domain.ErrConflict
			}
		}
	}
	return nil
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{tasks: map[string]*domain.Task{}, events: map[string][]domain.Event{}}
}
func (s *MemoryStore) Create(t *domain.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[t.ID]; ok {
		return domain.ErrConflict
	}
	s.tasks[t.ID] = cloneTask(t)
	return nil
}
func (s *MemoryStore) Get(id string) (*domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneTask(t), nil
}
func (s *MemoryStore) FindByIdempotencyKey(key string) (*domain.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		if t.IdempotencyKey == key {
			return cloneTask(t), nil
		}
	}
	return nil, domain.ErrNotFound
}
func (s *MemoryStore) Save(t *domain.Task, e domain.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur, ok := s.tasks[t.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if t.Version < cur.Version {
		return domain.ErrConflict
	}
	s.tasks[t.ID] = cloneTask(t)
	s.seq++
	e.Sequence = s.seq
	e.SchemaVersion = 1
	raw, _ := json.Marshal(e)
	sum := sha256.Sum256(raw)
	e.Digest = hex.EncodeToString(sum[:])
	s.events[t.ID] = append(s.events[t.ID], e)
	return nil
}
func (s *MemoryStore) Events(id string) ([]domain.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if _, ok := s.tasks[id]; !ok {
		return nil, domain.ErrNotFound
	}
	return append([]domain.Event(nil), s.events[id]...), nil
}
func cloneTask(t *domain.Task) *domain.Task {
	b, _ := json.Marshal(t)
	var out domain.Task
	_ = json.Unmarshal(b, &out)
	if out.Configs == nil {
		out.Configs = map[string]*domain.Config{}
	}
	if out.Risks == nil {
		out.Risks = map[string]*domain.Risk{}
	}
	return &out
}
