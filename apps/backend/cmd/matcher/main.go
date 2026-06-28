// Command matcher は日次の交換マッチングを 1 回実行して終了する（cron から起動する前提）。
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/0muji4/Karin/apps/backend/internal/config"
	"github.com/0muji4/Karin/apps/backend/internal/db"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(logger); err != nil {
		logger.Error("マッチング実行に失敗", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		return err
	}
	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	matcher := exchange.NewMatcher(postgres.NewMatchStore(pool), cfg.KoTTL)
	if err := matcher.RunDaily(ctx); err != nil {
		return err
	}
	logger.Info("マッチング完了")
	return nil
}
