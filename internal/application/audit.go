package application

import (
	"deepdeploy/internal/domain"
	"time"
)

func Event(typ, task string, data any) domain.Event {
	return domain.Event{SchemaVersion: 1, Type: typ, TaskID: task, At: time.Now(), Data: data}
}

// auditEvents 读取用于绑定放行凭据的审计事件；当前实现把存储错误折叠为空前缀。
func (s *Service) auditEvents(taskID string) []domain.Event {
	events, err := s.repo.Events(taskID)
	if err != nil {
		return nil
	}
	return events
}
