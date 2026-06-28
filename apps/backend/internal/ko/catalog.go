package ko

import "context"

// Meta は候のメタ情報（名称・読み・意味・節気・季節）。候番号は Number が計算し、
// 名称などのメタは Catalog が外部（永続層）から取得する。
type Meta struct {
	Number  int
	Name    string
	Kana    string
	Meaning string
	Sekki   int
	Season  string
}

// Catalog は候メタの読み取りポート。実装（永続層アダプタ）は ko を import する側に置き、
// 依存の向きを内向き（アダプタ → ドメイン）に保つ。
type Catalog interface {
	Get(ctx context.Context, ko int) (Meta, error)
	List(ctx context.Context) ([]Meta, error)
}
