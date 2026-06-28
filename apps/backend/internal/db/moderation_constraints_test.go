package db_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
)

// M3 関門スキーマの不変条件を実 DB の CHECK / UNIQUE で守れているか確かめる。
func TestModerationConstraints(t *testing.T) {
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
	reporter, _ := q.CreateUser(ctx)

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

	t.Run("tanzaku_origin は一枚につき高々一件(PK)", func(t *testing.T) {
		tid := insertTanzaku(t)
		if _, err := pool.Exec(ctx,
			`INSERT INTO tanzaku_origin (tanzaku_id, author_id) VALUES ($1, $2)`, tid, author.ID); err != nil {
			t.Fatalf("1 件目の origin: %v", err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO tanzaku_origin (tanzaku_id, author_id) VALUES ($1, $2)`, tid, author.ID); err == nil {
			t.Error("同じ tanzaku に origin を二重登録できてしまった")
		}
	})

	t.Run("pending_submission の status CHECK と attempts>=0", func(t *testing.T) {
		if _, err := pool.Exec(ctx,
			`INSERT INTO pending_submission (author_id, body, ko_written, status) VALUES ($1, $2, $3, $4)`,
			author.ID, "x", 29, "bogus"); err == nil {
			t.Error("status='bogus' が CHECK を通ってしまった")
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO pending_submission (author_id, body, ko_written, attempts) VALUES ($1, $2, $3, -1)`,
			author.ID, "x", 29); err == nil {
			t.Error("attempts=-1 が CHECK を通ってしまった")
		}
	})

	t.Run("pending_submission の ko_written CHECK", func(t *testing.T) {
		for _, ko := range []int{0, 73} {
			if _, err := pool.Exec(ctx,
				`INSERT INTO pending_submission (author_id, body, ko_written) VALUES ($1, $2, $3)`,
				author.ID, "x", ko); err == nil {
				t.Errorf("ko_written=%d が CHECK を通ってしまった", ko)
			}
		}
	})

	t.Run("report の reason CHECK と二重通報の禁止(UNIQUE)", func(t *testing.T) {
		tid := insertTanzaku(t)
		if _, err := pool.Exec(ctx,
			`INSERT INTO report (tanzaku_id, reporter_id, reason) VALUES ($1, $2, $3)`,
			tid, reporter.ID, "bogus"); err == nil {
			t.Error("reason='bogus' が CHECK を通ってしまった")
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO report (tanzaku_id, reporter_id, reason) VALUES ($1, $2, $3)`,
			tid, reporter.ID, "harassment"); err != nil {
			t.Fatalf("1 件目の通報: %v", err)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO report (tanzaku_id, reporter_id, reason) VALUES ($1, $2, $3)`,
			tid, reporter.ID, "spam"); err == nil {
			t.Error("同じ受け手が同じ一枚を二重通報できてしまった")
		}
	})

	t.Run("gate_verdict の subject_kind / verdict CHECK", func(t *testing.T) {
		if _, err := pool.Exec(ctx,
			`INSERT INTO gate_verdict (subject_kind, subject_id, verdict) VALUES ($1, $2, $3)`,
			"bogus", uuid.New(), "safe"); err == nil {
			t.Error("subject_kind='bogus' が CHECK を通ってしまった")
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO gate_verdict (subject_kind, subject_id, verdict) VALUES ($1, $2, $3)`,
			"tanzaku", uuid.New(), "bogus"); err == nil {
			t.Error("verdict='bogus' が CHECK を通ってしまった")
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO gate_verdict (subject_kind, subject_id, verdict) VALUES ($1, $2, $3)`,
			"tanzaku", uuid.New(), "child"); err != nil {
			t.Fatalf("正当な verdict が通らない: %v", err)
		}
	})
}
