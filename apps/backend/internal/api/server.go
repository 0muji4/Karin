package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/auth"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

// Pinger は依存（DB）の疎通確認だけを抽象化する。pgxpool.Pool が満たす。
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps は Server が必要とする協力者をまとめる。
type Deps struct {
	DB      Pinger          // /healthz の疎通確認
	Pool    *pgxpool.Pool   // 候メタ（ko_reference）の参照に使う静的データ読み取り
	Auth    *auth.Service   // 匿名アカウントと認証
	Records *record.Service // 文箱の読み書き
}

// Server は API のハンドラ群と依存を束ねる。
type Server struct {
	logger  *slog.Logger
	db      Pinger
	pool    *pgxpool.Pool
	auth    *auth.Service
	records *record.Service
}

// NewServer は Server を組み立てる。logger が nil なら既定の slog を使う。
func NewServer(logger *slog.Logger, deps Deps) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		logger:  logger,
		db:      deps.DB,
		pool:    deps.Pool,
		auth:    deps.Auth,
		records: deps.Records,
	}
}

// Handler はルーティングと middleware を組み上げた http.Handler を返す。
// Go 1.22+ の method-pattern ServeMux を使い、フレームワークは入れない。
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// 認証不要: 稼働確認・アカウント発行。
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("POST /accounts", s.handleCreateAccount)

	// 認証必須: 文箱は本人だけ。
	mux.Handle("POST /records", s.requireAuth(http.HandlerFunc(s.handleCreateRecord)))
	mux.Handle("GET /box", s.requireAuth(http.HandlerFunc(s.handleListBox)))

	// middleware は外側から: recover -> log -> mux。
	var h http.Handler = mux
	h = logMiddleware(s.logger, h)
	h = recoverMiddleware(s.logger, h)
	return h
}
