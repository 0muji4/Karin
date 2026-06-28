// Command migrate はデータベースのマイグレーションを適用・巻き戻しする薄い CLI。
//
//	DATABASE_URL=postgres://... migrate up
//	DATABASE_URL=postgres://... migrate down
//
// デプロイ前の適用やローカルでの up/down 確認に使う。
package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/0muji4/Karin/apps/backend/internal/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate up|down")
		os.Exit(2)
	}
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		logger.Error("DATABASE_URL が未設定")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "up":
		err = db.MigrateUp(url)
	case "down":
		err = db.MigrateDown(url)
	default:
		fmt.Fprintln(os.Stderr, "usage: migrate up|down")
		os.Exit(2)
	}
	if err != nil {
		logger.Error("マイグレーション失敗", "command", os.Args[1], "error", err)
		os.Exit(1)
	}
	logger.Info("マイグレーション完了", "command", os.Args[1])
}
