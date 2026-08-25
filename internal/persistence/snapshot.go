package persistence

import (
	"encoding/json"
	"os"
	"path/filepath"

	"deepdeploy/internal/domain"
)

func WriteSnapshot(dir string, t *domain.Task) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	tmp := filepath.Join(dir, t.ID+".tmp")
	final := filepath.Join(dir, t.ID+".json")
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	f, err := os.OpenFile(tmp, os.O_APPEND|os.O_WRONLY, 0600)
	if err == nil {
		_ = f.Sync()
		_ = f.Close()
	}
	return os.Rename(tmp, final)
}
func ReadSnapshot(dir, id string) (*domain.Task, error) {
	b, err := os.ReadFile(filepath.Join(dir, id+".json"))
	if err != nil {
		return nil, err
	}
	var t domain.Task
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
