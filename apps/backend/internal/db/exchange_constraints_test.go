package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
)

// 交換コアの不変条件を DB の CHECK / UNIQUE で守れているか実 DB で確かめる。
func TestExchangeConstraints(t *testing.T) {
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
	author, _ := q.CreateUser(ctx)
	rcpt, _ := q.CreateUser(ctx)
	rcpt2, _ := q.CreateUser(ctx)
	day := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)

	insertTanzaku := func(t *testing.T) uuid.UUID {
		t.Helper()
		var id uuid.UUID
		if err := pool.QueryRow(ctx,
			`INSERT INTO tanzaku (author_id, body, ko_written) VALUES ($1, $2, $3) RETURNING id`,
			author.ID, "x", 29).Scan(&id); err != nil {
			t.Fatalf("tanzaku 作成: %v", err)
		}
		return id
	}

	t.Run("tanzaku status CHECK", func(t *testing.T) {
		_, err := pool.Exec(ctx,
			`INSERT INTO tanzaku (author_id, body, ko_written, status) VALUES ($1, $2, $3, $4)`,
			author.ID, "x", 29, "bogus")
		if err == nil {
			t.Error("status='bogus' が CHECK を通ってしまった")
		}
	})

	t.Run("tanzaku ko_written CHECK", func(t *testing.T) {
		for _, ko := range []int{0, 73} {
			if _, err := pool.Exec(ctx,
				`INSERT INTO tanzaku (author_id, body, ko_written) VALUES ($1, $2, $3)`,
				author.ID, "x", ko); err == nil {
				t.Errorf("ko_written=%d が CHECK を通ってしまった", ko)
			}
		}
	})

	t.Run("delivery のユニキャスト(tanzaku_id UNIQUE)と1日1回(recipient_id,delivered_on UNIQUE)", func(t *testing.T) {
		t1 := insertTanzaku(t)
		// 1 件目の配信は通る。
		if _, err := pool.Exec(ctx,
			`INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on) VALUES ($1, $2, $3)`,
			t1, rcpt.ID, day); err != nil {
			t.Fatalf("1 件目の配信: %v", err)
		}
		// 同じ tanzaku を別の受け手へ → ユニキャスト違反。
		if _, err := pool.Exec(ctx,
			`INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on) VALUES ($1, $2, $3)`,
			t1, rcpt2.ID, day); err == nil {
			t.Error("同じ tanzaku の二重配信が許された（ユニキャスト違反）")
		}
		// 同じ受け手へ同じ日に別の tanzaku → 1日1回違反。
		t2 := insertTanzaku(t)
		if _, err := pool.Exec(ctx,
			`INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on) VALUES ($1, $2, $3)`,
			t2, rcpt.ID, day); err == nil {
			t.Error("同じ受け手へ同じ日に 2 回配信が許された（1日1回違反）")
		}
	})

	t.Run("exchange_ledger の receive_credits >= 0", func(t *testing.T) {
		if _, err := pool.Exec(ctx,
			`INSERT INTO exchange_ledger (user_id, receive_credits) VALUES ($1, 0)`, author.ID); err != nil {
			t.Fatalf("credits=0 が通らない: %v", err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO exchange_ledger (user_id, receive_credits) VALUES ($1, -1)`, rcpt.ID); err == nil {
			t.Error("receive_credits=-1 が CHECK を通ってしまった")
		}
	})
}
