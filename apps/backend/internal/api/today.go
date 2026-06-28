package api

import (
	"net/http"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/db/sqlcdb"
	"github.com/0muji4/Karin/apps/backend/internal/ko"
)

// jst は候の判定・表示に使う日本標準時（DST なしなので固定オフセットで十分）。
var jst = time.FixedZone("JST", 9*60*60)

type todayResponse struct {
	Date string `json:"date"`
	Ko   koJSON `json:"ko"`
}

// handleTodayKo は今日にあたる七十二候とそのメタを返す（一般情報・認証不要）。
// 候番号は太陽黄経から計算し、名称などのメタは ko_reference から引く。
func (s *Server) handleTodayKo(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	n := ko.Number(now)

	meta, err := sqlcdb.New(s.pool).GetKoReference(r.Context(), int16(n))
	if err != nil {
		s.logger.Error("候メタの取得に失敗", "ko", n, "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "今日の候を取得できなかった")
		return
	}

	writeJSON(w, s.logger, http.StatusOK, todayResponse{
		Date: now.In(jst).Format("2006-01-02"),
		Ko:   toKoJSON(meta),
	})
}
