package api_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// 風に乗せると、複製が pooled で投入され、著者のクレジットが +1 される。
func TestCast_PoolsAndCredits(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()
	uidStr, token := createAccount(t, h)
	uid := uuid.MustParse(uidStr)

	// 記録を 1 枚作る。
	rr := doJSON(t, h, http.MethodPost, "/records", token, map[string]any{"body": "いちばん長い昼を、何もせず", "ko_written": 29})
	if rr.Code != http.StatusCreated {
		t.Fatalf("記録作成 = %d, want 201", rr.Code)
	}
	var rec recordResp
	mustJSON(t, rr.Body.Bytes(), &rec)

	// 風に乗せる。
	rr = doJSON(t, h, http.MethodPost, "/records/"+rec.ID+"/cast", token, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("cast = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}

	ctx := context.Background()
	// 複製が pooled・著者＝本人で投入されている。
	var pooled int
	if err := testPool.QueryRow(ctx,
		`SELECT count(*) FROM tanzaku WHERE author_id = $1 AND status = 'pooled' AND body = $2`,
		uid, "いちばん長い昼を、何もせず").Scan(&pooled); err != nil {
		t.Fatalf("tanzaku 件数取得: %v", err)
	}
	if pooled != 1 {
		t.Errorf("pooled な tanzaku = %d, want 1", pooled)
	}
	// クレジットが +1。
	var credits int
	if err := testPool.QueryRow(ctx,
		`SELECT receive_credits FROM exchange_ledger WHERE user_id = $1`, uid).Scan(&credits); err != nil {
		t.Fatalf("クレジット取得: %v", err)
	}
	if credits != 1 {
		t.Errorf("クレジット = %d, want 1", credits)
	}
	// 元の記録は文箱に残る。
	var boxA boxResp
	rr = doJSON(t, h, http.MethodGet, "/box", token, nil)
	mustJSON(t, rr.Body.Bytes(), &boxA)
	if len(boxA.Groups) != 1 || len(boxA.Groups[0].Records) != 1 {
		t.Errorf("元の記録が文箱に残っていない: %+v", boxA.Groups)
	}
}

// 存在しない記録・他人の記録を風に乗せようとすると 404。
func TestCast_NotFoundOrNotOwned(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()
	_, tokenA := createAccount(t, h)
	_, tokenB := createAccount(t, h)

	// 存在しない記録 ID。
	if rr := doJSON(t, h, http.MethodPost, "/records/"+uuid.NewString()+"/cast", tokenA, nil); rr.Code != http.StatusNotFound {
		t.Errorf("存在しない記録の cast = %d, want 404", rr.Code)
	}

	// A の記録を B が風に乗せる → owner-only で 404。
	rr := doJSON(t, h, http.MethodPost, "/records", tokenA, map[string]any{"body": "Aの記録", "ko_written": 11})
	var rec recordResp
	mustJSON(t, rr.Body.Bytes(), &rec)
	if rr := doJSON(t, h, http.MethodPost, "/records/"+rec.ID+"/cast", tokenB, nil); rr.Code != http.StatusNotFound {
		t.Errorf("他人の記録の cast = %d, want 404", rr.Code)
	}
}
