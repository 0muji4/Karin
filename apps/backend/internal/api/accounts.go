package api

import "net/http"

type createAccountResponse struct {
	UserID string `json:"user_id"`
	// Token は生成直後に 1 度だけ返す。サーバは hash しか保持しない。
	Token string `json:"token"`
}

// handleCreateAccount は匿名アカウントを作り、Bearer トークンを返す。
func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	userID, token, err := s.auth.CreateAccount(r.Context())
	if err != nil {
		s.logger.Error("アカウント作成に失敗", "error", err)
		writeError(w, s.logger, http.StatusInternalServerError, "internal", "アカウントを作成できなかった")
		return
	}
	writeJSON(w, s.logger, http.StatusCreated, createAccountResponse{
		UserID: userID.String(),
		Token:  token,
	})
}
