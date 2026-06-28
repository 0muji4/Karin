package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

// RecordRepo は record.Repository を record テーブルで満たす。
type RecordRepo struct {
	q *sqlcdb.Queries
}

// NewRecordRepo は接続から RecordRepo を作る。
func NewRecordRepo(db sqlcdb.DBTX) *RecordRepo {
	return &RecordRepo{q: sqlcdb.New(db)}
}

func (r *RecordRepo) Create(ctx context.Context, ownerID uuid.UUID, body string, ko int) (record.Record, error) {
	row, err := r.q.CreateRecord(ctx, sqlcdb.CreateRecordParams{
		OwnerID:   ownerID,
		Body:      body,
		KoWritten: int16(ko),
	})
	if err != nil {
		return record.Record{}, fmt.Errorf("記録の保存に失敗: %w", err)
	}
	return record.Record{ID: row.ID, KoWritten: int(row.KoWritten), Body: row.Body, CreatedAt: row.CreatedAt}, nil
}

func (r *RecordRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]record.Record, error) {
	rows, err := r.q.ListRecordsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("文箱の取得に失敗: %w", err)
	}
	out := make([]record.Record, 0, len(rows))
	for _, row := range rows {
		out = append(out, record.Record{ID: row.ID, KoWritten: int(row.KoWritten), Body: row.Body, CreatedAt: row.CreatedAt})
	}
	return out, nil
}

func (r *RecordRepo) Get(ctx context.Context, id, ownerID uuid.UUID) (record.Record, error) {
	row, err := r.q.GetRecord(ctx, sqlcdb.GetRecordParams{ID: id, OwnerID: ownerID})
	if errors.Is(err, pgx.ErrNoRows) {
		return record.Record{}, record.ErrNotFound
	}
	if err != nil {
		return record.Record{}, fmt.Errorf("記録の取得に失敗: %w", err)
	}
	return record.Record{ID: row.ID, KoWritten: int(row.KoWritten), Body: row.Body, CreatedAt: row.CreatedAt}, nil
}
