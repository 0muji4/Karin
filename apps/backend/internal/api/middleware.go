package api

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder は応答ステータスを記録するための薄い ResponseWriter ラッパ。
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// recoverMiddleware は panic を 500 に変換し、原因をログに流す（握り潰さない）。
func recoverMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				logger.Error("ハンドラで panic", "recover", rec, "path", r.URL.Path)
				writeError(w, logger, http.StatusInternalServerError, "internal", "内部エラーが発生した")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// logMiddleware は各リクエストの結果を構造化ログに記録する。
func logMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
