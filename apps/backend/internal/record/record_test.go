package record_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

func TestRecord_OwnerOnlyAndValidation(t *testing.T) {
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
	ownerA, err := q.CreateUser(ctx)
	if err != nil {
		t.Fatalf("ユーザーA作成: %v", err)
	}
	ownerB, err := q.CreateUser(ctx)
	if err != nil {
		t.Fatalf("ユーザーB作成: %v", err)
	}

	svc := record.NewService(pool)

	rec, err := svc.Create(ctx, ownerA.ID, "桜が咲いた", 11)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// A は自分の記録を取得できる。
	if _, err := svc.Get(ctx, rec.ID, ownerA.ID); err != nil {
		t.Errorf("A による Get に失敗: %v", err)
	}
	// B は A の記録 ID を指定しても取得できない（owner-only）。
	if _, err := svc.Get(ctx, rec.ID, ownerB.ID); !errors.Is(err, record.ErrNotFound) {
		t.Errorf("B による Get: err = %v, want ErrNotFound", err)
	}
	// 一覧も owner ごとに分離される。
	listA, err := svc.ListByOwner(ctx, ownerA.ID)
	if err != nil || len(listA) != 1 {
		t.Errorf("A の一覧 = %d 件 (err=%v), want 1", len(listA), err)
	}
	listB, err := svc.ListByOwner(ctx, ownerB.ID)
	if err != nil || len(listB) != 0 {
		t.Errorf("B の一覧 = %d 件 (err=%v), want 0", len(listB), err)
	}

	// バリデーション: 空本文・候範囲外・長すぎる本文は ErrInvalid。
	bad := []struct {
		name string
		body string
		ko   int
	}{
		{"空本文", "   ", 11},
		{"候0", "x", 0},
		{"候73", "x", 73},
		{"本文超過", strings.Repeat("あ", record.MaxBodyRunes+1), 11},
	}
	for _, b := range bad {
		if _, err := svc.Create(ctx, ownerA.ID, b.body, b.ko); !errors.Is(err, record.ErrInvalid) {
			t.Errorf("%s: err = %v, want ErrInvalid", b.name, err)
		}
	}
}
