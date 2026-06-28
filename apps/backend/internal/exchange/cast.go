package exchange

import (
	"context"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

// CastService は「風に乗せる」ユースケース。本人の記録を関門に通し、安全なら複製をプールへ投入する。
type CastService struct {
	records record.Repository
	gate    moderation.Moderator
	effects GateEffects
}

// NewCastService は依存ポートから CastService を作る。
func NewCastService(records record.Repository, gate moderation.Moderator, effects GateEffects) *CastService {
	return &CastService{records: records, gate: gate, effects: effects}
}

// CastOutcome は風に乗せた結果のうち、著者に返してよい情報。判定そのものは見せない。
// 危機（自傷）と判定したときだけ、本人に支援先の案内を促す。
type CastOutcome struct {
	ShowCrisisSupport bool
}

// CastToWind は本人の記録 recordID を関門に通し、判定に応じて効果を適用する。
// 元の記録は文箱に残る。記録が本人のものでなければ record.ErrNotFound を返す。
//
// 四分岐: 安全→origin と不可分にプール投入（配信で匿名化後も著者を辿れる）。他者害/児童→配信せず
// rejected として記録（児童は保全ホールドも）。危機→配信しないが本人に支援先を案内。
// 判定が確定しない→fail-closed で保留し復旧後に再判定。
// 判定そのものは著者に見せないため、危機の支援案内を除き応答は一律になる。
func (s *CastService) CastToWind(ctx context.Context, ownerID, recordID uuid.UUID) (CastOutcome, error) {
	rec, err := s.records.Get(ctx, recordID, ownerID)
	if err != nil {
		return CastOutcome{}, err
	}
	in := CastInput{
		AuthorID:       ownerID,
		SourceRecordID: recordID,
		Body:           rec.Body,
		Ko:             rec.KoWritten,
	}

	dec, err := s.gate.Review(ctx, rec.Body)
	if err != nil {
		// fail-closed: 判定が確定しないものは配信せず保留し、復旧後に再判定する。
		return CastOutcome{}, s.effects.HoldForReview(ctx, in, err.Error())
	}

	switch dec.Verdict {
	case moderation.Safe:
		return CastOutcome{}, s.effects.PoolSafe(ctx, in)
	case moderation.Crisis:
		// 配信しないが、本人にだけ支援先を案内する。
		return CastOutcome{ShowCrisisSupport: true}, s.effects.RecordRejected(ctx, in, dec.Verdict, dec.Reason)
	default:
		// 他者害・児童は配信しない。応答は安全時と同じ一律（判定を見せない）。
		return CastOutcome{}, s.effects.RecordRejected(ctx, in, dec.Verdict, dec.Reason)
	}
}
