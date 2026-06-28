// Package auth は匿名アカウントと Bearer トークン認証を担う。
// アカウントに個人情報を紐づけない。トークンは平文を保存せず hash だけを持つ。
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
)

// ErrUnauthorized はトークンが無効・未知のときに返る。
var ErrUnauthorized = errors.New("認証に失敗")

// tokenBytes は Bearer トークンの乱数長（バイト）。
const tokenBytes = 32

// Principal は認証済みの呼び出し主体。
type Principal struct {
	UserID    uuid.UUID
	Role      string
	Suspended bool
}

// Service は匿名アカウントの発行と認証を提供する。
type Service struct {
	pool *pgxpool.Pool
}

// NewService は接続プールから Service を作る。
func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// CreateAccount は匿名アカウントを 1 つ作り、生のトークンを 1 度だけ返す。
// ユーザー行とトークン行をひとつのトランザクションで作る（中途半端な状態を残さない）。
func (s *Service) CreateAccount(ctx context.Context) (userID uuid.UUID, token string, err error) {
	token, hash, err := newToken()
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("トークン生成に失敗: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	u, err := q.CreateUser(ctx)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("アカウント作成に失敗: %w", err)
	}
	if _, err := q.CreateAuthToken(ctx, sqlcdb.CreateAuthTokenParams{UserID: u.ID, TokenHash: hash}); err != nil {
		return uuid.Nil, "", fmt.Errorf("トークン保存に失敗: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, "", fmt.Errorf("コミットに失敗: %w", err)
	}
	return u.ID, token, nil
}

// Authenticate は提示トークンから Principal を返す。未知・失効なら ErrUnauthorized。
func (s *Service) Authenticate(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrUnauthorized
	}
	hash := hashToken(token)
	q := sqlcdb.New(s.pool)
	row, err := q.GetActiveUserByTokenHash(ctx, hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Principal{}, ErrUnauthorized
	}
	if err != nil {
		return Principal{}, fmt.Errorf("トークン照合に失敗: %w", err)
	}
	// 最終利用時刻の更新は best-effort。失敗しても認証自体は通す。
	_ = q.TouchAuthToken(ctx, hash)
	return Principal{UserID: row.ID, Role: row.Role, Suspended: row.SuspendedAt.Valid}, nil
}

// newToken は乱数トークン（base64url 文字列）とその SHA-256 hash を返す。
func newToken() (token string, hash []byte, err error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

// hashToken は保存・照合に使うトークンの SHA-256 を返す。
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
