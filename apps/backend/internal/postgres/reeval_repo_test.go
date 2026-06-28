package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
)

// 保留の再判定アダプタが昇格・却下・据え置きを正しく永続化することを実 DB で確かめる。
func TestReevalRepo_昇格と却下と据え置き(t *testing.T) {
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
	poolRepo := postgres.NewPoolRepo(pool)
	repo := postgres.NewReevalRepo(pool)

	// HoldForReview で awaiting の保留を 1 件作り、その PendingItem を返す。
	hold := func(t *testing.T, body string) (uuid.UUID, exchange.PendingItem) {
		t.Helper()
		author, _ := q.CreateUser(ctx)
		rec, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{OwnerID: author.ID, Body: body, KoWritten: 30})
		if err != nil {
			t.Fatalf("記録作成: %v", err)
		}
		in := exchange.CastInput{AuthorID: author.ID, SourceRecordID: rec.ID, Body: body, Ko: 30}
		if err := poolRepo.HoldForReview(ctx, in, "llm err"); err != nil {
			t.Fatalf("HoldForReview: %v", err)
		}
		var pid uuid.UUID
		if err := pool.QueryRow(ctx, `SELECT id FROM pending_submission WHERE author_id=$1`, author.ID).Scan(&pid); err != nil {
			t.Fatalf("保留 id 取得: %v", err)
		}
		return author.ID, exchange.PendingItem{ID: pid, AuthorID: author.ID, SourceRecordID: rec.ID, Body: body, Ko: 30}
	}

	t.Run("RunOnce(AllPass) は保留を昇格しプールへ出す", func(t *testing.T) {
		authorID, _ := hold(t, "昇格対象")
		if err := exchange.NewReevaluator(moderation.AllPass{}, repo).RunOnce(ctx, 100); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}
		var status string
		var promoted *uuid.UUID
		if err := pool.QueryRow(ctx,
			`SELECT status, promoted_tanzaku_id FROM pending_submission WHERE author_id=$1`, authorID).
			Scan(&status, &promoted); err != nil {
			t.Fatalf("保留の状態取得: %v", err)
		}
		if status != "pooled" || promoted == nil {
			t.Errorf("昇格されていない: status=%s promoted=%v", status, promoted)
		}
		var tanzaku int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM tanzaku WHERE author_id=$1 AND status='pooled'`, authorID).Scan(&tanzaku)
		if tanzaku != 1 {
			t.Errorf("プールへの投入数 = %d, want 1", tanzaku)
		}
		credits, _ := q.GetCredits(ctx, authorID)
		if credits != 1 {
			t.Errorf("クレジット = %d, want 1", credits)
		}
	})

	t.Run("Reject(児童) は却下し判定監査と保全を残す", func(t *testing.T) {
		authorID, item := hold(t, "却下対象")
		if err := repo.Reject(ctx, item, moderation.Child, "児童"); err != nil {
			t.Fatalf("Reject: %v", err)
		}
		var status string
		_ = pool.QueryRow(ctx, `SELECT status FROM pending_submission WHERE id=$1`, item.ID).Scan(&status)
		if status != "rejected" {
			t.Errorf("却下されていない: status=%s", status)
		}
		var verdicts, alerts int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM gate_verdict WHERE subject_id=$1 AND verdict='child'`, item.ID).Scan(&verdicts)
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM child_safety_alert WHERE author_id=$1`, authorID).Scan(&alerts)
		if verdicts != 1 || alerts != 1 {
			t.Errorf("判定監査/保全が不正: verdicts=%d alerts=%d", verdicts, alerts)
		}
	})

	t.Run("Defer は据え置き試行回数を進める", func(t *testing.T) {
		_, item := hold(t, "据え置き対象")
		if err := repo.Defer(ctx, item, "まだ確定しない"); err != nil {
			t.Fatalf("Defer: %v", err)
		}
		var status string
		var attempts int
		_ = pool.QueryRow(ctx, `SELECT status, attempts FROM pending_submission WHERE id=$1`, item.ID).Scan(&status, &attempts)
		if status != "awaiting" || attempts != 1 {
			t.Errorf("据え置きが不正: status=%s attempts=%d", status, attempts)
		}
	})
}
