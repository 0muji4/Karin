package report_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/report"
)

// fakeStore は report.Store の DB 不要なテスト実装。
type fakeStore struct {
	submitErr error
	subj      report.Subject
	resolved  bool
	gotOut    report.Outcome
}

func (f *fakeStore) Submit(_ context.Context, tanzakuID, _ uuid.UUID, _, _ string) (uuid.UUID, report.Subject, error) {
	if f.submitErr != nil {
		return uuid.Nil, report.Subject{}, f.submitErr
	}
	f.subj.TanzakuID = tanzakuID
	return uuid.New(), f.subj, nil
}

func (f *fakeStore) Resolve(_ context.Context, out report.Outcome) error {
	f.resolved = true
	f.gotOut = out
	return nil
}

type verdictModerator struct{ v moderation.Verdict }

func (m verdictModerator) Review(context.Context, string) (moderation.Decision, error) {
	return moderation.Decision{Verdict: m.v, Reason: "test"}, nil
}

type errModerator struct{}

func (errModerator) Review(context.Context, string) (moderation.Decision, error) {
	return moderation.Decision{}, errors.New("llm down")
}

// 通報は記録され本文が再判定され、判定に応じた決着・評判・保全の方針が Resolve に渡る。
func TestReport_再判定の結果を反映する(t *testing.T) {
	cases := []struct {
		v           moderation.Verdict
		wantRes     report.Resolution
		wantDelta   int
		wantChild   bool
	}{
		{moderation.Safe, report.ResolutionDismissed, 0, false},
		{moderation.HarmToOthers, report.ResolutionUpheld, -1, false},
		{moderation.Child, report.ResolutionUpheld, -1, true},
	}
	for _, tc := range cases {
		t.Run(tc.v.Label(), func(t *testing.T) {
			store := &fakeStore{subj: report.Subject{AuthorID: uuid.New(), Body: "本文"}}
			svc := report.NewService(verdictModerator{v: tc.v}, store)

			if err := svc.Report(context.Background(), uuid.New(), uuid.New(), "harassment", ""); err != nil {
				t.Fatalf("Report: %v", err)
			}
			if !store.resolved {
				t.Fatal("再判定の結果が反映されていない")
			}
			out := store.gotOut
			if out.Verdict != tc.v || out.Resolution != tc.wantRes || out.ReputationDelta != tc.wantDelta || out.ChildSafety != tc.wantChild {
				t.Errorf("方針が不正: %+v (want res=%s delta=%d child=%v)", out, tc.wantRes, tc.wantDelta, tc.wantChild)
			}
		})
	}
}

// 再判定できないときは決着させない（通報は記録済み）。応答はエラーにしない。
func TestReport_再判定不能なら決着させない(t *testing.T) {
	store := &fakeStore{subj: report.Subject{Body: "本文"}}
	svc := report.NewService(errModerator{}, store)

	if err := svc.Report(context.Background(), uuid.New(), uuid.New(), "spam", ""); err != nil {
		t.Fatalf("再判定不能時はエラーにしない: %v", err)
	}
	if store.resolved {
		t.Error("再判定できないのに決着させた")
	}
}

// 受信していない一枚への通報は ErrNotReceived を伝播し、再判定しない。
func TestReport_未受信は弾く(t *testing.T) {
	store := &fakeStore{submitErr: report.ErrNotReceived}
	svc := report.NewService(verdictModerator{v: moderation.Safe}, store)

	if err := svc.Report(context.Background(), uuid.New(), uuid.New(), "spam", ""); !errors.Is(err, report.ErrNotReceived) {
		t.Errorf("err = %v, want ErrNotReceived", err)
	}
	if store.resolved {
		t.Error("未受信なのに再判定・決着した")
	}
}
