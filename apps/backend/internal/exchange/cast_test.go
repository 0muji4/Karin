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

// fakeEffects は exchange.GateEffects の DB 不要なテスト実装。
type fakeEffects struct {
	pooled      bool
	got         exchange.CastInput
	held        bool
	heldCause   string
	rejected    bool
	rejectedVer moderation.Verdict
}

func (f *fakeEffects) PoolSafe(_ context.Context, in exchange.CastInput) error {
	f.pooled = true
	f.got = in
	return nil
}

func (f *fakeEffects) HoldForReview(_ context.Context, in exchange.CastInput, cause string) error {
	f.held = true
	f.got = in
	f.heldCause = cause
	return nil
}

func (f *fakeEffects) RecordRejected(_ context.Context, in exchange.CastInput, v moderation.Verdict, _ string) error {
	f.rejected = true
	f.got = in
	f.rejectedVer = v
	return nil
}

// errModerator は常に判定エラーを返す関門スタブ（fail-closed の検証用）。
type errModerator struct{ err error }

func (m errModerator) Review(context.Context, string) (moderation.Decision, error) {
	return moderation.Decision{}, m.err
}

// verdictModerator は指定した判定を返す関門スタブ（四分岐の検証用）。
type verdictModerator struct{ v moderation.Verdict }

func (m verdictModerator) Review(context.Context, string) (moderation.Decision, error) {
	return moderation.Decision{Verdict: m.v, Reason: "test"}, nil
}

// 安全な記録は、本人のものとして読まれ、複製がプールへ投入される。
// 著者＝本人・本文と候を引き継ぎ、由来(SourceRecordID)を origin 用に渡す。
func TestCastToWind_PoolsCopy(t *testing.T) {
	owner, recID := uuid.New(), uuid.New()
	recs := &fakeRecords{rec: record.Record{Body: "桜が咲いた", KoWritten: 11}}
	effects := &fakeEffects{}
	svc := exchange.NewCastService(recs, moderation.AllPass{}, effects)

	if _, err := svc.CastToWind(context.Background(), owner, recID); err != nil {
		t.Fatalf("CastToWind: %v", err)
	}
	if recs.gotID != recID || recs.gotOwner != owner {
		t.Errorf("owner-only で記録を読んでいない: id=%v owner=%v", recs.gotID, recs.gotOwner)
	}
	if !effects.pooled {
		t.Fatal("プールへ投入されていない")
	}
	if effects.got.AuthorID != owner || effects.got.Body != "桜が咲いた" || effects.got.Ko != 11 {
		t.Errorf("投入内容が不正: %+v", effects.got)
	}
	if effects.got.SourceRecordID != recID {
		t.Errorf("origin に残す由来が不正: SourceRecordID=%v, want %v", effects.got.SourceRecordID, recID)
	}
}

// 判定が確定しない（関門がエラー）ときは配信せず保留する（fail-closed）。
// 判定は著者に見せないので、保留は成功と同じ一律の応答になる（エラーにしない）。
func TestCastToWind_GateErrorHolds(t *testing.T) {
	recs := &fakeRecords{rec: record.Record{Body: "曖昧な一文", KoWritten: 20}}
	effects := &fakeEffects{}
	svc := exchange.NewCastService(recs, errModerator{err: errors.New("llm timeout")}, effects)

	if _, err := svc.CastToWind(context.Background(), uuid.New(), uuid.New()); err != nil {
		t.Fatalf("保留時はエラーにしない: %v", err)
	}
	if effects.pooled {
		t.Error("判定エラーなのにプールへ投入された（fail-closed が崩れる）")
	}
	if !effects.held {
		t.Fatal("判定エラーなのに保留されていない")
	}
	if effects.heldCause == "" {
		t.Error("保留原因が監査に渡っていない")
	}
}

// 他者害・児童は配信せず rejected として記録し、プールしない。応答は一律（支援案内なし）。
func TestCastToWind_非安全は却下しプールしない(t *testing.T) {
	for _, v := range []moderation.Verdict{moderation.HarmToOthers, moderation.Child} {
		t.Run(v.Label(), func(t *testing.T) {
			recs := &fakeRecords{rec: record.Record{Body: "x", KoWritten: 20}}
			effects := &fakeEffects{}
			svc := exchange.NewCastService(recs, verdictModerator{v: v}, effects)

			out, err := svc.CastToWind(context.Background(), uuid.New(), uuid.New())
			if err != nil {
				t.Fatalf("CastToWind: %v", err)
			}
			if effects.pooled {
				t.Error("非安全なのにプールへ投入された")
			}
			if !effects.rejected || effects.rejectedVer != v {
				t.Errorf("却下が記録されていない: rejected=%v ver=%v", effects.rejected, effects.rejectedVer)
			}
			if out.ShowCrisisSupport {
				t.Error("他者害/児童で危機支援を案内した（判定が漏れる）")
			}
		})
	}
}

// 危機（自傷）は配信しないが、本人にだけ支援先を案内する。
func TestCastToWind_危機は支援先を案内(t *testing.T) {
	recs := &fakeRecords{rec: record.Record{Body: "x", KoWritten: 20}}
	effects := &fakeEffects{}
	svc := exchange.NewCastService(recs, verdictModerator{v: moderation.Crisis}, effects)

	out, err := svc.CastToWind(context.Background(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("CastToWind: %v", err)
	}
	if effects.pooled {
		t.Error("危機なのにプールへ投入された")
	}
	if !effects.rejected || effects.rejectedVer != moderation.Crisis {
		t.Error("危機が却下として記録されていない")
	}
	if !out.ShowCrisisSupport {
		t.Error("危機なのに支援先を案内していない")
	}
}

// 他人の記録（見つからない）は record.ErrNotFound を伝播し、プールしない。
func TestCastToWind_NotFoundPropagates(t *testing.T) {
	recs := &fakeRecords{err: record.ErrNotFound}
	effects := &fakeEffects{}
	svc := exchange.NewCastService(recs, moderation.AllPass{}, effects)

	if _, err := svc.CastToWind(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, record.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
	if effects.pooled {
		t.Error("見つからないのにプールへ投入された")
	}
}
