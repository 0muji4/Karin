package api

import (
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/report"
)

type createReportRequest struct {
	TanzakuID string `json:"tanzaku_id"`
	Reason    string `json:"reason"`
	Note      string `json:"note"`
}

// validReportReasons は report.reason の CHECK と一致させる（不正な理由は受け付けず 400 にする）。
var validReportReasons = map[string]bool{
	"harassment": true, "sexual": true, "self_harm": true,
	"child_safety": true, "spam": true, "other": true,
}

// handleCreateReport は受け手が配信された一枚を通報する。本文を再判定し結果を反映する。
// 応答は一律で、再判定の結果（判定や著者）は通報者に見せない。
func (s *Server) handleCreateReport(w http.ResponseWriter, r *http.Request) {
	p, ok := principalFrom(r.Context())
	if !ok {
		writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "認証が必要")
		return
	}

	var req createReportRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_json", "リクエストの形式が不正")
		return
	}
	tanzakuID, err := uuid.Parse(req.TanzakuID)
	if err != nil {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_id", "短冊 ID が不正")
		return
	}
	if !validReportReasons[req.Reason] {
		writeError(w, s.logger, http.StatusBadRequest, "invalid_reason", "通報理由が不正")
		return
	}

	err = s.reports.Report(r.Context(), tanzakuID, p.UserID, req.Reason, req.Note)
	switch {
	case errors.Is(err, report.ErrNotReceived):
		// 受け取っていない一枚は通報できない。存在を漏らさないため 404。
		writeError(w, s.logger, http.StatusNotFound, "not_found", "対象が見つからない")
		return
	case errors.Is(err, report.ErrAlreadyReported):
		writeError(w, s.logger, http.StatusConflict, "already_reported", "すでに通報済み")
		return
	case err != nil:
		s.logger.Error("通報の処理に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "通報を受け付けられなかった")
		return
	}
	writeJSON(w, s.logger, http.StatusOK, map[string]string{"status": "reported"})
}
