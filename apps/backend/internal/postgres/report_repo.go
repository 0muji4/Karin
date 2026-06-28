package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/report"
)

// gate_verdict.subject_kind の値（tanzaku は既存の一枚への判定）。
const subjectTanzaku = "tanzaku"

// ReportRepo は report.Store を満たす。著者は tanzaku_origin からのみ辿り、受け手向けには開かない。
type ReportRepo struct {
	pool *pgxpool.Pool
}

// NewReportRepo は接続プールから ReportRepo を作る。
func NewReportRepo(pool *pgxpool.Pool) *ReportRepo {
	return &ReportRepo{pool: pool}
}

// Submit は通報を記録し、再判定に要る本文と著者を返す。受信していなければ ErrNotReceived、
// 二重通報なら ErrAlreadyReported を返す。著者は運営専用クエリ(GetOriginForReview)からのみ得る。
func (r *ReportRepo) Submit(ctx context.Context, tanzakuID, reporterID uuid.UUID, reason, note string) (uuid.UUID, report.Subject, error) {
	// その一枚を実際に受け取った者だけが通報できる（任意の id への通報を防ぐ）。
	var one int
	err := r.pool.QueryRow(ctx,
		`SELECT 1 FROM delivery WHERE tanzaku_id = $1 AND recipient_id = $2`, tanzakuID, reporterID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, report.Subject{}, report.ErrNotReceived
	}
	if err != nil {
		return uuid.Nil, report.Subject{}, fmt.Errorf("受信確認に失敗: %w", err)
	}

	q := sqlcdb.New(r.pool)
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	reportID, err := q.CreateReport(ctx, sqlcdb.CreateReportParams{
		TanzakuID:  tanzakuID,
		ReporterID: reporterID,
		Reason:     reason,
		Note:       notePtr,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, report.Subject{}, report.ErrAlreadyReported
		}
		return uuid.Nil, report.Subject{}, fmt.Errorf("通報の記録に失敗: %w", err)
	}

	o, err := q.GetOriginForReview(ctx, tanzakuID)
	if err != nil {
		return uuid.Nil, report.Subject{}, fmt.Errorf("再判定対象の取得に失敗: %w", err)
	}
	return reportID, report.Subject{TanzakuID: tanzakuID, AuthorID: o.AuthorID, Body: o.Body}, nil
}

// Resolve は再判定の方針を不可分に反映する: 通報の決着・著者評判の増減・判定監査・（児童なら）保全。
func (r *ReportRepo) Resolve(ctx context.Context, out report.Outcome) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	res := string(out.Resolution)
	if err := q.ResolveReport(ctx, sqlcdb.ResolveReportParams{ID: out.ReportID, Resolution: &res}); err != nil {
		return fmt.Errorf("通報の決着に失敗: %w", err)
	}
	if out.ReputationDelta != 0 {
		if err := q.AdjustReputation(ctx, sqlcdb.AdjustReputationParams{
			ID:         out.Subject.AuthorID,
			Reputation: int32(out.ReputationDelta),
		}); err != nil {
			return fmt.Errorf("評判の更新に失敗: %w", err)
		}
	}
	if err := q.RecordGateVerdict(ctx, sqlcdb.RecordGateVerdictParams{
		SubjectKind: subjectTanzaku,
		SubjectID:   out.Subject.TanzakuID,
		Verdict:     out.Verdict.Label(),
		Raw:         causeJSON(out.Reason),
	}); err != nil {
		return fmt.Errorf("判定監査の記録に失敗: %w", err)
	}
	if out.ChildSafety {
		if err := q.CreateChildSafetyAlert(ctx, sqlcdb.CreateChildSafetyAlertParams{
			TanzakuID:      pgtype.UUID{Bytes: out.Subject.TanzakuID, Valid: true},
			AuthorID:       out.Subject.AuthorID,
			BodySnapshot:   out.Subject.Body,
			SourceReportID: pgtype.UUID{Bytes: out.ReportID, Valid: true},
		}); err != nil {
			return fmt.Errorf("児童保全ホールドの作成に失敗: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("コミットに失敗: %w", err)
	}
	return nil
}

// isUniqueViolation は UNIQUE 制約違反(23505)かどうかを判定する。
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
