// Package record は記録・文箱のユースケースとドメインを担う。記録は本人だけがアクセスできる。
// 永続化は Repository ポートに委ね、この層は具体 DB を知らない（依存性ルール）。
package record

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// MaxBodyRunes は短冊本文の上限（短い言葉のための上限）。
const MaxBodyRunes = 280

// ErrInvalid は本文や候が不正なときに返る（呼び出し側で 400 に対応づける）。
var ErrInvalid = errors.New("記録の内容が不正")

// ErrNotFound は本人の記録が見つからないときに返る。
var ErrNotFound = errors.New("記録が見つからない")

// Record は文箱の一枚（ドメインの実体）。owner_id は本人のものなので応答には含めない。
type Record struct {
	ID        uuid.UUID
	KoWritten int
	Body      string
	CreatedAt time.Time
}

// Repository は記録の永続化ポート。owner-only は実装側が owner_id を必ず条件に入れて守る。
// Get は見つからなければ ErrNotFound を返す（具体 DB のエラーはアダプタが変換する）。
type Repository interface {
	Create(ctx context.Context, ownerID uuid.UUID, body string, ko int) (Record, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]Record, error)
	Get(ctx context.Context, id, ownerID uuid.UUID) (Record, error)
}

// Service は文箱の読み書きユースケース。入力検証を行い、永続化はポートに委ねる。
type Service struct {
	repo Repository
}

// NewService は Repository ポートから Service を作る。
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create は本文と候を検証してから記録を 1 枚保存する。
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
	return s.repo.Create(ctx, ownerID, body, ko)
}

// ListByOwner は本人の文箱を候別（昇順）・新しい順で返す。
func (s *Service) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]Record, error) {
	return s.repo.ListByOwner(ctx, ownerID)
}

// Get は本人の短冊を 1 枚返す。見つからなければ ErrNotFound。
func (s *Service) Get(ctx context.Context, id, ownerID uuid.UUID) (Record, error) {
	return s.repo.Get(ctx, id, ownerID)
}
