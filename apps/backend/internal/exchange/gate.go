package exchange

import (
	"context"

	"github.com/google/uuid"
)

// CastInput は風に乗せる一枚の素材（記録から複製した本文・候と、その由来）。
// SourceRecordID は著者特定の私的リンク(origin)に残す元の記録。
type CastInput struct {
	AuthorID       uuid.UUID
	SourceRecordID uuid.UUID
	Body           string
	Ko             int
}

// GateEffects は関門の判定結果に応じた効果を永続化するポート。
// 分類は moderation が担い、どの効果を起こすかは exchange が決め、実際の書き込み
// （不可分トランザクションを含む）は実装(postgres)が担う。四分岐の各効果はマイルストーンを
// 追って本ポートに足していく。
type GateEffects interface {
	// PoolSafe は安全な一枚を pooled で投入し、著者特定の私的リンク(origin)と
	// 著者のクレジット +1 を不可分に書く。配信時に author_id を NULL 化しても
	// origin から著者を辿れる不変条件を、ここで保証する（通報→評判・児童保全の根拠）。
	PoolSafe(ctx context.Context, in CastInput) error
}
