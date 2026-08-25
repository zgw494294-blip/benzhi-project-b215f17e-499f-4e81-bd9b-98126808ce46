package domain

import "time"

type Event struct {
	SchemaVersion int       `json:"schemaVersion"`
	Sequence      int64     `json:"sequence"`
	Type          string    `json:"type"`
	TaskID        string    `json:"taskID"`
	At            time.Time `json:"at"`
	Data          any       `json:"data"`
	Digest        string    `json:"digest"`
}
type Repository interface {
	Create(*Task) error
	Get(string) (*Task, error)
	Save(*Task, Event) error
	Events(string) ([]Event, error)
}

// IdempotencyLookup lets durable stores recover create-request associations after restart.
type IdempotencyLookup interface {
	FindByIdempotencyKey(string) (*Task, error)
}
