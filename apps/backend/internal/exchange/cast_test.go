package exchange_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/exchange"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

// fakeRecords は record.Repository の DB 不要なテスト実装。
type fakeRecords struct {
	rec      record.Record
	err      error
	gotID    uuid.UUID
	gotOwner uuid.UUID
}

func (f *fakeRecords) Create(context.Context, uuid.UUID, string, int) (record.Record, error) {
	return record.Record{}, nil
}
func (f *fakeRecords) ListByOwner(context.Context, uuid.UUID) ([]record.Record, error) {
	return nil, nil
}
func (f *fakeRecords) Get(_ context.Context, id, ownerID uuid.UUID) (record.Record, error) {
	f.gotID, f.gotOwner = id, ownerID
	return f.rec, f.err
}

// fakePool は exchange.Pool の DB 不要なテスト実装。
type fakePool struct {
	called    bool
	gotAuthor uuid.UUID
	gotBody   string
	gotKo     int
}

func (f *fakePool) Pool(_ context.Context, authorID uuid.UUID, body string, ko int, _ bool) (exchange.Tanzaku, error) {
	f.called = true
	f.gotAuthor, f.gotBody, f.gotKo = authorID, body, ko
	return exchange.Tanzaku{}, nil
}

// 安全な記録は、本人のものとして読まれ、複製がプールへ投入される（著者＝本人・本文と候を引き継ぐ）。
func TestCastToWind_PoolsCopy(t *testing.T) {
	owner, recID := uuid.New(), uuid.New()
	recs := &fakeRecords{rec: record.Record{Body: "桜が咲いた", KoWritten: 11}}
	pool := &fakePool{}
	svc := exchange.NewCastService(recs, moderation.AllPass{}, pool)

	if err := svc.CastToWind(context.Background(), owner, recID); err != nil {
		t.Fatalf("CastToWind: %v", err)
	}
	if recs.gotID != recID || recs.gotOwner != owner {
		t.Errorf("owner-only で記録を読んでいない: id=%v owner=%v", recs.gotID, recs.gotOwner)
	}
	if !pool.called {
		t.Fatal("プールへ投入されていない")
	}
	if pool.gotAuthor != owner || pool.gotBody != "桜が咲いた" || pool.gotKo != 11 {
		t.Errorf("投入内容が不正: author=%v body=%q ko=%d", pool.gotAuthor, pool.gotBody, pool.gotKo)
	}
}

// 他人の記録（見つからない）は record.ErrNotFound を伝播し、プールしない。
func TestCastToWind_NotFoundPropagates(t *testing.T) {
	recs := &fakeRecords{err: record.ErrNotFound}
	pool := &fakePool{}
	svc := exchange.NewCastService(recs, moderation.AllPass{}, pool)

	if err := svc.CastToWind(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, record.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if pool.called {
		t.Error("見つからないのにプールへ投入された")
	}
}
