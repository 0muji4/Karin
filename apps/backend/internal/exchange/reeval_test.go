package exchange_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
)

var errReeval = errors.New("reeval: llm 不調")

// fakeReevalStore は exchange.ReevalStore の DB 不要なテスト実装。
type fakeReevalStore struct {
	items       []exchange.PendingItem
	promoted    []uuid.UUID
	rejected    []uuid.UUID
	rejectedVer moderation.Verdict
	deferred    []uuid.UUID
}

func (f *fakeReevalStore) ListAwaiting(context.Context, int) ([]exchange.PendingItem, error) {
	return f.items, nil
}
func (f *fakeReevalStore) Promote(_ context.Context, item exchange.PendingItem) error {
	f.promoted = append(f.promoted, item.ID)
	return nil
}
func (f *fakeReevalStore) Reject(_ context.Context, item exchange.PendingItem, v moderation.Verdict, _ string) error {
	f.rejected = append(f.rejected, item.ID)
	f.rejectedVer = v
	return nil
}
func (f *fakeReevalStore) Defer(_ context.Context, item exchange.PendingItem, _ string) error {
	f.deferred = append(f.deferred, item.ID)
	return nil
}

// 再判定の結果に応じて、保留は昇格・却下・据え置きへ振り分けられる。
// verdictModerator / errModerator は cast_test.go の定義を共用する（同一 _test パッケージ）。
func TestReeval_判定で振り分ける(t *testing.T) {
	id := uuid.New()
	items := []exchange.PendingItem{{ID: id, AuthorID: uuid.New(), Body: "本文", Ko: 30}}

	t.Run("安全は昇格", func(t *testing.T) {
		store := &fakeReevalStore{items: items}
		r := exchange.NewReevaluator(verdictModerator{v: moderation.Safe}, store)
		if err := r.RunOnce(context.Background(), 10); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}
		if len(store.promoted) != 1 || store.promoted[0] != id {
			t.Errorf("昇格されていない: %v", store.promoted)
		}
		if len(store.rejected) != 0 || len(store.deferred) != 0 {
			t.Error("安全なのに却下/据え置きが起きた")
		}
	})

	t.Run("児童は却下", func(t *testing.T) {
		store := &fakeReevalStore{items: items}
		r := exchange.NewReevaluator(verdictModerator{v: moderation.Child}, store)
		if err := r.RunOnce(context.Background(), 10); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}
		if len(store.rejected) != 1 || store.rejectedVer != moderation.Child {
			t.Errorf("却下されていない: rejected=%v ver=%v", store.rejected, store.rejectedVer)
		}
		if len(store.promoted) != 0 {
			t.Error("児童なのに昇格された")
		}
	})

	t.Run("判定不能は据え置き", func(t *testing.T) {
		store := &fakeReevalStore{items: items}
		r := exchange.NewReevaluator(errModerator{err: errReeval}, store)
		if err := r.RunOnce(context.Background(), 10); err != nil {
			t.Fatalf("RunOnce: %v", err)
		}
		if len(store.deferred) != 1 {
			t.Errorf("据え置かれていない: %v", store.deferred)
		}
		if len(store.promoted) != 0 || len(store.rejected) != 0 {
			t.Error("確定不能なのに昇格/却下された")
		}
	})
}
