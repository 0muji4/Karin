package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/exchange"
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
