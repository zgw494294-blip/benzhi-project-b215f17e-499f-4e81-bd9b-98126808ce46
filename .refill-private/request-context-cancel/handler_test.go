package requestcontextcancel_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"deepdeploy/internal/application"
	"deepdeploy/internal/persistence"
	"deepdeploy/internal/transport"
)

type blockingBody struct {
	entered chan struct{}
}

func (b *blockingBody) Read([]byte) (int, error) {
	close(b.entered)
	select {}
}

func (*blockingBody) Close() error { return nil }

func TestRequestCancellationStopsBodyDecode(t *testing.T) {
	app := application.NewService(persistence.NewMemoryStore())
	srv := transport.NewServer(app, "127.0.0.1:19082")
	body := &blockingBody{entered: make(chan struct{})}
	req := httptest.NewRequest("POST", "/v1/tasks", body)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		srv.Handler().ServeHTTP(rec, req)
		close(done)
	}()
	<-body.entered
	cancel()
	<-done
	if rec.Code != 400 {
		t.Fatalf("expected canceled request to return 400, got %d", rec.Code)
	}
}
