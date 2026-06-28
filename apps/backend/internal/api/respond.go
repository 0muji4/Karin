// Package api はクライアント向けの JSON HTTP API（ルーティング・middleware・ハンドラ）を提供する。
package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// errorBody は API のエラー応答の形。
// 関門の判定など、内部の理由は載せない（不変条件: 著者にフィルタの手がかりを与えない）。
type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// writeJSON は値を JSON で書き出す。エンコード失敗はログに流して握り潰さない。
func writeJSON(w http.ResponseWriter, logger *slog.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil && logger != nil {
		logger.Error("JSON 応答の書き出しに失敗", "error", err)
	}
}

// writeError は機械可読な code と人向けの message でエラーを返す。
func writeError(w http.ResponseWriter, logger *slog.Logger, status int, code, message string) {
	writeJSON(w, logger, status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

// decodeJSON はリクエストボディを厳格にデコードする（未知フィールドを拒否）。
func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}
