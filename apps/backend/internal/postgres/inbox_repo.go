package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
)

// InboxRepo は exchange.Inbox を delivery / tanzaku / record テーブルで満たす。
type InboxRepo struct {
	pool *pgxpool.Pool
}

// NewInboxRepo は接続プールから InboxRepo を作る。
func NewInboxRepo(pool *pgxpool.Pool) *InboxRepo {
	return &InboxRepo{pool: pool}
}

// ListReceived は本人が受信した一枚を新しい順に返す（送り主は含めない）。
func (r *InboxRepo) ListReceived(ctx context.Context, recipientID uuid.UUID) ([]exchange.ReceivedCard, error) {
	rows, err := sqlcdb.New(r.pool).ListReceivedByRecipient(ctx, recipientID)
	if err != nil {
		return nil, fmt.Errorf("受信一覧の取得に失敗: %w", err)
	}
	out := make([]exchange.ReceivedCard, 0, len(rows))
	for _, row := range rows {
		out = append(out, exchange.ReceivedCard{
			TanzakuID:   row.ID,
			Body:        row.Body,
			Ko:          int(row.KoWritten),
			IsOfficial:  row.IsOfficial,
			DeliveredOn: row.DeliveredOn.Time,
			Kept:        row.KeptAt.Valid,
		})
	}
	return out, nil
}

// Keep は受信した一枚を文箱に複製してしまう。複製と「しまった」記録を 1 トランザクションで行う（冪等）。
func (r *InboxRepo) Keep(ctx context.Context, recipientID, tanzakuID uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	got, err := q.GetReceivedForKeep(ctx, sqlcdb.GetReceivedForKeepParams{TanzakuID: tanzakuID, RecipientID: recipientID})
	if errors.Is(err, pgx.ErrNoRows) {
		return exchange.ErrNotReceived
	}
	if err != nil {
		return fmt.Errorf("受信の確認に失敗: %w", err)
	}
	if got.KeptAt.Valid {
		return nil // 既にしまってある（冪等）。二重に文箱へ複製しない。
	}
	if _, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{
		OwnerID:   recipientID,
		Body:      got.Body,
		KoWritten: got.KoWritten,
	}); err != nil {
		return fmt.Errorf("文箱への複製に失敗: %w", err)
	}
	if err := q.MarkKept(ctx, sqlcdb.MarkKeptParams{TanzakuID: tanzakuID, RecipientID: recipientID}); err != nil {
		return fmt.Errorf("しまった記録の更新に失敗: %w", err)
	}
	return tx.Commit(ctx)
}
