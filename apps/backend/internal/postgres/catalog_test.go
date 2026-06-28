package postgres_test

import (
	"context"
	"testing"

	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
	"github.com/0muji4/Karin/apps/backend/internal/postgres"
)

// KoCatalog アダプタが seed 済みの ko_reference を正しく読むことを実 DB で検証する。
func TestKoCatalog(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	ctx := context.Background()
	pool, _, terminate, err := dbtest.MigratedPool(ctx)
	if err != nil {
		t.Fatalf("PG 起動失敗: %v", err)
	}
	defer terminate()

	cat := postgres.NewKoCatalog(pool)

	// Get: 候29 は菖蒲華。
	m, err := cat.Get(ctx, 29)
	if err != nil {
		t.Fatalf("Get(29): %v", err)
	}
	if m.Number != 29 || m.Name != "菖蒲華" || m.Season != "summer" {
		t.Errorf("Get(29) = %+v, want {29 菖蒲華 ... summer}", m)
	}

	// List: 72 件そろう。
	all, err := cat.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(all) != 72 {
		t.Fatalf("List 件数 = %d, want 72", len(all))
	}
	if all[0].Number != 1 || all[71].Number != 72 {
		t.Errorf("List の並びが候番号順でない: first=%d last=%d", all[0].Number, all[71].Number)
	}
}
