package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
	"github.com/0muji4/Karin/apps/backend/internal/report"
)

// 通報アダプタは、配信で匿名化した後も origin から著者を辿れること、受信者だけが通報できること、
// 二重通報を弾くこと、そして再判定の方針（決着・評判・児童保全）を不可分に反映することを実 DB で確かめる。
func TestReportRepo_提出と決着(t *testing.T) {
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
	recipient, _ := q.CreateUser(ctx)
	stranger, _ := q.CreateUser(ctx)
	rec, err := q.CreateRecord(ctx, sqlcdb.CreateRecordParams{OwnerID: author.ID, Body: "通報対象の本文", KoWritten: 40})
	if err != nil {
		t.Fatalf("記録作成: %v", err)
	}

	// 安全な一枚として投入（tanzaku + origin が作られる）。
	if err := postgres.NewPoolRepo(pool).PoolSafe(ctx, exchange.CastInput{
		AuthorID: author.ID, SourceRecordID: rec.ID, Body: "通報対象の本文", Ko: 40,
	}); err != nil {
		t.Fatalf("PoolSafe: %v", err)
	}
	var tanzakuID uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM tanzaku WHERE author_id = $1`, author.ID).Scan(&tanzakuID); err != nil {
		t.Fatalf("tanzaku 取得: %v", err)
	}
	// 受け手へ配信し、tanzaku.author_id を NULL 化（匿名化）。著者は origin にだけ残る。
	day := time.Date(2026, 6, 28, 0, 0, 0, 0, time.UTC)
	if _, err := pool.Exec(ctx,
		`INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on) VALUES ($1, $2, $3)`,
		tanzakuID, recipient.ID, day); err != nil {
		t.Fatalf("配信: %v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE tanzaku SET status='delivered', author_id=NULL WHERE id=$1`, tanzakuID); err != nil {
		t.Fatalf("匿名化: %v", err)
	}

	repo := postgres.NewReportRepo(pool)

	// 受信していない第三者は通報できない。
	if _, _, err := repo.Submit(ctx, tanzakuID, stranger.ID, "spam", ""); !errors.Is(err, report.ErrNotReceived) {
		t.Errorf("未受信の通報 err=%v, want ErrNotReceived", err)
	}

	// 受け手の通報は通り、匿名化後も origin から著者と本文を復元できる。
	reportID, subj, err := repo.Submit(ctx, tanzakuID, recipient.ID, "harassment", "ひどい")
	if err != nil {
		t.Fatalf("Submit: %v", err)
	}
	if subj.AuthorID != author.ID || subj.Body != "通報対象の本文" {
		t.Errorf("著者/本文の復元が不正: author=%v body=%q", subj.AuthorID, subj.Body)
	}

	// 同じ受け手の二重通報は弾く。
	if _, _, err := repo.Submit(ctx, tanzakuID, recipient.ID, "spam", ""); !errors.Is(err, report.ErrAlreadyReported) {
		t.Errorf("二重通報 err=%v, want ErrAlreadyReported", err)
	}

	// 児童として決着 → 通報 upheld・評判 −1・判定監査・保全ホールドが不可分に入る。
	if err := repo.Resolve(ctx, report.Outcome{
		ReportID: reportID, Subject: subj, Verdict: moderation.Child, Reason: "児童",
		Resolution: report.ResolutionUpheld, ReputationDelta: -1, ChildSafety: true,
	}); err != nil {
		t.Fatalf("Resolve: %v", err)
	}

	var resolution string
	if err := pool.QueryRow(ctx, `SELECT resolution FROM report WHERE id=$1`, reportID).Scan(&resolution); err != nil || resolution != "upheld" {
		t.Errorf("通報の決着が不正: resolution=%s err=%v", resolution, err)
	}
	u, _ := q.GetUserByID(ctx, author.ID)
	if u.Reputation != -1 {
		t.Errorf("著者評判 = %d, want -1", u.Reputation)
	}
	var verdict string
	if err := pool.QueryRow(ctx, `SELECT verdict FROM gate_verdict WHERE subject_kind='tanzaku' AND subject_id=$1`, tanzakuID).Scan(&verdict); err != nil || verdict != "child" {
		t.Errorf("判定監査が不正: verdict=%s err=%v", verdict, err)
	}
	var snap string
	var srcReport uuid.UUID
	if err := pool.QueryRow(ctx,
		`SELECT body_snapshot, source_report_id FROM child_safety_alert WHERE tanzaku_id=$1`, tanzakuID).
		Scan(&snap, &srcReport); err != nil {
		t.Fatalf("児童保全ホールドが作られていない: %v", err)
	}
	if snap != "通報対象の本文" || srcReport != reportID {
		t.Errorf("保全の中身が不正: snap=%q srcReport=%v", snap, srcReport)
	}
}
