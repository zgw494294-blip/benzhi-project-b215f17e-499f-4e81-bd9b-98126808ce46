package persistence

import (
	"bufio"
	"deepdeploy/internal/domain"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type FileStore struct {
	mu     sync.Mutex
	dir    string
	memory *MemoryStore
	ledger *Ledger
}

func NewFileStore(dir string) (*FileStore, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}
	f := &FileStore{dir: dir, memory: NewMemoryStore(), ledger: NewLedger(filepath.Join(dir, "events.jsonl"))}
	if err := f.load(); err != nil {
		return nil, err
	}
	return f, nil
}
func (f *FileStore) load() error {
	entries, err := os.ReadDir(f.dir)
	if err != nil {
		return err
	}
	snapshotCount := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		t, er := ReadSnapshot(f.dir, id)
		if er != nil {
			continue
		}
		snapshotCount++
		_ = f.memory.Create(t)
	}
	// Load events from the ledger so they are available for queries and so the
	// sequence counter is advanced past the highest persisted event.
	eventsByTask := map[string][]domain.Event{}
	if events, err := ReadEvents(filepath.Join(f.dir, "events.jsonl")); err == nil {
		for _, ev := range events {
			eventsByTask[ev.TaskID] = append(eventsByTask[ev.TaskID], ev)
			if ev.Sequence > f.memory.seq {
				f.memory.seq = ev.Sequence
			}
			if ev.Sequence > f.ledger.seq {
				f.ledger.seq = ev.Sequence
			}
		}
	}
	// Replay events for tasks that have no snapshot so that the in-memory state
	// is recovered from the durable ledger.
	for taskID, evs := range eventsByTask {
		if _, ok := f.memory.tasks[taskID]; ok {
			// Snapshot already loaded; just attach the event history.
			f.memory.events[taskID] = evs
			continue
		}
		t, er := ReplayTask(evs)
		if er != nil {
			continue
		}
		_ = f.memory.Create(t)
		f.memory.events[taskID] = evs
	}
	return nil
}
func (f *FileStore) Create(t *domain.Task) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.memory.Create(t); err != nil {
		return err
	}
	return WriteSnapshot(f.dir, t)
}
func (f *FileStore) Get(id string) (*domain.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.memory.Get(id)
}
func (f *FileStore) FindByIdempotencyKey(key string) (*domain.Task, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.memory.FindByIdempotencyKey(key)
}
func (f *FileStore) Save(t *domain.Task, e domain.Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.memory.Save(t, e); err != nil {
		return err
	}
	if err := f.ledger.Append(e); err != nil {
		return err
	}
	return WriteSnapshot(f.dir, t)
}
func (f *FileStore) Events(id string) ([]domain.Event, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.memory.Events(id)
}
func (f *FileStore) Verify() error { return f.ledger.Verify() }
func ReadEvents(path string) ([]domain.Event, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var out []domain.Event
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		var e domain.Event
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, sc.Err()
}
func (f *FileStore) SnapshotPath(id string) string { return filepath.Join(f.dir, id+".json") }
func (f *FileStore) Remove(id string) error        { return os.Remove(f.SnapshotPath(id)) }
