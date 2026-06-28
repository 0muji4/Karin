// Package vertex は Vertex AI 上の Gemini を関門の LLMClient として使うアダプタ。
// 認証は ADC（Application Default Credentials = サービスアカウント等）。データは学習に使われず、
// リージョン(location)を固定できるため、機微なコンテンツの判定に向く。
// 実呼び出しは本番でのみ起き、CI では呼ばない（seam は moderation 側で fake によりテスト済み）。
package vertex

import (
	"context"
	"fmt"

	"google.golang.org/genai"

	"github.com/0muji4/Karin/apps/backend/internal/config"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
)

// Client は moderation.LLMClient を Vertex AI 上の Gemini で満たす。
type Client struct {
	gc    *genai.Client
	model string
}

// NewClient は Vertex AI バックエンドの genai クライアントを作る（認証は ADC）。
func NewClient(ctx context.Context, project, location, model string) (*Client, error) {
	gc, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  project,
		Location: location,
	})
	if err != nil {
		return nil, fmt.Errorf("Vertex AI クライアントの初期化に失敗: %w", err)
	}
	return &Client{gc: gc, model: model}, nil
}

// Complete は分類プロンプトを Gemini に送り、生のテキスト応答を返す。
// プロンプト構築と応答の解釈は moderation.LLMModerator が担い、本メソッドは送受信だけを行う。
func (c *Client) Complete(ctx context.Context, prompt string) (string, error) {
	resp, err := c.gc.Models.GenerateContent(ctx, c.model, genai.Text(prompt), nil)
	if err != nil {
		return "", fmt.Errorf("Vertex AI 呼び出しに失敗: %w", err)
	}
	return resp.Text(), nil
}

// NewModerator は LLM 設定から関門を組み立てる合成ヘルパ。
// Provider が空なら AllPass（LLM 無しでも記録・交換は動く）、"vertex" なら Vertex AI の
// LLMModerator を返す。設定の検証は config.Load が済ませている前提。
func NewModerator(ctx context.Context, cfg config.LLMConfig) (moderation.Moderator, error) {
	if cfg.Provider == "" {
		return moderation.AllPass{}, nil
	}
	if cfg.Provider != "vertex" {
		return nil, fmt.Errorf("未対応の LLM provider: %s", cfg.Provider)
	}
	client, err := NewClient(ctx, cfg.Project, cfg.Location, cfg.Model)
	if err != nil {
		return nil, err
	}
	return moderation.NewLLMModerator(client, cfg.Model), nil
}
