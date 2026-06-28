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

// CastToWind は本人の記録 recordID を関門に通し、安全なら複製を未配信プールへ投入する。
// 元の記録は文箱に残る。記録が本人のものでなければ record.ErrNotFound を返す。
//
// 判定結果（プールしたか否か）は著者に見せない。安全な一枚は origin と不可分に投入し、
// 配信で匿名化したあとも著者を辿れるようにする。
// 他者害/危機/児童の効果適用と、判定が確定しなかったときの fail-closed 保留は後続 PR で足す。
func (s *CastService) CastToWind(ctx context.Context, ownerID, recordID uuid.UUID) error {
	rec, err := s.records.Get(ctx, recordID, ownerID)
	if err != nil {
		return err
	}

	dec, err := s.gate.Review(ctx, rec.Body)
	if err != nil {
		// fail-closed（保留）は後続 PR。現状は配信せずエラーを伝播する。
		return err
	}
	if dec.Verdict != moderation.Safe {
		// 他者害/危機/児童の効果適用は後続 PR。現状は配信しない。
		return nil
	}

	return s.effects.PoolSafe(ctx, CastInput{
		AuthorID:       ownerID,
		SourceRecordID: recordID,
		Body:           rec.Body,
		Ko:             rec.KoWritten,
	})
}
