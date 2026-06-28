package exchange

import (
	"context"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/moderation"
)

// PendingItem は再判定対象の保留（fail-closed で留め置かれた一枚）。
type PendingItem struct {
	ID             uuid.UUID
	AuthorID       uuid.UUID
	SourceRecordID uuid.UUID // 風に乗せた元の記録。未設定なら uuid.Nil
	Body           string
	Ko             int
}

// ReevalStore は保留の再判定で使う永続化ポート。確定操作（昇格・却下）は不可分に行う。
type ReevalStore interface {
	// ListAwaiting は再判定待ちの保留を古い順に最大 limit 件返す。
	ListAwaiting(ctx context.Context, limit int) ([]PendingItem, error)
	// Promote は安全と確定した保留を pooled に昇格する: tanzaku+origin+クレジット +1 を作り、
	// 保留を pooled にする（不可分）。これで保留が通常の配信対象になる。
	Promote(ctx context.Context, item PendingItem) error
	// Reject は不適切と確定した保留を rejected にし、判定監査を残す（児童なら保全も）。不可分。
	Reject(ctx context.Context, item PendingItem, v moderation.Verdict, reason string) error
	// Defer は判定がまだ確定しない保留を据え置く（試行回数を進め原因を残す）。次回また再判定する。
	Defer(ctx context.Context, item PendingItem, cause string) error
}

// Reevaluator は保留の再判定ユースケース。マッチングの前に走らせ、復旧後の保留を捌く。
// LLM 呼び出しを伴うため、配信の単一ライタ・トランザクションの外で実行する。
type Reevaluator struct {
	gate  moderation.Moderator
	store ReevalStore
}

// NewReevaluator は関門と ReevalStore から Reevaluator を作る。
func NewReevaluator(gate moderation.Moderator, store ReevalStore) *Reevaluator {
	return &Reevaluator{gate: gate, store: store}
}

// RunOnce は最大 limit 件の保留を再判定し、安全→昇格・不適切→却下・確定不能→据え置きに振り分ける。
// 1 件の失敗で全体を止めず、原因を伝播して呼び手（cron）に検知させる。
func (r *Reevaluator) RunOnce(ctx context.Context, limit int) error {
	items, err := r.store.ListAwaiting(ctx, limit)
	if err != nil {
		return err
	}
	for _, item := range items {
		dec, err := r.gate.Review(ctx, item.Body)
		if err != nil {
			// まだ確定できない（LLM 不調など）。次回に回す。
			if derr := r.store.Defer(ctx, item, err.Error()); derr != nil {
				return derr
			}
			continue
		}
		if dec.Verdict == moderation.Safe {
			if perr := r.store.Promote(ctx, item); perr != nil {
				return perr
			}
			continue
		}
		if rerr := r.store.Reject(ctx, item, dec.Verdict, dec.Reason); rerr != nil {
			return rerr
		}
	}
	return nil
}
