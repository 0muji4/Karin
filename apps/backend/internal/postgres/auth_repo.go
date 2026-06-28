package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/0muji4/Karin/apps/backend/internal/auth"
	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
)

// AuthRepo は auth.Repository を users / auth_token テーブルで満たす。
// アカウント作成とトークン保存のトランザクションはこの層が持つ。
type AuthRepo struct {
	pool *pgxpool.Pool
}

// NewAuthRepo は接続プールから AuthRepo を作る（トランザクション開始にプールが要る）。
func NewAuthRepo(pool *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{pool: pool}
}

// CreateAccountWithToken はユーザー行とトークン行をひとつのトランザクションで作る。
func (r *AuthRepo) CreateAccountWithToken(ctx context.Context, tokenHash []byte) (uuid.UUID, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("トランザクション開始に失敗: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // Commit 済みなら no-op

	q := sqlcdb.New(tx)
	u, err := q.CreateUser(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("アカウント作成に失敗: %w", err)
	}
	if _, err := q.CreateAuthToken(ctx, sqlcdb.CreateAuthTokenParams{UserID: u.ID, TokenHash: tokenHash}); err != nil {
		return uuid.Nil, fmt.Errorf("トークン保存に失敗: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("コミットに失敗: %w", err)
	}
	return u.ID, nil
}

// FindByActiveToken は失効していないトークンの所有者を返す。無ければ ErrUnauthorized。
func (r *AuthRepo) FindByActiveToken(ctx context.Context, tokenHash []byte) (auth.Principal, error) {
	row, err := sqlcdb.New(r.pool).GetActiveUserByTokenHash(ctx, tokenHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return auth.Principal{}, auth.ErrUnauthorized
	}
	if err != nil {
		return auth.Principal{}, fmt.Errorf("トークン照合に失敗: %w", err)
	}
	return auth.Principal{UserID: row.ID, Role: row.Role, Suspended: row.SuspendedAt.Valid}, nil
}

// TouchToken は最終利用時刻を更新する。
func (r *AuthRepo) TouchToken(ctx context.Context, tokenHash []byte) error {
	return sqlcdb.New(r.pool).TouchAuthToken(ctx, tokenHash)
}
