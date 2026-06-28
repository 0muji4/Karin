package db_test

import (
	"context"
	"testing"

	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
)

// 七十二候 seed が 72 件そろい、sekki / season が候番号と整合することを確認する。
func TestKoReferenceSeed(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	ctx := context.Background()
	pool, _, terminate, err := dbtest.MigratedPool(ctx)
	if err != nil {
		t.Fatalf("PG 起動失敗: %v", err)
	}
	defer terminate()

	// 件数。
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM ko_reference").Scan(&count); err != nil {
		t.Fatalf("count 失敗: %v", err)
	}
	if count != 72 {
		t.Fatalf("ko_reference の件数 = %d, want 72", count)
	}

	// 全候で sekki = (ko-1)/3 + 1、かつ season が二十四節気のバケットと一致すること。
	rows, err := pool.Query(ctx, `SELECT ko, sekki, season FROM ko_reference ORDER BY ko`)
	if err != nil {
		t.Fatalf("query 失敗: %v", err)
	}
	defer rows.Close()
	seen := 0
	for rows.Next() {
		var ko, sekki int
		var season string
		if err := rows.Scan(&ko, &sekki, &season); err != nil {
			t.Fatalf("scan 失敗: %v", err)
		}
		seen++
		if want := (ko-1)/3 + 1; sekki != want {
			t.Errorf("ko=%d: sekki=%d, want %d", ko, sekki, want)
		}
		if want := seasonForSekki(sekki); season != want {
			t.Errorf("ko=%d sekki=%d: season=%q, want %q", ko, sekki, season, want)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows 反復で失敗: %v", err)
	}
	if seen != 72 {
		t.Fatalf("反復した行数 = %d, want 72", seen)
	}

	// 代表的な候の名称を抜き取り確認（夏至の候）。
	var name string
	if err := pool.QueryRow(ctx, "SELECT name FROM ko_reference WHERE ko = 29").Scan(&name); err != nil {
		t.Fatalf("ko=29 取得失敗: %v", err)
	}
	if name != "菖蒲華" {
		t.Errorf("ko=29 の name = %q, want 菖蒲華", name)
	}
}

func seasonForSekki(sekki int) string {
	switch {
	case sekki <= 6:
		return "spring"
	case sekki <= 12:
		return "summer"
	case sekki <= 18:
		return "autumn"
	default:
		return "winter"
	}
}
