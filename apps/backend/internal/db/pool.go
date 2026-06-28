// Package db は PostgreSQL への接続プールとマイグレーション適用を担う。
// 状態はすべて DB が持ち、アプリケーションサーバは状態を持たない（入れ替え可能）。
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool は接続文字列から pgxpool を作り、疎通を確認して返す。
// 呼び出し側は使い終わりに pool.Close() する。
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("接続文字列の解析に失敗: %w", err)
	}
	// 低トラフィック前提の控えめな既定値。運用で調整する。
	cfg.MaxConns = 10
	cfg.MinConns = 0
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("接続プールの作成に失敗: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("DB への疎通確認に失敗: %w", err)
	}
	return pool, nil
}
