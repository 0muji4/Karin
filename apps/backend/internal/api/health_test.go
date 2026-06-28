package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0muji4/Karin/apps/backend/internal/api"
)

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

func newTestServer(p api.Pinger) *api.Server {
	return api.NewServer(slog.New(slog.NewTextHandler(io.Discard, nil)), api.Deps{DB: p})
}

func TestHealthz_OK(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	newTestServer(fakePinger{}).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	var body struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("JSON デコード失敗: %v", err)
	}
	if body.Status != "ok" {
		t.Errorf("status = %q, want ok", body.Status)
	}
}

func TestHealthz_DBDown(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	newTestServer(fakePinger{err: errors.New("接続不可")}).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rr.Code)
	}
}

func TestHealthz_MethodNotAllowed(t *testing.T) {
	// method-pattern ServeMux は GET 限定の /healthz に POST すると 405 を返す。
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	newTestServer(fakePinger{}).Handler().ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rr.Code)
	}
}
