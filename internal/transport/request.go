package transport

import (
	"net/http"
	"strconv"
)

func expectedVersion(r *http.Request) (int, error) {
	v := r.Header.Get("X-Expected-Version")
	if v == "" {
		return 0, nil
	}
	return strconv.Atoi(v)
}
func idempotencyKey(r *http.Request) string { return r.Header.Get("Idempotency-Key") }
func contentTypeJSON(r *http.Request) bool {
	return r.Header.Get("Content-Type") == "application/json" || r.Header.Get("Content-Type") == ""
}
