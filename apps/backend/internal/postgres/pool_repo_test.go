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

// PoolSafe は安全な一枚の投入・著者特定リンク(origin)・クレジット +1 を不可分に書く。
// 配信で author_id を NULL 化しても origin から著者を辿れる不変条件を実 DB で確かめる。
func TestPoolSafe_書き込みは不可分(t *testing.T) {
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
	rec, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{OwnerID: author.ID, Body: "蝉の声", KoWritten: 33})
	if err != nil {
		t.Fatalf("記録作成: %v", err)
	}

	repo := postgres.NewPoolRepo(pool)
	in := exchange.CastInput{AuthorID: author.ID, SourceRecordID: rec.ID, Body: "蝉の声", Ko: 33}
	if err := repo.PoolSafe(ctx, in); err != nil {
		t.Fatalf("PoolSafe: %v", err)
	}

	// 1) tanzaku が pooled で投入されている。
	var tanzakuID uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT id FROM tanzaku WHERE author_id = $1 AND status = 'pooled'`, author.ID).Scan(&tanzakuID); err != nil {
		t.Fatalf("投入された tanzaku が見つからない: %v", err)
	}

	// 2) origin が同じ tanzaku に対して作られ、著者と由来を保持している。
	var gotAuthor, gotSource uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT author_id, source_record_id FROM tanzaku_origin WHERE tanzaku_id = $1`, tanzakuID).
		Scan(&gotAuthor, &gotSource); err != nil {
		t.Fatalf("origin が作られていない（不可分性が崩れている）: %v", err)
	}
	if gotAuthor != author.ID || gotSource != rec.ID {
		t.Errorf("origin の中身が不正: author=%v source=%v", gotAuthor, gotSource)
	}

	// 3) 著者のクレジットが +1 されている。
	credits, err := q.GetCredits(ctx, author.ID)
	if err != nil {
		t.Fatalf("クレジット取得: %v", err)
	}
	if credits != 1 {
		t.Errorf("クレジット = %d, want 1", credits)
	}
}

// HoldForReview は判定が確定しない一枚を awaiting で保留し、llm_error の判定監査を不可分に残す。
// 配信せず、クレジットも付けない（fail-closed）。
func TestHoldForReview_保留と監査(t *testing.T) {
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
	rec, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{OwnerID: author.ID, Body: "判定保留の本文", KoWritten: 33})
	if err != nil {
		t.Fatalf("記録作成: %v", err)
	}

	repo := postgres.NewPoolRepo(pool)
	in := exchange.CastInput{AuthorID: author.ID, SourceRecordID: rec.ID, Body: "判定保留の本文", Ko: 33}
	if err := repo.HoldForReview(ctx, in, "llm timeout"); err != nil {
		t.Fatalf("HoldForReview: %v", err)
	}

	// 1) pending_submission が awaiting・スナップショット本文・原因を保持して作られる。
	var pendingID uuid.UUID
	var status, body string
	var lastErr *string
	if err := pool.QueryRow(ctx,
		`SELECT id, status, body, last_error FROM pending_submission WHERE author_id = $1`, author.ID).
		Scan(&pendingID, &status, &body, &lastErr); err != nil {
		t.Fatalf("保留が作られていない: %v", err)
	}
	if status != "awaiting" || body != "判定保留の本文" || lastErr == nil || *lastErr != "llm timeout" {
		t.Errorf("保留の中身が不正: status=%s body=%q lastErr=%v", status, body, lastErr)
	}

	// 2) gate_verdict が pending/llm_error として、同 tx で残る。
	var verdict, kind string
	if err := pool.QueryRow(ctx,
		`SELECT verdict, subject_kind FROM gate_verdict WHERE subject_id = $1`, pendingID).
		Scan(&verdict, &kind); err != nil {
		t.Fatalf("判定監査が残っていない（不可分性が崩れている）: %v", err)
	}
	if verdict != "llm_error" || kind != "pending" {
		t.Errorf("判定監査が不正: verdict=%s kind=%s", verdict, kind)
	}

	// 3) 配信用プールには出ていない・クレジットも付かない（fail-closed）。
	var pooled int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tanzaku WHERE author_id = $1`, author.ID).Scan(&pooled); err != nil {
		t.Fatalf("tanzaku 数の取得: %v", err)
	}
	if pooled != 0 {
		t.Errorf("保留なのに tanzaku が %d 件投入された", pooled)
	}
}

// RecordRejected は配信しない確定判定を rejected として残し判定監査を書く。
// 児童の場合は本文スナップショット付きの児童保全ホールドも同 tx で作る。配信もクレジットもしない。
func TestRecordRejected_却下と監査と児童保全(t *testing.T) {
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
	repo := postgres.NewPoolRepo(pool)

	rejectedFor := func(t *testing.T, v moderation.Verdict) (uuid.UUID, uuid.UUID) {
		t.Helper()
		author, _ := q.CreateUser(ctx)
		rec, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{OwnerID: author.ID, Body: "本文", KoWritten: 33})
		if err != nil {
			t.Fatalf("記録作成: %v", err)
		}
		in := exchange.CastInput{AuthorID: author.ID, SourceRecordID: rec.ID, Body: "本文", Ko: 33}
		if err := repo.RecordRejected(ctx, in, v, "却下理由"); err != nil {
			t.Fatalf("RecordRejected(%s): %v", v.Label(), err)
		}
		var pendingID uuid.UUID
		var status, verdict string
		if err := pool.QueryRow(ctx,
			`SELECT p.id, p.status, g.verdict FROM pending_submission p
			   JOIN gate_verdict g ON g.subject_id = p.id
			  WHERE p.author_id = $1`, author.ID).Scan(&pendingID, &status, &verdict); err != nil {
			t.Fatalf("却下と監査が残っていない: %v", err)
		}
		if status != "rejected" || verdict != v.Label() {
			t.Errorf("却下/監査が不正: status=%s verdict=%s", status, verdict)
		}
		var pooled int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM tanzaku WHERE author_id = $1`, author.ID).Scan(&pooled)
		if pooled != 0 {
			t.Errorf("却下なのに tanzaku が %d 件投入された", pooled)
		}
		return author.ID, pendingID
	}

	t.Run("他者害は却下記録のみ・児童保全なし", func(t *testing.T) {
		authorID, _ := rejectedFor(t, moderation.HarmToOthers)
		var alerts int
		_ = pool.QueryRow(ctx, `SELECT count(*) FROM child_safety_alert WHERE author_id = $1`, authorID).Scan(&alerts)
		if alerts != 0 {
			t.Errorf("他者害なのに児童保全が %d 件作られた", alerts)
		}
	})

	t.Run("児童は本文付きの保全ホールドも作る", func(t *testing.T) {
		authorID, _ := rejectedFor(t, moderation.Child)
		var snap string
		if err := pool.QueryRow(ctx,
			`SELECT body_snapshot FROM child_safety_alert WHERE author_id = $1`, authorID).Scan(&snap); err != nil {
			t.Fatalf("児童保全ホールドが作られていない: %v", err)
		}
		if snap != "本文" {
			t.Errorf("保全の本文スナップショットが不正: %q", snap)
		}
	})
}
