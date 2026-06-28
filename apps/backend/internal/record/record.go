// Package record は記録・文箱を担う。記録は本人だけがアクセスできる。
// owner-only は全クエリで owner_id を必ず条件に入れて守る（DB は行単位認可を持たない）。
package record

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
)

// MaxBodyRunes は短冊本文の上限（短い言葉のための上限）。
const MaxBodyRunes = 280

// ErrInvalid は本文や候が不正なときに返る（呼び出し側で 400 に対応づける）。
var ErrInvalid = errors.New("記録の内容が不正")

// ErrNotFound は本人の記録が見つからないときに返る。
var ErrNotFound = errors.New("記録が見つからない")

// Record は文箱の一枚（API 応答の素）。owner_id は本人のものなので応答には含めない。
type Record struct {
	ID        uuid.UUID
	KoWritten int
	Body      string
	CreatedAt time.Time
}

// Service は文箱の読み書きを提供する。
type Service struct {
	pool *pgxpool.Pool
}

// NewService は接続プールから Service を作る。
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// Create は本人の文箱に短冊を 1 枚保存する。
func (s *Service) Create(ctx context.Context, ownerID uuid.UUID, body string, ko int) (Record, error) {
	body = strings.TrimSpace(body)
	if body == "" {
		return Record{}, fmt.Errorf("%w: 本文が空", ErrInvalid)
	}
	if utf8.RuneCountInString(body) > MaxBodyRunes {
		return Record{}, fmt.Errorf("%w: 本文が長すぎる（最大 %d 文字）", ErrInvalid, MaxBodyRunes)
	}
	if ko < 1 || ko > 72 {
		return Record{}, fmt.Errorf("%w: 候は 1〜72", ErrInvalid)
	}

	q := sqlcdb.New(s.pool)
	row, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{
		OwnerID:   ownerID,
		Body:      body,
		KoWritten: int16(ko),
	})
	if err != nil {
		return Record{}, fmt.Errorf("記録の保存に失敗: %w", err)
	}
	return Record{ID: row.ID, KoWritten: int(row.KoWritten), Body: row.Body, CreatedAt: row.CreatedAt}, nil
}

// ListByOwner は本人の文箱を候別（昇順）・新しい順で返す。
func (s *Service) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]Record, error) {
	q := sqlcdb.New(s.pool)
	rows, err := q.ListRecordsByOwner(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf("文箱の取得に失敗: %w", err)
	}
	out := make([]Record, 0, len(rows))
	for _, r := range rows {
		out = append(out, Record{ID: r.ID, KoWritten: int(r.KoWritten), Body: r.Body, CreatedAt: r.CreatedAt})
	}
	return out, nil
}

// Get は本人の短冊を 1 枚返す（owner_id で必ず絞る）。見つからなければ ErrNotFound。
func (s *Service) Get(ctx context.Context, id, ownerID uuid.UUID) (Record, error) {
	q := sqlcdb.New(s.pool)
	row, err := q.GetRecord(ctx, sqlcdb.GetRecordParams{ID: id, OwnerID: ownerID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, ErrNotFound
	}
	if err != nil {
		return Record{}, fmt.Errorf("記録の取得に失敗: %w", err)
	}
	return Record{ID: row.ID, KoWritten: int(row.KoWritten), Body: row.Body, CreatedAt: row.CreatedAt}, nil
}
