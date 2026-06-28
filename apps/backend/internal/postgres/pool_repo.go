package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
)

// PoolRepo は exchange.Pool を満たす。短冊の投入と著者クレジット +1 を 1 トランザクションで行う。
type PoolRepo struct {
	pool *pgxpool.Pool
}

// NewPoolRepo は接続プールから PoolRepo を作る（トランザクション開始にプールが要る）。
func NewPoolRepo(pool *pgxpool.Pool) *PoolRepo {
	return &PoolRepo{pool: pool}
}

// Pool は短冊を pooled で投入し、著者のクレジットを +1 する（不可分）。
// クレジットを投入時（＝関門通過時）に付けることで、害のある投稿で受信権を稼げないようにする。
func (r *PoolRepo) Pool(ctx context.Context, authorID uuid.UUID, body string, ko int, isOfficial bool) (exchange.Tanzaku, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return exchange.Tanzaku{}, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	row, err := q.PoolTanzaku(ctx, sqlcdb.PoolTanzakuParams{
		AuthorID:   pgtype.UUID{Bytes: authorID, Valid: true},
		Body:       body,
		KoWritten:  int16(ko),
		IsOfficial: isOfficial,
	})
	if err != nil {
		return exchange.Tanzaku{}, fmt.Errorf("プール投入に失敗: %w", err)
	}
	if err := q.IncrementCredit(ctx, authorID); err != nil {
		return exchange.Tanzaku{}, fmt.Errorf("クレジット加算に失敗: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return exchange.Tanzaku{}, fmt.Errorf("コミットに失敗: %w", err)
	}
	return toExchangeTanzaku(row), nil
}

func toExchangeTanzaku(t sqlcdb.Tanzaku) exchange.Tanzaku {
	var author uuid.UUID
	if t.AuthorID.Valid {
		author = uuid.UUID(t.AuthorID.Bytes)
	}
	return exchange.Tanzaku{
		ID:         t.ID,
		AuthorID:   author,
		Body:       t.Body,
		Ko:         int(t.KoWritten),
		Status:     exchange.Status(t.Status),
		PooledAt:   t.PooledAt,
		IsOfficial: t.IsOfficial,
	}
}
