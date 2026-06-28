package record_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/record"
)

// fakeRepo は Repository ポートの DB 不要なテスト実装。
type fakeRepo struct {
	createCalled bool
	getErr       error
}

func (f *fakeRepo) Create(_ context.Context, ownerID uuid.UUID, body string, ko int) (record.Record, error) {
	f.createCalled = true
	return record.Record{ID: uuid.New(), Body: body, KoWritten: ko}, nil
}

func (f *fakeRepo) ListByOwner(context.Context, uuid.UUID) ([]record.Record, error) {
	return nil, nil
}

func (f *fakeRepo) Get(context.Context, uuid.UUID, uuid.UUID) (record.Record, error) {
	return record.Record{}, f.getErr
}

// 検証は永続化の前に行われ、不正入力は repo に到達しない（DB 不要の単体テスト）。
func TestService_CreateValidation(t *testing.T) {
	owner := uuid.New()
	bad := []struct {
		name string
		body string
		ko   int
	}{
		{"空本文", "   ", 11},
		{"候0", "x", 0},
		{"候73", "x", 73},
		{"本文超過", strings.Repeat("あ", record.MaxBodyRunes+1), 11},
	}
	for _, b := range bad {
		t.Run(b.name, func(t *testing.T) {
			repo := &fakeRepo{}
			svc := record.NewService(repo)
			if _, err := svc.Create(context.Background(), owner, b.body, b.ko); !errors.Is(err, record.ErrInvalid) {
				t.Errorf("err = %v, want ErrInvalid", err)
			}
			if repo.createCalled {
				t.Errorf("不正入力なのに repo.Create が呼ばれた")
			}
		})
	}

	// 妥当な入力は検証を通り、repo に委譲される。
	repo := &fakeRepo{}
	svc := record.NewService(repo)
	if _, err := svc.Create(context.Background(), owner, "  桜が咲いた  ", 11); err != nil {
		t.Fatalf("妥当な入力で失敗: %v", err)
	}
	if !repo.createCalled {
		t.Errorf("妥当な入力なのに repo.Create が呼ばれていない")
	}
}

// Get はポートの ErrNotFound をそのまま返す。
func TestService_GetPassesThroughNotFound(t *testing.T) {
	svc := record.NewService(&fakeRepo{getErr: record.ErrNotFound})
	if _, err := svc.Get(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, record.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}
