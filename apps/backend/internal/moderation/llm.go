package moderation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// LLMClient は文脈を読む LLM への問い合わせを抽象化する provider 非依存の seam。
// プロンプト構築と応答の解釈は moderation（ポリシー）が持ち、本 IF は送受信だけを担う。
// 実アダプタ（Gemini/Vertex AI 等）は別パッケージで実装し、テストには fake を注入する。
type LLMClient interface {
	// Complete は分類プロンプトを送り、モデルの生のテキスト応答を返す。
	Complete(ctx context.Context, prompt string) (string, error)
}

// LLMModerator は LLMClient を使って本文を四分岐に分類する Moderator 実装。
type LLMModerator struct {
	client LLMClient
	model  string // 監査・記録に残すモデル識別子
}

// NewLLMModerator は LLMClient とモデル識別子から関門を作る。
func NewLLMModerator(client LLMClient, model string) *LLMModerator {
	return &LLMModerator{client: client, model: model}
}

// Review は本文を分類する。送受信エラーも応答の解釈不能も error にして返し、
// 呼び手（exchange）が fail-closed（配信せず保留）に倒せるようにする。
func (m *LLMModerator) Review(ctx context.Context, body string) (Decision, error) {
	raw, err := m.client.Complete(ctx, buildPrompt(body))
	if err != nil {
		return Decision{}, fmt.Errorf("関門 LLM 呼び出し: %w", err)
	}
	v, reason, err := parseVerdict(raw)
	if err != nil {
		return Decision{}, fmt.Errorf("関門 LLM 応答の解釈: %w", err)
	}
	return Decision{Verdict: v, Reason: reason, Provider: m.model}, nil
}

// buildPrompt は四分岐の分類指示を組み立てる。季語・詩的表現の文脈を読ませ、
// 「枯れる」「散る」「終わり」など季節の自然な比喩を危機と取り違えないようにする。
// 安全と確信できなければ最も保護的な非安全側へ倒す（fail-safe）。
func buildPrompt(body string) string {
	return `あなたは短い日本語の季節の言葉を匿名で交換するアプリの安全判定者です。
次の本文を、以下のいずれか 1 つに分類してください。

- safe: 配信してよい。
- harm: 他者への害（嫌がらせ・差別・個人情報の暴露・性的に露骨）。
- crisis: 投稿者自身の危機・自傷・自殺念慮。
- child: 児童に関わる性的・搾取的な内容。

季語や詩的表現の文脈を考慮し、季節の自然な比喩（枯れる・散る・終わり等）を
危機と取り違えないこと。安全だと確信できない場合は、最も保護的な非安全側（多くは harm）に倒すこと。

出力は次の JSON のみ（前後に説明やコードフェンスを付けない）:
{"verdict":"safe|harm|crisis|child","reason":"日本語で簡潔に"}

本文:
` + body
}

// parseVerdict はモデルの応答を四分岐へ写像する。応答に説明文やコードフェンスが
// 混ざっても拾えるよう、最初の '{' から最後の '}' までを JSON として読む。
// 未知ラベルや解釈不能は error にして fail-closed に倒す。
func parseVerdict(raw string) (Verdict, string, error) {
	start := strings.IndexByte(raw, '{')
	end := strings.LastIndexByte(raw, '}')
	if start < 0 || end < start {
		return 0, "", fmt.Errorf("JSON オブジェクトが見つからない: %q", raw)
	}
	var out struct {
		Verdict string `json:"verdict"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw[start:end+1]), &out); err != nil {
		return 0, "", fmt.Errorf("JSON として読めない: %w", err)
	}
	switch strings.ToLower(strings.TrimSpace(out.Verdict)) {
	case "safe":
		return Safe, out.Reason, nil
	case "harm":
		return HarmToOthers, out.Reason, nil
	case "crisis":
		return Crisis, out.Reason, nil
	case "child":
		return Child, out.Reason, nil
	default:
		return 0, "", fmt.Errorf("未知の判定ラベル: %q", out.Verdict)
	}
}
