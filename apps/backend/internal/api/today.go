package api

import (
	"net/http"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/ko"
)

// jst は候の判定・表示に使う日本標準時（DST なしなので固定オフセットで十分）。
var jst = time.FixedZone("JST", 9*60*60)

type sekkiJSON struct {
	Number int    `json:"number"`
	Name   string `json:"name"`
	Kana   string `json:"kana"`
}

type wafuMonthJSON struct {
	Name string `json:"name"`
	Kana string `json:"kana"`
}

type todayResponse struct {
	Date      string        `json:"date"`
	WafuMonth wafuMonthJSON `json:"wafu_month"`
	Sekki     sekkiJSON     `json:"sekki"`
	Ko        koJSON        `json:"ko"`
}

// handleTodayKo は今日にあたる七十二候と、その二十四節気・和風月名を返す（一般情報・認証不要）。
// 候番号は太陽黄経から計算し、候名などのメタは ko_reference から、節気名・和風月名は ko の暦知識から得る。
func (s *Server) handleTodayKo(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	n := ko.Number(now)

	meta, err := s.ko.Get(r.Context(), n)
	if err != nil {
		s.logger.Error("候メタの取得に失敗", "ko", n, "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "今日の候を取得できなかった")
		return
	}

	sek := ko.Sekki(ko.SekkiOf(n))
	wm := ko.WafuMonthOf(now)
	writeJSON(w, s.logger, http.StatusOK, todayResponse{
		Date:      now.In(jst).Format("2006-01-02"),
		WafuMonth: wafuMonthJSON{Name: wm.Name, Kana: wm.Kana},
		Sekki:     sekkiJSON{Number: sek.Number, Name: sek.Name, Kana: sek.Kana},
		Ko:        toKoJSON(meta),
	})
}
