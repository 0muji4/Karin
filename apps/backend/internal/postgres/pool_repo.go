package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
)

// gate_verdict.subject_kind の値（DB の CHECK と一致）。保留や短冊など判定対象の種別を表す。
const subjectPending = "pending"

// gate_verdict.verdict の値（DB の CHECK と一致）。
const verdictLLMError = "llm_error"

// PoolRepo は exchange.GateEffects を満たす。安全な一枚の投入・著者特定リンク・クレジット +1 を
// 1 トランザクションで不可分に行う。
type PoolRepo struct {
	pool *pgxpool.Pool
}

// NewPoolRepo は接続プールから PoolRepo を作る（トランザクション開始にプールが要る）。
func NewPoolRepo(pool *pgxpool.Pool) *PoolRepo {
	return &PoolRepo{pool: pool}
}

// PoolSafe は安全な一枚を pooled で投入し、著者特定の私的リンク(origin)を作り、
// 著者のクレジットを +1 する（不可分）。これにより配信で author_id を NULL 化しても
// origin から著者を辿れる（通報→評判・児童保全の根拠）。
// クレジットを投入時（＝関門通過時）に付けることで、害のある投稿で受信権を稼げないようにする。
func (r *PoolRepo) PoolSafe(ctx context.Context, in exchange.CastInput) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	row, err := q.PoolTanzaku(ctx, sqlcdb.PoolTanzakuParams{
		AuthorID:   pgtype.UUID{Bytes: in.AuthorID, Valid: true},
		Body:       in.Body,
		KoWritten:  int16(in.Ko),
		IsOfficial: false,
	})
	if err != nil {
		return fmt.Errorf("プール投入に失敗: %w", err)
	}
	if err := q.CreateTanzakuOrigin(ctx, sqlcdb.CreateTanzakuOriginParams{
		TanzakuID:      row.ID,
		AuthorID:       in.AuthorID,
		SourceRecordID: pgtype.UUID{Bytes: in.SourceRecordID, Valid: true},
	}); err != nil {
		return fmt.Errorf("著者リンクの作成に失敗: %w", err)
	}
	if err := q.IncrementCredit(ctx, in.AuthorID); err != nil {
		return fmt.Errorf("クレジット加算に失敗: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミットに失敗: %w", err)
	}
	return nil
}

// HoldForReview は判定が確定しなかった一枚を awaiting で保留し、判定監査(llm_error)を残す（不可分）。
// 配信はせず、クレジットも付けない。本文はスナップショットとして pending_submission に持つ。
func (r *PoolRepo) HoldForReview(ctx context.Context, in exchange.CastInput, cause string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	pendingID, err := q.CreatePendingSubmission(ctx, sqlcdb.CreatePendingSubmissionParams{
		AuthorID:       in.AuthorID,
		SourceRecordID: pgtype.UUID{Bytes: in.SourceRecordID, Valid: true},
		Body:           in.Body,
		KoWritten:      int16(in.Ko),
		LastError:      &cause,
	})
	if err != nil {
		return fmt.Errorf("保留の作成に失敗: %w", err)
	}
	if err := q.RecordGateVerdict(ctx, sqlcdb.RecordGateVerdictParams{
		SubjectKind: subjectPending,
		SubjectID:   pendingID,
		Verdict:     verdictLLMError,
		Raw:         causeJSON(cause),
	}); err != nil {
		return fmt.Errorf("判定監査の記録に失敗: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミットに失敗: %w", err)
	}
	return nil
}

// causeJSON は保留原因を gate_verdict.raw 用の JSON にする。失敗しても監査の本筋は止めない。
func causeJSON(cause string) []byte {
	b, err := json.Marshal(map[string]string{"cause": cause})
	if err != nil {
		return nil
	}
	return b
}
