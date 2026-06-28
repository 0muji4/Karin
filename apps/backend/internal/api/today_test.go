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
		Date string `json:"date"`
		Ko   struct {
			Number int    `json:"number"`
			Name   string `json:"name"`
			Season string `json:"season"`
		} `json:"ko"`
	}
	mustJSON(t, rr.Body.Bytes(), &resp)
	if want := ko.Number(time.Now()); resp.Ko.Number != want {
		t.Errorf("今日の候 = %d, want %d", resp.Ko.Number, want)
	}
	if resp.Ko.Name == "" || resp.Ko.Season == "" {
		t.Errorf("候メタが空: %+v", resp.Ko)
	}
	if resp.Date == "" {
		t.Errorf("日付が空")
	}
}
