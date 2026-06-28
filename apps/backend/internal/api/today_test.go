package api_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/ko"
)

// 今日の候エンドポイントは計算した候とメタを返す。
func TestKoToday(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()

	rr := doJSON(t, h, http.MethodGet, "/ko/today", "", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("/ko/today = %d, want 200", rr.Code)
	}
	var resp struct {
		Date      string `json:"date"`
		WafuMonth struct {
			Name string `json:"name"`
			Kana string `json:"kana"`
		} `json:"wafu_month"`
		Sekki struct {
			Number int    `json:"number"`
			Name   string `json:"name"`
			Kana   string `json:"kana"`
		} `json:"sekki"`
		Ko struct {
			Number int    `json:"number"`
			Name   string `json:"name"`
			Season string `json:"season"`
		} `json:"ko"`
	}
	mustJSON(t, rr.Body.Bytes(), &resp)

	n := ko.Number(time.Now())
	if resp.Ko.Number != n {
		t.Errorf("今日の候 = %d, want %d", resp.Ko.Number, n)
	}
	if resp.Ko.Name == "" || resp.Ko.Season == "" {
		t.Errorf("候メタが空: %+v", resp.Ko)
	}
	// 節気は候から導いたものと一致し、名称・読みが入る。
	wantSekki := ko.Sekki(ko.SekkiOf(n))
	if resp.Sekki.Number != wantSekki.Number || resp.Sekki.Name != wantSekki.Name || resp.Sekki.Name == "" {
		t.Errorf("節気 = %+v, want %+v", resp.Sekki, wantSekki)
	}
	if resp.WafuMonth.Name == "" || resp.WafuMonth.Kana == "" {
		t.Errorf("和風月名が空: %+v", resp.WafuMonth)
	}
	if resp.Date == "" {
		t.Errorf("日付が空")
	}
}
