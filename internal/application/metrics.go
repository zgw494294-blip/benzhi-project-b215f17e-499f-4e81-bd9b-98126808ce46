package application

import "sync/atomic"

type Metrics struct{ Tasks, Configs, Validations, Permits atomic.Uint64 }

func (m *Metrics) Snapshot() map[string]uint64 {
	return map[string]uint64{"tasks": m.Tasks.Load(), "configs": m.Configs.Load(), "validations": m.Validations.Load(), "permits": m.Permits.Load()}
}
