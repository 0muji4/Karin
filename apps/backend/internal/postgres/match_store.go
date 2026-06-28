package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
)

// MatchStore は exchange.MatchStore を満たす。日次マッチャの単一トランザクションを張る。
type MatchStore struct {
	pool *pgxpool.Pool
}

// NewMatchStore は接続プールから MatchStore を作る。
func NewMatchStore(pool *pgxpool.Pool) *MatchStore {
	return &MatchStore{pool: pool}
}

// RunDaily は 1 トランザクションを開き、トランザクション内の advisory lock で単一ライタを保証して fn を実行する。
// 二重起動した別プロセスはロック取得で待ち、先のコミット後に空のプールを見ることになる。
func (s *MatchStore) RunDaily(ctx context.Context, fn func(exchange.MatchTx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext('karin.matcher')::bigint)`); err != nil {
		return fmt.Errorf("マッチャのロック取得に失敗: %w", err)
	}
	if err := fn(&matchTx{q: sqlcdb.New(tx)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// matchTx は MatchTx をトランザクション上の sqlc クエリで満たす。
type matchTx struct {
	q *sqlcdb.Queries
}

func (m *matchTx) ExpirePooledBefore(ctx context.Context, cutoff time.Time) (int, error) {
	n, err := m.q.ExpirePooledBefore(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("候TTL の期限切れ処理に失敗: %w", err)
	}
	return int(n), nil
}

func (m *matchTx) EligibleRecipients(ctx context.Context, runDate time.Time) ([]exchange.Recipient, error) {
	rows, err := m.q.ListEligibleRecipients(ctx, dateOf(runDate))
	if err != nil {
		return nil, fmt.Errorf("受信資格者の取得に失敗: %w", err)
	}
	out := make([]exchange.Recipient, 0, len(rows))
	for _, r := range rows {
		out = append(out, exchange.Recipient{UserID: r.UserID, Assignable: int(r.Assignable)})
	}
	return out, nil
}

func (m *matchTx) PickOldestPooledFor(ctx context.Context, recipientID uuid.UUID) (uuid.UUID, bool, error) {
	id, err := m.q.PickOldestPooledForRecipient(ctx, pgtype.UUID{Bytes: recipientID, Valid: true})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("最古の未配信の取得に失敗: %w", err)
	}
	return id, true, nil
}

func (m *matchTx) Deliver(ctx context.Context, tanzakuID, recipientID uuid.UUID, runDate time.Time) error {
	if err := m.q.CreateDelivery(ctx, sqlcdb.CreateDeliveryParams{
		TanzakuID:   tanzakuID,
		RecipientID: recipientID,
		DeliveredOn: dateOf(runDate),
	}); err != nil {
		return fmt.Errorf("配信の記録に失敗: %w", err)
	}
	if err := m.q.MarkDelivered(ctx, tanzakuID); err != nil {
		return fmt.Errorf("配信済みへの更新に失敗: %w", err)
	}
	if err := m.q.DecrementCredit(ctx, recipientID); err != nil {
		return fmt.Errorf("クレジット減算に失敗: %w", err)
	}
	return nil
}

func dateOf(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, Valid: true}
}
