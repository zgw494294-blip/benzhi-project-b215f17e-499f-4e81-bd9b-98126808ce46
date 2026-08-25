package application

import (
	"deepdeploy/internal/domain"
	"time"
)

func Event(typ, task string, data any) domain.Event {
	return domain.Event{SchemaVersion: 1, Type: typ, TaskID: task, At: time.Now(), Data: data}
}
