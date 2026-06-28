// Package postgres は各ドメインのポート（リポジトリ・カタログ）を sqlc/pgx で満たすアダプタ層。
// この層だけが sqlcdb/pgx を知り、ドメイン・ユースケースは知らない（依存性ルール）。
package postgres

import (
	"context"
	"fmt"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/ko"
)

// KoCatalog は ko.Catalog を ko_reference テーブルで満たす。
type KoCatalog struct {
	q *sqlcdb.Queries
}

// NewKoCatalog は接続（プール）から KoCatalog を作る。
func NewKoCatalog(db sqlcdb.DBTX) *KoCatalog {
	return &KoCatalog{q: sqlcdb.New(db)}
}

// Get は候番号からメタを引く。
func (c *KoCatalog) Get(ctx context.Context, n int) (ko.Meta, error) {
	row, err := c.q.GetKoReference(ctx, int16(n))
	if err != nil {
		return ko.Meta{}, fmt.Errorf("候メタの取得に失敗: %w", err)
	}
	return toMeta(row), nil
}

// List は全 72 候のメタを候番号順に返す。
func (c *KoCatalog) List(ctx context.Context) ([]ko.Meta, error) {
	rows, err := c.q.ListKoReference(ctx)
	if err != nil {
		return nil, fmt.Errorf("候メタ一覧の取得に失敗: %w", err)
	}
	out := make([]ko.Meta, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMeta(r))
	}
	return out, nil
}

func toMeta(m sqlcdb.KoReference) ko.Meta {
	return ko.Meta{
		Number:  int(m.Ko),
		Name:    m.Name,
		Kana:    m.Kana,
		Meaning: m.Meaning,
		Sekki:   int(m.Sekki),
		Season:  m.Season,
	}
}
