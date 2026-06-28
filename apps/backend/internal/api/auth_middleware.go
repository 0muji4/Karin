package api

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/0muji4/Karin/apps/backend/internal/auth"
)

type ctxKey int

const principalKey ctxKey = iota

// requireAuth は Bearer トークンを検証し、Principal を文脈に載せて次へ渡す。
// 未認証は 401、停止中アカウントは 403。
func (s *Server) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "Bearer トークンが必要")
			return
		}
		p, err := s.auth.Authenticate(r.Context(), token)
		if errors.Is(err, auth.ErrUnauthorized) {
			writeError(w, s.logger, http.StatusUnauthorized, "unauthorized", "トークンが無効")
			return
		}
		if err != nil {
			s.logger.Error("認証処理でエラー", "error", err)
			writeError(w, s.logger, http.StatusInternalServerError, "internal", "内部エラーが発生した")
			return
		}
		if p.Suspended {
			writeError(w, s.logger, http.StatusForbidden, "suspended", "このアカウントは利用を停止されている")
			return
		}
		ctx := context.WithValue(r.Context(), principalKey, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// principalFrom は文脈から認証済み Principal を取り出す。
func principalFrom(ctx context.Context) (auth.Principal, bool) {
	p, ok := ctx.Value(principalKey).(auth.Principal)
	return p, ok
}

// bearerToken は Authorization ヘッダから Bearer トークンを取り出す。
func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(h[len(prefix):])
	return token, token != ""
}
