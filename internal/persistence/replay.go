package persistence

import "deepdeploy/internal/domain"

func Replay(events []domain.Event) map[string]int {
	out := map[string]int{}
	for _, e := range events {
		out[e.Type]++
	}
	return out
}
