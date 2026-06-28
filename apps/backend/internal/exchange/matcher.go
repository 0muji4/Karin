package exchange

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/ko"
)

// Recipient は受信資格者と、その人へ割り当て可能な未配信の枚数。
type Recipient struct {
	UserID     uuid.UUID
	Assignable int
}

// MatchTx は日次マッチャの 1 トランザクション内で使える操作（単一ライタ前提）。
// 受け手≠書き手・最古から・割当可能少順といった SQL 寄りの選別はアダプタが担い、
// 並べる順序の利用と割り当てループはユースケース（Matcher）が担う。
type MatchTx interface {
	// ExpirePooledBefore は cutoff より前にプールされた未配信を expired にする（候TTL）。
	ExpirePooledBefore(ctx context.Context, cutoff time.Time) (int, error)
	// EligibleRecipients は受信資格者を、割当可能枚数の少ない順で返す。
	EligibleRecipients(ctx context.Context, runDate time.Time) ([]Recipient, error)
	// PickOldestPooledFor は recipient が著者でない最古の未配信を 1 枚ロックして引く。
	PickOldestPooledFor(ctx context.Context, recipientID uuid.UUID) (tanzakuID uuid.UUID, found bool, err error)
	// Deliver は配信を記録し、配信済みにして著者を剥がし、受け手のクレジットを −1 する。
	Deliver(ctx context.Context, tanzakuID, recipientID uuid.UUID, runDate time.Time) error
}

// MatchStore は単一ライタの日次トランザクションを張る（advisory lock はアダプタが取る）。
type MatchStore interface {
	RunDaily(ctx context.Context, fn func(MatchTx) error) error
}

// Matcher は日次マッチングのユースケース。
type Matcher struct {
	store MatchStore
	ttlKo int              // 候TTL（N候）
	now   func() time.Time // テストで時刻を固定できるよう注入可能
}

// NewMatcher は MatchStore と候TTL（N候）から Matcher を作る。
func NewMatcher(store MatchStore, ttlKo int) *Matcher {
	return &Matcher{store: store, ttlKo: ttlKo, now: time.Now}
}

// RunDaily は 1 回分のマッチングを実行する。
// 手順: ①古い未配信を候TTLで expired に ②資格者を割当少順に集める
// ③各人へ最古の一枚（受け手≠書き手）を割り当てる。割り当てられない人はその日は受け取らない。
func (m *Matcher) RunDaily(ctx context.Context) error {
	return m.store.RunDaily(ctx, func(tx MatchTx) error {
		now := m.now()
		if _, err := tx.ExpirePooledBefore(ctx, ko.TTLCutoff(now, m.ttlKo)); err != nil {
			return err
		}
		recipients, err := tx.EligibleRecipients(ctx, now)
		if err != nil {
			return err
		}
		for _, r := range recipients {
			id, found, err := tx.PickOldestPooledFor(ctx, r.UserID)
			if err != nil {
				return err
			}
			if !found {
				continue // 割り当て可能な短冊が無ければその日は受け取らない（クレジットは残る）
			}
			if err := tx.Deliver(ctx, id, r.UserID, now); err != nil {
				return err
			}
		}
		return nil
	})
}
