package api_test

import "testing"

// POST /accounts が匿名アカウントを作り、毎回別の user_id / token を返す。
func TestAccount_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()

	uid1, token1 := createAccount(t, h)
	uid2, token2 := createAccount(t, h)

	if uid1 == uid2 {
		t.Errorf("別アカウントのはずが user_id が重複: %s", uid1)
	}
	if token1 == token2 {
		t.Errorf("別アカウントのはずが token が重複している")
	}
}
