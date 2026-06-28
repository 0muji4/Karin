package exchange_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/exchange"
)

type pooledCard struct {
	id     uuid.UUID
	author uuid.UUID
}

type delivered struct {
	tanzakuID uuid.UUID
	recipient uuid.UUID
}

// fakeMatchTx はインメモリのプールで MatchTx を模す。受け手≠書き手・取り合い防止を素朴に再現する。
type fakeMatchTx struct {
	pool         []pooledCard
	recipients   []exchange.Recipient
	expireCalled bool
	delivered    []delivered
}

func (tx *fakeMatchTx) ExpirePooledBefore(context.Context, time.Time) (int, error) {
	tx.expireCalled = true
	return 0, nil
}
func (tx *fakeMatchTx) EligibleRecipients(context.Context, time.Time) ([]exchange.Recipient, error) {
	return tx.recipients, nil
}
func (tx *fakeMatchTx) PickOldestPooledFor(_ context.Context, recipientID uuid.UUID) (uuid.UUID, bool, error) {
	for i, c := range tx.pool {
		if c.author != recipientID { // 受け手≠書き手
			tx.pool = append(tx.pool[:i:i], tx.pool[i+1:]...) // 取った分は抜く（二重配信防止）
			return c.id, true, nil
		}
	}
	return uuid.Nil, false, nil
}
func (tx *fakeMatchTx) Deliver(_ context.Context, tanzakuID, recipientID uuid.UUID, _ time.Time) error {
	tx.delivered = append(tx.delivered, delivered{tanzakuID, recipientID})
	return nil
}

type fakeMatchStore struct{ tx *fakeMatchTx }

func (s *fakeMatchStore) RunDaily(_ context.Context, fn func(exchange.MatchTx) error) error {
	return fn(s.tx)
}

// 各資格者に最古の一枚が、受け手≠書き手で割り当てられる。TTL も呼ばれる。
func TestMatcher_RunDaily_AssignsRecipientNotAuthor(t *testing.T) {
	u1, u2 := uuid.New(), uuid.New()
	cA, cB := uuid.New(), uuid.New() // cA は u1 作、cB は u2 作
	tx := &fakeMatchTx{
		pool:       []pooledCard{{cA, u1}, {cB, u2}},
		recipients: []exchange.Recipient{{UserID: u1, Assignable: 1}, {UserID: u2, Assignable: 1}},
	}
	m := exchange.NewMatcher(&fakeMatchStore{tx: tx}, 6)

	if err := m.RunDaily(context.Background()); err != nil {
		t.Fatalf("RunDaily: %v", err)
	}
	if !tx.expireCalled {
		t.Error("候TTL（ExpirePooledBefore）が呼ばれていない")
	}
	if len(tx.delivered) != 2 {
		t.Fatalf("配信数 = %d, want 2", len(tx.delivered))
	}
	for _, d := range tx.delivered {
		if d.tanzakuID == cA && d.recipient == u1 {
			t.Error("自分の短冊(cA)が作者 u1 に返った")
		}
		if d.tanzakuID == cB && d.recipient == u2 {
			t.Error("自分の短冊(cB)が作者 u2 に返った")
		}
	}
}

// 割り当て可能な短冊が無い人（自分の一枚しか無い）は、その日は受け取らない。
func TestMatcher_RunDaily_SkipsWhenOnlyOwnCard(t *testing.T) {
	u1 := uuid.New()
	cA := uuid.New() // u1 作
	tx := &fakeMatchTx{
		pool:       []pooledCard{{cA, u1}},
		recipients: []exchange.Recipient{{UserID: u1, Assignable: 0}},
	}
	m := exchange.NewMatcher(&fakeMatchStore{tx: tx}, 6)

	if err := m.RunDaily(context.Background()); err != nil {
		t.Fatalf("RunDaily: %v", err)
	}
	if len(tx.delivered) != 0 {
		t.Errorf("自分の一枚しか無いのに配信された: %d", len(tx.delivered))
	}
}
