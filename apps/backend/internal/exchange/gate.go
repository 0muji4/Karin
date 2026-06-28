package exchange

import (
	"context"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/moderation"
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

	// HoldForReview は判定が安全と確定しなかった一枚を配信せず保留する（fail-closed）。
	// 本文をスナップショットして awaiting で保持し、復旧後にマッチャが再判定する。
	// cause は保留に至った原因（LLM の呼び出し/解釈エラー等）で、監査にだけ残し著者には見せない。
	HoldForReview(ctx context.Context, in CastInput, cause string) error

	// RecordRejected は配信しないと確定した一枚（他者害・危機・児童）を rejected として残し、
	// 判定監査を書く。児童の場合は児童保全のホールドも不可分に作る。配信もクレジット加算もしない。
	// reason は監査用で、著者には見せない。
	RecordRejected(ctx context.Context, in CastInput, v moderation.Verdict, reason string) error
}
