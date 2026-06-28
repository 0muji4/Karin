// Command server はクライアント向けの JSON HTTP API を提供する。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/api"
	"github.com/0muji4/Karin/apps/backend/internal/auth"
	"github.com/0muji4/Karin/apps/backend/internal/config"
	"github.com/0muji4/Karin/apps/backend/internal/db"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("サーバ起動に失敗", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}

	// シグナルで打ち切れる起動コンテキスト。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	apiServer := api.NewServer(logger, api.Deps{
		DB:      pool,
		Ko:      postgres.NewKoCatalog(pool),
		Auth:    auth.NewService(postgres.NewAuthRepo(pool)),
		Records: record.NewService(postgres.NewRecordRepo(pool)),
		Cast:    exchange.NewCastService(postgres.NewRecordRepo(pool), moderation.AllPass{}, postgres.NewPoolRepo(pool)),
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           apiServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second, // Slowloris 対策。ゼロ値の http.Server は無防備。
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// サーバを別ゴルーチンで動かし、シグナルで graceful shutdown する。
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP サーバ起動", "addr", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		logger.Info("終了シグナルを受信。graceful shutdown 開始")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
