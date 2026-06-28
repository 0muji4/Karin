package db_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	appdb "github.com/0muji4/Karin/apps/backend/internal/db"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
)

// tableExists は public スキーマに表が存在するかを返す。
func tableExists(t *testing.T, ctx context.Context, url, table string) bool {
	t.Helper()
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatalf("接続失敗: %v", err)
	}
	defer conn.Close(ctx)
	var reg *string
	if err := conn.QueryRow(ctx, "SELECT to_regclass($1)::text", "public."+table).Scan(&reg); err != nil {
		t.Fatalf("to_regclass 失敗: %v", err)
	}
	return reg != nil
}

func TestMigrate_UpDownIsIdempotentAndReversible(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	ctx := context.Background()
	url, terminate, err := dbtest.RunPostgres(ctx)
	if err != nil {
		t.Fatalf("PG 起動失敗: %v", err)
	}
	defer terminate()

	wantTables := []string{"users", "auth_token", "record", "ko_reference"}

	// Up: 全表ができる。
	if err := appdb.MigrateUp(url); err != nil {
		t.Fatalf("MigrateUp: %v", err)
	}
	for _, tb := range wantTables {
		if !tableExists(t, ctx, url, tb) {
			t.Errorf("Up 後に %s が無い", tb)
		}
	}

	// 冪等: 2 回目の Up はエラーにならない（ErrNoChange を握る）。
	if err := appdb.MigrateUp(url); err != nil {
		t.Fatalf("2 回目の MigrateUp: %v", err)
	}

	// Down: 全表が消える。
	if err := appdb.MigrateDown(url); err != nil {
		t.Fatalf("MigrateDown: %v", err)
	}
	for _, tb := range wantTables {
		if tableExists(t, ctx, url, tb) {
			t.Errorf("Down 後に %s が残っている", tb)
		}
	}

	// 再 Up: 巻き戻し後も再適用できる。
	if err := appdb.MigrateUp(url); err != nil {
		t.Fatalf("再 MigrateUp: %v", err)
	}
}
