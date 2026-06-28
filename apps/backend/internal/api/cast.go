package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/record"
)

type castResponse struct {
	Status string `json:"status"`
}

// handleCast は本人の記録を風に乗せる（複製を未配信プールへ投入）。
// 応答は一律で、プールしたか否かの判定結果は著者に見せない（M3 の四分岐に備える）。
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

	if err := s.cast.CastToWind(r.Context(), p.UserID, id); err != nil {
		if errors.Is(err, record.ErrNotFound) {
			writeError(w, s.logger, http.StatusNotFound, "not_found", "記録が見つからない")
			return
		}
		s.logger.Error("風に乗せるのに失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "風に乗せられなかった")
		return
	}
	writeJSON(w, s.logger, http.StatusOK, castResponse{Status: "cast"})
}
