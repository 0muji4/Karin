// Package moderation は、交換に出された本文を出口で判定する関門ポートと実装を提供する。
// 関門は分類だけを担い、判定結果に基づく効果（プール投入・配信停止など）は exchange が適用する。
package moderation

import "context"

// Verdict は関門の四分岐。
type Verdict int

const (
	// Safe は配信可。
	Safe Verdict = iota
	// HarmToOthers は他者への害（嫌がらせ・個人情報）・曖昧。配信しない。
	HarmToOthers
	// Crisis は危機・自傷。配信せず本人に支援先を案内する。
	Crisis
	// Child は児童関連。配信せず運営へ通報・保全する。
	Child
)

// Decision は関門の判定。Reason はログ・監査用で、著者には見せない。
type Decision struct {
	Verdict  Verdict
	Reason   string
	Provider string
}

// Moderator は本文を分類するポート。error は呼び手が fail-closed（配信しない）に倒す。
type Moderator interface {
	Review(ctx context.Context, body string) (Decision, error)
}

// AllPass は常に安全と判定するスタブ（M2）。M3 で LLM アダプタに差し替える。
type AllPass struct{}

// Review は常に Safe を返す。
func (AllPass) Review(context.Context, string) (Decision, error) {
	return Decision{Verdict: Safe, Provider: "allpass"}, nil
}
