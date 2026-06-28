package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/0muji4/Karin/apps/backend/internal/ko"
	"github.com/0muji4/Karin/apps/backend/internal/record"
)

type createRecordRequest struct {
	Body string `json:"body"`
	// KoWritten は省略可。省略時は今日の候を使う。
	KoWritten *int `json:"ko_written"`
}

type recordJSON struct {
	ID        string    `json:"id"`
	KoWritten int       `json:"ko_written"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func toRecordJSON(r record.Record) recordJSON {
	return recordJSON{
		ID:        r.ID.String(),
		KoWritten: r.KoWritten,
		Body:      r.Body,
		CreatedAt: r.CreatedAt,
	}
}

// handleCreateRecord は本人の文箱に短冊を保存する。候の指定が無ければ今日の候を使う。
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}

	var req createRecordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_json", "リクエストの形式が不正")
		return
	}

	koNum := ko.Number(time.Now())
	if req.KoWritten != nil {
		koNum = *req.KoWritten
	}

	rec, err := s.records.Create(r.Context(), p.UserID, req.Body, koNum)
	if errors.Is(err, record.ErrInvalid) {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_record", err.Error())
		return
	}
	if err != nil {
		s.logger.Error("記録の保存に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "記録を保存できなかった")
		return
	}
	writeJSON(w, s.logger, http.StatusCreated, toRecordJSON(rec))
}

type boxGroup struct {
	WafuMonth wafuMonthJSON `json:"wafu_month"`
	Sekki     sekkiJSON     `json:"sekki"`
	Records   []recordJSON  `json:"records"`
}

type boxResponse struct {
	Groups []boxGroup `json:"groups"`
}

// handleListBox は本人の文箱を二十四節気ごとにまとめて返す（節気昇順、各節気内は新しい順）。
// 節気名・和風月名は ko の暦知識から得る（候メタの DB 参照は不要）。
func (s *Server) handleListBox(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}

	recs, err := s.records.ListByOwner(r.Context(), p.UserID)
	if err != nil {
		s.logger.Error("文箱の取得に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "文箱を取得できなかった")
		return
	}

	// recs は候昇順に並ぶ＝節気昇順。節気が変わる境目でグループを切る。
	groups := make([]boxGroup, 0)
	for i := 0; i < len(recs); {
		sek := ko.SekkiOf(recs[i].KoWritten)
		meta := ko.Sekki(sek)
		wm := ko.WafuMonthOf(recs[i].CreatedAt)
		g := boxGroup{
			WafuMonth: wafuMonthJSON{Name: wm.Name, Kana: wm.Kana},
			Sekki:     sekkiJSON{Number: meta.Number, Name: meta.Name, Kana: meta.Kana},
			Records:   []recordJSON{},
		}
		for i < len(recs) && ko.SekkiOf(recs[i].KoWritten) == sek {
			g.Records = append(g.Records, toRecordJSON(recs[i]))
			i++
		}
		groups = append(groups, g)
	}
	writeJSON(w, s.logger, http.StatusOK, boxResponse{Groups: groups})
}
