// Command matcher は日次の交換マッチングを 1 回実行して終了する（cron から起動する前提）。
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/0muji4/Karin/apps/backend/internal/config"
	"github.com/0muji4/Karin/apps/backend/internal/db"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
)

// reevalBatch は 1 回の起動で再判定する保留の上限。低トラフィック前提の安全な既定値。
const reevalBatch = 200

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

	// 先に保留(fail-closed で留め置かれた一枚)を再判定し、安全なら配信対象に昇格させる。
	// LLM 呼び出しを伴うため、配信の単一ライタ・トランザクションの外で行う。関門は設定で差し替える（当面 AllPass）。
	reeval := exchange.NewReevaluator(moderation.AllPass{}, postgres.NewReevalRepo(pool))
	if err := reeval.RunOnce(ctx, reevalBatch); err != nil {
		return err
	}

	matcher := exchange.NewMatcher(postgres.NewMatchStore(pool), cfg.KoTTL)
	if err := matcher.RunDaily(ctx); err != nil {
		return err
	}
	logger.Info("マッチング完了")
	return nil
}
