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
	pool    Pool
}

// NewCastService は依存ポートから CastService を作る。
func NewCastService(records record.Repository, gate moderation.Moderator, pool Pool) *CastService {
	return &CastService{records: records, gate: gate, pool: pool}
}

// CastToWind は本人の記録 recordID を関門に通し、安全なら複製を未配信プールへ投入する。
// 元の記録は文箱に残る。記録が本人のものでなければ record.ErrNotFound を返す。
//
// M2 では関門は AllPass スタブのため常に安全と判定される。判定結果（プールしたか否か）は
// 呼び出し側で著者に見せない（M3 の四分岐・fail-closed はこのメソッドに実装する）。
func (s *CastService) CastToWind(ctx context.Context, ownerID, recordID uuid.UUID) error {
	rec, err := s.records.Get(ctx, recordID, ownerID)
	if err != nil {
		return err
	}

	dec, err := s.gate.Review(ctx, rec.Body)
	if err != nil {
		// M2: AllPass はエラーを返さない。M3 では fail-closed（保留）に倒す。
		return err
	}
	if dec.Verdict != moderation.Safe {
		// M2: 常に Safe。M3 では他者害/危機/児童の分岐をここに実装する。
		return nil
	}

	_, err = s.pool.Pool(ctx, ownerID, rec.Body, rec.KoWritten, false)
	return err
}
