package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/0muji4/Karin/apps/backend/internal/auth"
	"github.com/0muji4/Karin/apps/backend/internal/dbtest"
)

func TestAuth_CreateAndAuthenticate(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	ctx := context.Background()
	pool, _, terminate, err := dbtest.MigratedPool(ctx)
	if err != nil {
		t.Fatalf("PG 起動失敗: %v", err)
	}
	defer terminate()
	svc := auth.NewService(pool)

	// 発行したトークンで認証でき、Principal が一致する。
	uid, token, err := svc.CreateAccount(ctx)
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	p, err := svc.Authenticate(ctx, token)
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if p.UserID != uid {
		t.Errorf("UserID = %v, want %v", p.UserID, uid)
	}
	if p.Role != "member" {
		t.Errorf("Role = %q, want member", p.Role)
	}
	if p.Suspended {
		t.Errorf("新規アカウントが suspended になっている")
	}

	// 未知トークンは ErrUnauthorized。
	if _, err := svc.Authenticate(ctx, "totally-unknown-token"); !errors.Is(err, auth.ErrUnauthorized) {
		t.Errorf("未知トークン: err = %v, want ErrUnauthorized", err)
	}
	// 空トークンも ErrUnauthorized。
	if _, err := svc.Authenticate(ctx, ""); !errors.Is(err, auth.ErrUnauthorized) {
		t.Errorf("空トークン: err = %v, want ErrUnauthorized", err)
	}

	// 2 つのアカウントは別人で、トークンも別。
	uid2, token2, err := svc.CreateAccount(ctx)
	if err != nil {
		t.Fatalf("2 件目の CreateAccount: %v", err)
	}
	if uid2 == uid {
		t.Errorf("UserID が重複している")
	}
	if token2 == token {
		t.Errorf("トークンが重複している")
	}
	p2, err := svc.Authenticate(ctx, token2)
	if err != nil || p2.UserID != uid2 {
		t.Errorf("2 件目の認証に失敗: p=%+v err=%v", p2, err)
	}
}
