package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
)

// 不変条件は DB の CHECK / UNIQUE で守る。これらはモックでは検出できないため実 DB で確かめる。
func TestConstraints(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	ctx := context.Background()
	pool, _, terminate, err := dbtest.MigratedPool(ctx)
	if err != nil {
		t.Fatalf("PG 起動失敗: %v", err)
	}
	defer terminate()

	q := sqlcdb.New(pool)
	owner, err := q.CreateUser(ctx)
	if err != nil {
		t.Fatalf("ユーザー作成: %v", err)
	}

	t.Run("record の ko_written CHECK(1..72)", func(t *testing.T) {
		// 0 と 73 は CHECK 違反で弾かれる。
		for _, ko := range []int{0, 73} {
			_, err := pool.Exec(ctx,
				`INSERT INTO record (owner_id, body, ko_written) VALUES ($1, $2, $3)`,
				owner.ID, "x", ko)
			if err == nil {
				t.Errorf("ko_written=%d が CHECK を通ってしまった", ko)
			}
		}
		// 1 と 72 は通る。
		for _, ko := range []int{1, 72} {
			if _, err := pool.Exec(ctx,
				`INSERT INTO record (owner_id, body, ko_written) VALUES ($1, $2, $3)`,
				owner.ID, "x", ko); err != nil {
				t.Errorf("ko_written=%d が通らない: %v", ko, err)
			}
		}
	})

	t.Run("auth_token の token_hash UNIQUE", func(t *testing.T) {
		hash := []byte("0123456789abcdef0123456789abcdef") // 32 bytes
		if _, err := q.CreateAuthToken(ctx, sqlcdb.CreateAuthTokenParams{UserID: owner.ID, TokenHash: hash}); err != nil {
			t.Fatalf("1 件目のトークン作成: %v", err)
		}
		if _, err := q.CreateAuthToken(ctx, sqlcdb.CreateAuthTokenParams{UserID: owner.ID, TokenHash: hash}); err == nil {
			t.Errorf("同じ token_hash の重複が許されてしまった")
		}
	})

	t.Run("失効トークンは GetActiveUserByTokenHash から除外", func(t *testing.T) {
		hash := []byte("ffffffffffffffffffffffffffffffff") // 32 bytes
		if _, err := q.CreateAuthToken(ctx, sqlcdb.CreateAuthTokenParams{UserID: owner.ID, TokenHash: hash}); err != nil {
			t.Fatalf("トークン作成: %v", err)
		}
		// 失効前は引ける。
		if _, err := q.GetActiveUserByTokenHash(ctx, hash); err != nil {
			t.Fatalf("失効前に引けない: %v", err)
		}
		// 失効させると引けない。
		if err := q.RevokeAuthToken(ctx, hash); err != nil {
			t.Fatalf("失効処理: %v", err)
		}
		if _, err := q.GetActiveUserByTokenHash(ctx, hash); !errors.Is(err, pgx.ErrNoRows) {
			t.Errorf("失効トークンが引けてしまった: err=%v", err)
		}
	})
}
