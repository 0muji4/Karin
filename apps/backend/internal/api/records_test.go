package api_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/ko"
)

type recordResp struct {
	ID        string    `json:"id"`
	KoWritten int       `json:"ko_written"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type boxResp struct {
	Groups []struct {
		Ko struct {
			Number  int    `json:"number"`
			Name    string `json:"name"`
			Kana    string `json:"kana"`
			Meaning string `json:"meaning"`
			Sekki   int    `json:"sekki"`
			Season  string `json:"season"`
		} `json:"ko"`
		Records []recordResp `json:"records"`
	} `json:"groups"`
}

// 記録の核心条件: 別アカウントから他人の文箱を読めない（owner-only）。
func TestBox_OwnerOnly(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()

	_, tokenA := createAccount(t, h)
	_, tokenB := createAccount(t, h)

	for _, koNum := range []int{10, 29} {
		rr := doJSON(t, h, http.MethodPost, "/records", tokenA, map[string]any{"body": "Aの記録", "ko_written": koNum})
		if rr.Code != http.StatusCreated {
			t.Fatalf("A の記録作成 = %d, want 201 (body: %s)", rr.Code, rr.Body.String())
		}
	}

	var boxA boxResp
	rr := doJSON(t, h, http.MethodGet, "/box", tokenA, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("A の文箱取得 = %d, want 200", rr.Code)
	}
	mustJSON(t, rr.Body.Bytes(), &boxA)
	if len(boxA.Groups) != 2 {
		t.Fatalf("A の文箱グループ数 = %d, want 2", len(boxA.Groups))
	}
	if boxA.Groups[0].Ko.Number != 10 || boxA.Groups[1].Ko.Number != 29 {
		t.Errorf("候の並びが昇順でない: %d, %d", boxA.Groups[0].Ko.Number, boxA.Groups[1].Ko.Number)
	}
	if boxA.Groups[1].Ko.Name != "菖蒲華" {
		t.Errorf("候29 の name = %q, want 菖蒲華", boxA.Groups[1].Ko.Name)
	}

	var boxB boxResp
	rr = doJSON(t, h, http.MethodGet, "/box", tokenB, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("B の文箱取得 = %d, want 200", rr.Code)
	}
	mustJSON(t, rr.Body.Bytes(), &boxB)
	if len(boxB.Groups) != 0 {
		t.Fatalf("B の文箱は空のはず: groups=%d (他人の記録が漏れている)", len(boxB.Groups))
	}
}

// 認証なし・不正トークンは 401。
func TestEndpoints_AuthRequired(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()

	if rr := doJSON(t, h, http.MethodGet, "/box", "", nil); rr.Code != http.StatusUnauthorized {
		t.Errorf("トークンなしの /box = %d, want 401", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodGet, "/box", "deadbeef-not-a-real-token", nil); rr.Code != http.StatusUnauthorized {
		t.Errorf("不正トークンの /box = %d, want 401", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodPost, "/records", "", map[string]any{"body": "x"}); rr.Code != http.StatusUnauthorized {
		t.Errorf("トークンなしの /records = %d, want 401", rr.Code)
	}
}

// 候を省略すると今日の候が使われる。
func TestRecord_DefaultsToTodayKo(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()
	_, token := createAccount(t, h)

	rr := doJSON(t, h, http.MethodPost, "/records", token, map[string]any{"body": "候を省略した記録"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("記録作成 = %d, want 201", rr.Code)
	}
	var rec recordResp
	mustJSON(t, rr.Body.Bytes(), &rec)
	if want := ko.Number(time.Now()); rec.KoWritten != want {
		t.Errorf("既定の候 = %d, want 今日の候 %d", rec.KoWritten, want)
	}
}

// 不正な入力（候範囲外・空本文）は 400。
func TestRecord_Validation(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()
	_, token := createAccount(t, h)

	if rr := doJSON(t, h, http.MethodPost, "/records", token, map[string]any{"body": "x", "ko_written": 99}); rr.Code != http.StatusBadRequest {
		t.Errorf("候99 の記録 = %d, want 400", rr.Code)
	}
	if rr := doJSON(t, h, http.MethodPost, "/records", token, map[string]any{"body": "   "}); rr.Code != http.StatusBadRequest {
		t.Errorf("空本文の記録 = %d, want 400", rr.Code)
	}
}

// mustJSON は応答ボディを構造体へデコードする共有ヘルパ。
func mustJSON(t *testing.T, data []byte, dst any) {
	t.Helper()
	if err := json.Unmarshal(data, dst); err != nil {
		t.Fatalf("JSON デコードに失敗: %v (body: %s)", err, string(data))
	}
}
