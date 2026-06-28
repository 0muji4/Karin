// Package migrations は db/migrations/ の SQL を実行ファイルに埋め込む。
//
// //go:embed は親ディレクトリ（..）を辿れないため、埋め込み用のこのファイルを
// migrations/ の 1 つ上（apps/backend/db/）に置く。これにより testcontainers の
// 結合テストでも本番でも、同じマイグレーション群をバイナリ内から適用できる。
package migrations

import "embed"

// FS は番号付きの *.up.sql / *.down.sql を保持する。
//
//go:embed migrations/*.sql
var FS embed.FS
