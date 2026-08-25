package application

import "sync"

type Idempotency struct {
	mu     sync.Mutex
	values map[string]any
}

func NewIdempotency() *Idempotency { return &Idempotency{values: map[string]any{}} }
func (i *Idempotency) Get(k string) (any, bool) {
	i.mu.Lock()
	defer i.mu.Unlock()
	v, ok := i.values[k]
	return v, ok
}
func (i *Idempotency) Put(k string, v any) { i.mu.Lock(); defer i.mu.Unlock(); i.values[k] = v }
func (i *Idempotency) Clear()              { i.mu.Lock(); defer i.mu.Unlock(); i.values = map[string]any{} }
