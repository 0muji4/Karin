package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/0muji4/Karin/apps/backend/internal/auth"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/ko"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

// Pinger は依存（DB）の疎通確認だけを抽象化する。pgxpool.Pool が満たす。
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps は Server が必要とする協力者（ポート）をまとめる。具体実装は cmd（合成ルート）で注入する。
type Deps struct {
	DB      Pinger                // /healthz の疎通確認
	Ko      ko.Catalog            // 候メタの読み取りポート
	Auth    *auth.Service         // 匿名アカウントと認証
	Records *record.Service       // 文箱の読み書き
	Cast    *exchange.CastService // 風に乗せる
	Inbox   exchange.Inbox        // 受信（風だより）の読み取り・文箱にしまう
}

// Server は API のハンドラ群と依存を束ねる。
type Server struct {
	logger  *slog.Logger
	db      Pinger
	ko      ko.Catalog
	auth    *auth.Service
	records *record.Service
	cast    *exchange.CastService
	inbox   exchange.Inbox
}

// NewServer は Server を組み立てる。logger が nil なら既定の slog を使う。
func NewServer(logger *slog.Logger, deps Deps) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		logger:  logger,
		db:      deps.DB,
		ko:      deps.Ko,
		auth:    deps.Auth,
		records: deps.Records,
		cast:    deps.Cast,
		inbox:   deps.Inbox,
	}
}

// Handler はルーティングと middleware を組み上げた http.Handler を返す。
// Go 1.22+ の method-pattern ServeMux を使い、フレームワークは入れない。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 認証不要: 稼働確認・アカウント発行・今日の候（一般情報）。
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /accounts", s.handleCreateAccount)
	mux.HandleFunc("GET /ko/today", s.handleTodayKo)

	// 認証必須: 文箱は本人だけ。風に乗せるも本人の記録に対してのみ。
	mux.Handle("POST /records", s.requireAuth(http.HandlerFunc(s.handleCreateRecord)))
	mux.Handle("GET /box", s.requireAuth(http.HandlerFunc(s.handleListBox)))
	mux.Handle("POST /records/{id}/cast", s.requireAuth(http.HandlerFunc(s.handleCast)))
	mux.Handle("GET /deliveries", s.requireAuth(http.HandlerFunc(s.handleListDeliveries)))
	mux.Handle("POST /deliveries/{id}/keep", s.requireAuth(http.HandlerFunc(s.handleKeep)))

	// middleware は外側から: recover -> log -> mux。
	var h http.Handler = mux
	h = logMiddleware(s.logger, h)
	h = recoverMiddleware(s.logger, h)
	return h
}
