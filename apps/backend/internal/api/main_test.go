package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/api"
	"github.com/0muji4/Karin/apps/backend/internal/auth"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
	"github.com/0muji4/Karin/apps/backend/internal/record"
	"github.com/0muji4/Karin/apps/backend/internal/report"
)

// testPool は結合テストで共有する接続プール（TestMain で 1 度だけ用意する）。
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	flag.Parse() // testing.Short() を読む前にフラグを解釈する。
	if testing.Short() {
		// 結合テストは個別に Skip する。ユニットテスト（health）はそのまま走る。
		os.Exit(m.Run())
	}
	ctx := context.Background()
	pool, _, terminate, err := dbtest.MigratedPool(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "PG 起動失敗:", err)
		os.Exit(1)
	}
	testPool = pool
	code := m.Run()
	terminate()
	os.Exit(code)
}

// handlerForTest は実サービスを差した本番同等の Handler を返す。
func handlerForTest() http.Handler {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return api.NewServer(logger, api.Deps{
		DB:      testPool,
		Ko:      postgres.NewKoCatalog(testPool),
		Auth:    auth.NewService(postgres.NewAuthRepo(testPool)),
		Records: record.NewService(postgres.NewRecordRepo(testPool)),
		Cast:    exchange.NewCastService(postgres.NewRecordRepo(testPool), moderation.AllPass{}, postgres.NewPoolRepo(testPool)),
		Inbox:   postgres.NewInboxRepo(testPool),
		Reports: report.NewService(moderation.AllPass{}, postgres.NewReportRepo(testPool)),
	}).Handler()
}

// doJSON は JSON リクエストを送り、応答レコーダを返す。token が空なら Authorization を付けない。
func doJSON(t *testing.T, h http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("リクエストの JSON 化に失敗: %v", err)
		}
		reader = bytes.NewReader(buf)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

// createAccount は POST /accounts でアカウントを作り、user_id と token を返す。
func createAccount(t *testing.T, h http.Handler) (userID, token string) {
	t.Helper()
	rr := doJSON(t, h, http.MethodPost, "/accounts", "", nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("POST /accounts = %d, want 201 (body: %s)", rr.Code, rr.Body.String())
	}
	var resp struct {
		UserID string `json:"user_id"`
		Token  string `json:"token"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("アカウント応答の JSON デコードに失敗: %v", err)
	}
	if resp.UserID == "" || resp.Token == "" {
		t.Fatalf("user_id / token が空: %+v", resp)
	}
	return resp.UserID, resp.Token
}
