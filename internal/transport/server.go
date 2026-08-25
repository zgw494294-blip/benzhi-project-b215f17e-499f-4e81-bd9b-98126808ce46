package transport

import (
	"deepdeploy/internal/application"
	"encoding/json"
	"net/http"
	"time"
)

type Server struct {
	app  *application.Service
	addr string
	http *http.Server
}

func NewServer(app *application.Service, addr string) *Server {
	s := &Server{app: app, addr: addr}
	s.http = &http.Server{Addr: addr, Handler: s.routes(), ReadHeaderTimeout: 3 * time.Second}
	return s
}
func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }

// Handler exposes the configured HTTP routes for embedded callers and tests.
func (s *Server) Handler() http.Handler { return s.routes() }
func (s *Server) routes() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/v1/tasks", s.tasks)
	m.HandleFunc("/v1/tasks/", s.taskSubresource)
	m.HandleFunc("/healthz", healthHandler)
	return withHeaders(logging(m))
}
func decode(r *http.Request, v any) error { return json.NewDecoder(r.Body).Decode(v) }
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func writeErr(w http.ResponseWriter, err error) {
	status := statusFor(err)
	code := "invalid_request"
	if status == 404 {
		code = "not_found"
	} else if status == 409 {
		code = "conflict"
	}
	writeJSON(w, status, map[string]any{"error": err.Error(), "code": code})
}
