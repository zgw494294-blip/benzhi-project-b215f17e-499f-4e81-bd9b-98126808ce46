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
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		id := e.Name()[:len(e.Name())-5]
		t, er := ReadSnapshot(f.dir, id)
		if er != nil {
			continue
		}
		_ = f.memory.Create(t)
	}
	if events, err := ReadEvents(filepath.Join(f.dir, "events.jsonl")); err == nil {
		for _, ev := range events {
			f.memory.events[ev.TaskID] = append(f.memory.events[ev.TaskID], ev)
			if ev.Sequence > f.memory.seq {
				f.memory.seq = ev.Sequence
			}
			if ev.Sequence > f.ledger.seq {
				f.ledger.seq = ev.Sequence
			}
		}
	}
	return nil
}
func (f *FileStore) Create(t *domain.Task) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Persist first; only commit to the in-memory cache on success so a snapshot
	// failure cannot leave a task visible to readers that disappears after restart.
	snapshotErr := WriteSnapshot(f.dir, t)
	if snapshotErr != nil {
		return snapshotErr
	}
	return f.memory.Create(t)
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
	// Validate against the current in-memory task version without advancing any
	// in-memory state. Persistence steps run first; the in-memory task version,
	// event sequence and event index only advance once every durable step has
	// succeeded.
	cur, err := f.memory.Get(t.ID)
	if err != nil {
		return err
	}
	if t.Version < cur.Version {
		return domain.ErrConflict
	}
	if err := f.ledger.Append(&e); err != nil {
		return err
	}
	if err := WriteSnapshot(f.dir, t); err != nil {
		// The ledger line was durable but the snapshot was not, which would leave
		// a durable event with no corresponding in-memory state that survives a
		// restart. Roll the ledger back so the failed save leaves no trace.
		_ = f.ledger.Truncate()
		return err
	}
	return f.memory.CommitSave(t, e)
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
