package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/exchange"
)

type receivedCardJSON struct {
	TanzakuID   string `json:"tanzaku_id"`
	Body        string `json:"body"`
	Ko          int    `json:"ko"`
	IsOfficial  bool   `json:"is_official"`
	DeliveredOn string `json:"delivered_on"`
	Kept        bool   `json:"kept"`
}

type deliveriesResponse struct {
	Received []receivedCardJSON `json:"received"`
}

// handleListDeliveries は本人が受信した一枚（風だより）を新しい順に返す。送り主は含めない（匿名）。
func (s *Server) handleListDeliveries(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}
	cards, err := s.inbox.ListReceived(r.Context(), p.UserID)
	if err != nil {
		s.logger.Error("受信一覧の取得に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "風だよりを取得できなかった")
		return
	}
	out := make([]receivedCardJSON, 0, len(cards))
	for _, c := range cards {
		out = append(out, receivedCardJSON{
			TanzakuID:   c.TanzakuID.String(),
			Body:        c.Body,
			Ko:          c.Ko,
			IsOfficial:  c.IsOfficial,
			DeliveredOn: c.DeliveredOn.Format("2006-01-02"),
			Kept:        c.Kept,
		})
	}
	writeJSON(w, s.logger, http.StatusOK, deliveriesResponse{Received: out})
}

// handleKeep は受信した一枚を文箱にしまう（複製を記録として保存）。本人宛でなければ 404。
func (s *Server) handleKeep(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_id", "ID が不正")
		return
	}
	if err := s.inbox.Keep(r.Context(), p.UserID, id); err != nil {
		if errors.Is(err, exchange.ErrNotReceived) {
			writeError(w, s.logger, http.StatusNotFound, "not_found", "受信した一枚が見つからない")
			return
		}
		s.logger.Error("文箱にしまうのに失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "文箱にしまえなかった")
		return
	}
	writeJSON(w, s.logger, http.StatusOK, castResponse{Status: "kept"})
}
