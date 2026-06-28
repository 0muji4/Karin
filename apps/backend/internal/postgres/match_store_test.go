package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
)

// マッチャの M2 完了条件を実 DB で検証する（1 コンテナ・サブテスト間で TRUNCATE）。
func TestMatcher_Integration(t *testing.T) {
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
	now := time.Now()

	truncate := func() {
		if _, err := pool.Exec(ctx, `TRUNCATE tanzaku, delivery, exchange_ledger, users RESTART IDENTITY CASCADE`); err != nil {
			t.Fatalf("truncate: %v", err)
		}
	}
	mkUser := func() uuid.UUID {
		u, err := q.CreateUser(ctx)
		if err != nil {
			t.Fatalf("CreateUser: %v", err)
		}
		return u.ID
	}
	poolCard := func(author uuid.UUID, ko int, pooledAt time.Time) uuid.UUID {
		var id uuid.UUID
		if err := pool.QueryRow(ctx,
			`INSERT INTO tanzaku (author_id, body, ko_written, pooled_at) VALUES ($1,$2,$3,$4) RETURNING id`,
			author, "x", ko, pooledAt).Scan(&id); err != nil {
			t.Fatalf("poolCard: %v", err)
		}
		return id
	}
	setCredits := func(u uuid.UUID, n int) {
		if _, err := pool.Exec(ctx,
			`INSERT INTO exchange_ledger (user_id, receive_credits) VALUES ($1,$2)
			 ON CONFLICT (user_id) DO UPDATE SET receive_credits = EXCLUDED.receive_credits`, u, n); err != nil {
			t.Fatalf("setCredits: %v", err)
		}
	}
	run := func(ttlKo int) {
		if err := exchange.NewMatcher(postgres.NewMatchStore(pool), ttlKo).RunDaily(ctx); err != nil {
			t.Fatalf("RunDaily: %v", err)
		}
	}
	scanInt := func(sql string, args ...any) int {
		var n int
		if err := pool.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
			t.Fatalf("query %q: %v", sql, err)
		}
		return n
	}
	creditsOf := func(u uuid.UUID) int {
		return scanInt(`SELECT receive_credits FROM exchange_ledger WHERE user_id=$1`, u)
	}
	deliveriesTo := func(u uuid.UUID) int { return scanInt(`SELECT count(*) FROM delivery WHERE recipient_id=$1`, u) }
	deliveriesOf := func(id uuid.UUID) int { return scanInt(`SELECT count(*) FROM delivery WHERE tanzaku_id=$1`, id) }
	statusOf := func(id uuid.UUID) string {
		var s string
		if err := pool.QueryRow(ctx, `SELECT status FROM tanzaku WHERE id=$1`, id).Scan(&s); err != nil {
			t.Fatalf("statusOf: %v", err)
		}
		return s
	}
	authorIsNull := func(id uuid.UUID) bool {
		var isNull bool
		if err := pool.QueryRow(ctx, `SELECT author_id IS NULL FROM tanzaku WHERE id=$1`, id).Scan(&isNull); err != nil {
			t.Fatalf("authorIsNull: %v", err)
		}
		return isNull
	}

	t.Run("自分の一枚は返らない・最古から・著者剥離・受信で-1", func(t *testing.T) {
		truncate()
		u1, u2 := mkUser(), mkUser()
		setCredits(u1, 1)
		own := poolCard(u1, 11, now.Add(-3*time.Hour))    // u1 自作（最古だが対象外）
		oldest := poolCard(u2, 12, now.Add(-2*time.Hour)) // 他者作の最古
		poolCard(u2, 13, now.Add(-1*time.Hour))           // 他者作の新しい方
		run(6)

		if got := deliveriesTo(u1); got != 1 {
			t.Fatalf("u1 への配信 = %d, want 1", got)
		}
		if deliveriesOf(own) != 0 {
			t.Error("自分の一枚が自分に返った")
		}
		if deliveriesOf(oldest) != 1 {
			t.Error("最古（他者作）が配信されていない")
		}
		if !authorIsNull(oldest) {
			t.Error("配信後に author_id が剥がれていない")
		}
		if statusOf(oldest) != "delivered" {
			t.Errorf("status = %s, want delivered", statusOf(oldest))
		}
		if got := creditsOf(u1); got != 0 {
			t.Errorf("受信後の u1 クレジット = %d, want 0", got)
		}
	})

	t.Run("二重配信しない・1日1回", func(t *testing.T) {
		truncate()
		u1, u2, u3 := mkUser(), mkUser(), mkUser()
		setCredits(u1, 1)
		setCredits(u2, 1)
		cB := poolCard(u3, 11, now.Add(-time.Hour))
		run(6)

		if got := deliveriesOf(cB); got != 1 {
			t.Errorf("cB の配信数 = %d, want 1（二重配信）", got)
		}
		if got := deliveriesTo(u1) + deliveriesTo(u2); got != 1 {
			t.Errorf("配信総数 = %d, want 1", got)
		}
	})

	t.Run("繰越（クレジットは残る・1日1回）", func(t *testing.T) {
		truncate()
		u1, u2 := mkUser(), mkUser()
		setCredits(u1, 2)
		poolCard(u2, 11, now.Add(-2*time.Hour))
		poolCard(u2, 12, now.Add(-time.Hour))
		run(6)

		if got := deliveriesTo(u1); got != 1 {
			t.Errorf("u1 への配信 = %d, want 1（1日1回）", got)
		}
		if got := creditsOf(u1); got != 1 {
			t.Errorf("受信後の u1 クレジット = %d, want 1（繰越）", got)
		}
	})

	t.Run("候TTL 超過は expired・配信されない", func(t *testing.T) {
		truncate()
		u1, u2 := mkUser(), mkUser()
		setCredits(u1, 1)
		stale := poolCard(u2, 11, now.AddDate(0, 0, -60)) // 60 日前（TTL ~30 日超過）
		run(6)

		if got := statusOf(stale); got != "expired" {
			t.Errorf("status = %s, want expired", got)
		}
		if deliveriesTo(u1) != 0 {
			t.Error("期限切れの一枚が配信された")
		}
	})
}
