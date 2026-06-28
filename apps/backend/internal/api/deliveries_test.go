package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// 受信一覧の取得と「文箱にしまう」を検証する。送り主は応答に出ない（匿名）。
func TestDeliveries_ListAndKeep(t *testing.T) {
	if testing.Short() {
		t.Skip("結合テスト: Docker が要る（-short で除外）")
	}
	h := handlerForTest()
	uidB, tokenB := createAccount(t, h)
	_, tokenA := createAccount(t, h)
	bID := uuid.MustParse(uidB)
	ctx := context.Background()

	// B 宛に配信済みの一枚を用意（マッチャを介さず DB に直接）。配信済みなので author_id は NULL。
	const body = "青梅を漬けた。瓶の中で、夏を待つ"
	var tid uuid.UUID
	if err := testPool.QueryRow(ctx,
		`INSERT INTO tanzaku (author_id, body, ko_written, status) VALUES (NULL, $1, $2, 'delivered') RETURNING id`,
		body, 27).Scan(&tid); err != nil {
		t.Fatalf("tanzaku 用意: %v", err)
	}
	if _, err := testPool.Exec(ctx,
		`INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on) VALUES ($1, $2, $3)`,
		tid, bID, time.Now()); err != nil {
		t.Fatalf("delivery 用意: %v", err)
	}

	type received struct {
		TanzakuID  string `json:"tanzaku_id"`
		Body       string `json:"body"`
		Ko         int    `json:"ko"`
		IsOfficial bool   `json:"is_official"`
		Kept       bool   `json:"kept"`
	}
	type deliveriesResp struct {
		Received []received `json:"received"`
	}

	// B は風だよりで 1 枚受け取る。送り主は応答に含まれない。
	rr := doJSON(t, h, http.MethodGet, "/deliveries", tokenB, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("B の /deliveries = %d, want 200", rr.Code)
	}
	if strings.Contains(rr.Body.String(), "author") {
		t.Error("受信応答に author が含まれている（匿名性違反）")
	}
	var dB deliveriesResp
	mustJSON(t, rr.Body.Bytes(), &dB)
	if len(dB.Received) != 1 || dB.Received[0].Body != body || dB.Received[0].Kept {
		t.Fatalf("B の受信が不正: %+v", dB.Received)
	}

	// A は何も受信していない。
	rr = doJSON(t, h, http.MethodGet, "/deliveries", tokenA, nil)
	var dA deliveriesResp
	mustJSON(t, rr.Body.Bytes(), &dA)
	if len(dA.Received) != 0 {
		t.Errorf("A の受信 = %d, want 0", len(dA.Received))
	}

	// B が文箱にしまう。
	if rr := doJSON(t, h, http.MethodPost, "/deliveries/"+tid.String()+"/keep", tokenB, nil); rr.Code != http.StatusOK {
		t.Fatalf("keep = %d, want 200 (body: %s)", rr.Code, rr.Body.String())
	}
	// 文箱に複製が現れる。
	countBoxRecords := func() int {
		rr := doJSON(t, h, http.MethodGet, "/box", tokenB, nil)
		var box boxResp
		mustJSON(t, rr.Body.Bytes(), &box)
		n := 0
		for _, g := range box.Groups {
			n += len(g.Records)
		}
		return n
	}
	if got := countBoxRecords(); got != 1 {
		t.Fatalf("keep 後の文箱の記録数 = %d, want 1", got)
	}
	// 冪等: もう一度しまっても増えない。
	if rr := doJSON(t, h, http.MethodPost, "/deliveries/"+tid.String()+"/keep", tokenB, nil); rr.Code != http.StatusOK {
		t.Fatalf("2 回目の keep = %d, want 200", rr.Code)
	}
	if got := countBoxRecords(); got != 1 {
		t.Errorf("冪等でない: 2 回 keep 後の記録数 = %d, want 1", got)
	}
	// /deliveries で kept=true になる。
	rr = doJSON(t, h, http.MethodGet, "/deliveries", tokenB, nil)
	mustJSON(t, rr.Body.Bytes(), &dB)
	if !dB.Received[0].Kept {
		t.Error("keep 後も kept=false")
	}

	// 本人宛でない一枚を A がしまおうとすると 404。
	if rr := doJSON(t, h, http.MethodPost, "/deliveries/"+tid.String()+"/keep", tokenA, nil); rr.Code != http.StatusNotFound {
		t.Errorf("他人宛の keep = %d, want 404", rr.Code)
	}
}
