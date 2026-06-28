package moderation

import (
	"context"
	"errors"
	"testing"
)

// fakeClient は LLMClient を差し替えるテスト用スタブ（実 API を叩かない）。
type fakeClient struct {
	resp string
	err  error
}

func (f fakeClient) Complete(context.Context, string) (string, error) {
	return f.resp, f.err
}

func TestLLMModerator_四分岐への写像(t *testing.T) {
	cases := []struct {
		name string
		resp string
		want Verdict
	}{
		{"safe", `{"verdict":"safe","reason":"季節の言葉"}`, Safe},
		{"harm", `{"verdict":"harm","reason":"嫌がらせ"}`, HarmToOthers},
		{"crisis", `{"verdict":"crisis","reason":"自傷"}`, Crisis},
		{"child", `{"verdict":"child","reason":"児童"}`, Child},
		{"コードフェンス付き", "```json\n{\"verdict\":\"safe\",\"reason\":\"ok\"}\n```", Safe},
		{"前後に説明文が混ざる", "判定します。\n{\"verdict\":\"harm\",\"reason\":\"x\"}\n以上です。", HarmToOthers},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod := NewLLMModerator(fakeClient{resp: tc.resp}, "fake-model")
			d, err := mod.Review(context.Background(), "本文")
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if d.Verdict != tc.want {
				t.Errorf("Verdict=%v, want %v", d.Verdict, tc.want)
			}
			if d.Provider != "fake-model" {
				t.Errorf("Provider=%q, want fake-model", d.Provider)
			}
		})
	}
}

func TestLLMModerator_failClosed(t *testing.T) {
	// 送受信エラー・解釈不能・未知ラベルは、いずれも error にして呼び手が配信を止められるようにする。
	cases := []struct {
		name   string
		client LLMClient
	}{
		{"送受信エラー", fakeClient{err: errors.New("timeout")}},
		{"JSON でない応答", fakeClient{resp: "わかりません"}},
		{"未知ラベル", fakeClient{resp: `{"verdict":"maybe","reason":"?"}`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod := NewLLMModerator(tc.client, "fake-model")
			if _, err := mod.Review(context.Background(), "本文"); err == nil {
				t.Error("error を返すべき場面で nil（fail-closed が崩れる）")
			}
		})
	}
}
