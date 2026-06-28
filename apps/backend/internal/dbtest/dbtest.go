// Package dbtest は結合テスト用に本物の PostgreSQL 18 を起動する。
//
// 設計文書の教訓: DB の CHECK / UNIQUE / NOT NULL・SKIP LOCKED・advisory lock は
// DB をモックした unit test では検出できない。状態遷移と制約は本物の DB で検証する。
// パッケージごとに TestMain でコンテナを 1 つ起動し、テスト間で使い回す。
package dbtest

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	appdb "github.com/0muji4/Karin/apps/backend/internal/db"
)

// Image は local / CI / 本番に揃える PostgreSQL のバージョン。
const Image = "postgres:18"

// RunPostgres は素の PG18 を起動し、接続 URL と後始末関数を返す（マイグレーション未適用）。
// マイグレーション自体を検証するテスト向け。
func RunPostgres(ctx context.Context) (url string, terminate func(), err error) {
	container, err := postgres.Run(ctx, Image,
		postgres.WithDatabase("karin_test"),
		postgres.WithUsername("karin"),
		postgres.WithPassword("karin"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		return "", nil, fmt.Errorf("PostgreSQL コンテナの起動に失敗: %w", err)
	}
	stop := func() { _ = container.Terminate(context.Background()) }

	url, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		stop()
		return "", nil, fmt.Errorf("接続文字列の取得に失敗: %w", err)
	}
	return url, stop, nil
}

// MigratedPool は PG18 を起動し、全マイグレーションを適用した接続プールを返す。
// ほとんどの結合テストはこちらを使う。terminate はプールを閉じてコンテナを止める。
func MigratedPool(ctx context.Context) (pool *pgxpool.Pool, url string, terminate func(), err error) {
	url, stop, err := RunPostgres(ctx)
	if err != nil {
		return nil, "", nil, err
	}
	if err := appdb.MigrateUp(url); err != nil {
		stop()
		return nil, "", nil, fmt.Errorf("マイグレーション適用に失敗: %w", err)
	}
	pool, err = appdb.NewPool(ctx, url)
	if err != nil {
		stop()
		return nil, "", nil, err
	}
	return pool, url, func() { pool.Close(); stop() }, nil
}
