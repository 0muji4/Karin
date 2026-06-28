package api

import (
	"context"
	"net/http"
	"time"
)

type healthBody struct {
	Status string `json:"status"`
}

// handleHealthz は稼働確認。DB へ疎通できれば 200、できなければ 503 を返す。
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if s.db != nil {
		if err := s.db.Ping(ctx); err != nil {
			s.logger.Warn("healthz: DB 疎通に失敗", "error", err)
			writeError(w, s.logger, http.StatusServiceUnavailable, "db_unavailable", "データベースに接続できない")
			return
		}
	}
	writeJSON(w, s.logger, http.StatusOK, healthBody{Status: "ok"})
}
