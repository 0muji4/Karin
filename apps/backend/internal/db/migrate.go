package db

import (
	"errors"
	"fmt"
	"io/fs"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5" // scheme "pgx5" を登録
	"github.com/golang-migrate/migrate/v4/source/iofs"

	migrationfs "github.com/0muji4/Karin/apps/backend/db"
)

// migrationsSubdir は埋め込み FS 内のマイグレーション格納ディレクトリ。
const migrationsSubdir = "migrations"

// MigrateUp は埋め込み済みのマイグレーションをすべて適用する。
// 変更が無い場合（適用済み）はエラーにしない。
func MigrateUp(databaseURL string) error {
	return withMigrator(migrationfs.FS, migrationsSubdir, databaseURL, func(m *migrate.Migrate) error {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("マイグレーション適用に失敗: %w", err)
		}
		return nil
	})
}

// MigrateDown はすべてのマイグレーションを巻き戻す（主にテスト・検証用）。
func MigrateDown(databaseURL string) error {
	return withMigrator(migrationfs.FS, migrationsSubdir, databaseURL, func(m *migrate.Migrate) error {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("マイグレーション巻き戻しに失敗: %w", err)
		}
		return nil
	})
}

// withMigrator は任意の fs.FS からマイグレータを組み立て、fn を実行して後始末する。
// fs.FS を引数に取ることで、本番（埋め込み）とテスト（testdata）を同じ経路で扱える。
func withMigrator(fsys fs.FS, subdir, databaseURL string, fn func(*migrate.Migrate) error) error {
	src, err := iofs.New(fsys, subdir)
	if err != nil {
		return fmt.Errorf("マイグレーション元の読み込みに失敗: %w", err)
	}
	dbURL, err := toPgx5URL(databaseURL)
	if err != nil {
		return err
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		return fmt.Errorf("マイグレータの初期化に失敗: %w", err)
	}
	// source/database のクローズ由来エラーは握り潰さず、fn のエラーを優先して返す。
	runErr := fn(m)
	srcErr, dbErr := m.Close()
	if runErr != nil {
		return runErr
	}
	return errors.Join(srcErr, dbErr)
}

// toPgx5URL は postgres:// を golang-migrate の pgx5 ドライバ用スキームに変換する。
func toPgx5URL(databaseURL string) (string, error) {
	u, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("接続文字列の解析に失敗: %w", err)
	}
	switch u.Scheme {
	case "postgres", "postgresql", "pgx5":
		u.Scheme = "pgx5"
	default:
		return "", fmt.Errorf("未対応の接続スキーム: %q（postgres:// を渡す）", u.Scheme)
	}
	return u.String(), nil
}
