package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/record"
)

type castResponse struct {
	Status  string       `json:"status"`
	Support *supportInfo `json:"support,omitempty"`
}

// supportInfo は危機（自傷）と判定したときにだけ本人へ返す支援先の案内。
type supportInfo struct {
	Message string `json:"message"`
	URL     string `json:"url"`
}

// handleCast は本人の記録を風に乗せる。応答は一律で、プールしたか否かの判定結果は著者に見せない。
// 例外として、危機（自傷）と判定したときだけ本人に支援先を案内する。
func (s *Server) handleCast(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_id", "記録 ID が不正")
		return
	}

	outcome, err := s.cast.CastToWind(r.Context(), p.UserID, id)
	if err != nil {
		if errors.Is(err, record.ErrNotFound) {
			writeError(w, s.logger, http.StatusNotFound, "not_found", "記録が見つからない")
			return
		}
		s.logger.Error("風に乗せるのに失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "風に乗せられなかった")
		return
	}

	resp := castResponse{Status: "cast"}
	if outcome.ShowCrisisSupport {
		resp.Support = &supportInfo{
			Message: "つらい気持ちが続くときは、ひとりで抱えこまないでください。相談できる窓口があります。",
			URL:     "https://www.mhlw.go.jp/mamorouyokokoro/",
		}
	}
	writeJSON(w, s.logger, http.StatusOK, resp)
}
