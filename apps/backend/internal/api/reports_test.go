package api_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
)

// 通報エンドポイントの認証・入力検証・エラー写像を確かめる（配信の用意は adapter の結合テストで担保）。
func TestCreateReport_認証と入力検証(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()
	_, token := createAccount(t, h)

	t.Run("未認証は 401", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/reports", "", map[string]string{
			"tanzaku_id": uuid.NewString(), "reason": "spam",
		})
		if rr.Code != http.StatusUnauthorized {
			t.Errorf("code=%d, want 401", rr.Code)
		}
	})

	t.Run("不正な理由は 400", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/reports", token, map[string]string{
			"tanzaku_id": uuid.NewString(), "reason": "bogus",
		})
		if rr.Code != http.StatusBadRequest {
			t.Errorf("code=%d, want 400", rr.Code)
		}
	})

	t.Run("受け取っていない一枚は 404", func(t *testing.T) {
		rr := doJSON(t, h, http.MethodPost, "/reports", token, map[string]string{
			"tanzaku_id": uuid.NewString(), "reason": "harassment",
		})
		if rr.Code != http.StatusNotFound {
			t.Errorf("code=%d, want 404", rr.Code)
		}
	})
}
