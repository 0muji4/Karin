package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
)

// ReevalRepo は exchange.ReevalStore を満たす。保留の再判定結果（昇格・却下・据え置き）を永続化する。
type ReevalRepo struct {
	pool *pgxpool.Pool
}

// NewReevalRepo は接続プールから ReevalRepo を作る。
func NewReevalRepo(pool *pgxpool.Pool) *ReevalRepo {
	return &ReevalRepo{pool: pool}
}

// ListAwaiting は再判定待ちの保留を古い順に最大 limit 件返す。
func (r *ReevalRepo) ListAwaiting(ctx context.Context, limit int) ([]exchange.PendingItem, error) {
	rows, err := sqlcdb.New(r.pool).ListAwaitingSubmissions(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("保留の取得に失敗: %w", err)
	}
	items := make([]exchange.PendingItem, 0, len(rows))
	for _, row := range rows {
		var src uuid.UUID
		if row.SourceRecordID.Valid {
			src = uuid.UUID(row.SourceRecordID.Bytes)
		}
		items = append(items, exchange.PendingItem{
			ID:             row.ID,
			AuthorID:       row.AuthorID,
			SourceRecordID: src,
			Body:           row.Body,
			Ko:             int(row.KoWritten),
		})
	}
	return items, nil
}

// Promote は保留を pooled に昇格する: tanzaku+origin+クレジット +1 を作り、保留を pooled にする（不可分）。
func (r *ReevalRepo) Promote(ctx context.Context, item exchange.PendingItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	row, err := q.PoolTanzaku(ctx, sqlcdb.PoolTanzakuParams{
		AuthorID:  pgtype.UUID{Bytes: item.AuthorID, Valid: true},
		Body:      item.Body,
		KoWritten: int16(item.Ko),
	})
	if err != nil {
		return fmt.Errorf("プール投入に失敗: %w", err)
	}
	if err := q.CreateTanzakuOrigin(ctx, sqlcdb.CreateTanzakuOriginParams{
		TanzakuID:      row.ID,
		AuthorID:       item.AuthorID,
		SourceRecordID: pgtype.UUID{Bytes: item.SourceRecordID, Valid: item.SourceRecordID != uuid.Nil},
	}); err != nil {
		return fmt.Errorf("著者リンクの作成に失敗: %w", err)
	}
	if err := q.IncrementCredit(ctx, item.AuthorID); err != nil {
		return fmt.Errorf("クレジット加算に失敗: %w", err)
	}
	if err := q.MarkPendingPooled(ctx, sqlcdb.MarkPendingPooledParams{
		ID:                item.ID,
		PromotedTanzakuID: pgtype.UUID{Bytes: row.ID, Valid: true},
	}); err != nil {
		return fmt.Errorf("保留の昇格に失敗: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミットに失敗: %w", err)
	}
	return nil
}

// Reject は保留を rejected にし判定監査を残す（児童なら保全ホールドも）。不可分。
func (r *ReevalRepo) Reject(ctx context.Context, item exchange.PendingItem, v moderation.Verdict, reason string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	if err := q.MarkPendingRejected(ctx, sqlcdb.MarkPendingRejectedParams{ID: item.ID, LastError: &reason}); err != nil {
		return fmt.Errorf("却下への更新に失敗: %w", err)
	}
	if err := q.RecordGateVerdict(ctx, sqlcdb.RecordGateVerdictParams{
		SubjectKind: subjectPending,
		SubjectID:   item.ID,
		Verdict:     v.Label(),
		Raw:         causeJSON(reason),
	}); err != nil {
		return fmt.Errorf("判定監査の記録に失敗: %w", err)
	}
	if v == moderation.Child {
		if err := q.CreateChildSafetyAlert(ctx, sqlcdb.CreateChildSafetyAlertParams{
			AuthorID:     item.AuthorID,
			BodySnapshot: item.Body,
		}); err != nil {
			return fmt.Errorf("児童保全ホールドの作成に失敗: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミットに失敗: %w", err)
	}
	return nil
}

// Defer は保留を据え置く（試行回数を進め原因を残す）。次回また再判定される。
func (r *ReevalRepo) Defer(ctx context.Context, item exchange.PendingItem, cause string) error {
	if err := sqlcdb.New(r.pool).IncrementPendingAttempt(ctx, sqlcdb.IncrementPendingAttemptParams{
		ID:        item.ID,
		LastError: &cause,
	}); err != nil {
		return fmt.Errorf("保留の据え置きに失敗: %w", err)
	}
	return nil
}
