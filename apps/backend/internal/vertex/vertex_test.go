package vertex_test

import (
	"context"
	"testing"

	"github.com/0muji4/Karin/apps/backend/internal/config"
	"github.com/0muji4/Karin/apps/backend/internal/moderation"
	"github.com/0muji4/Karin/apps/backend/internal/vertex"
)

// LLM 未設定なら関門は AllPass（記録・交換は LLM 無しで動く）。
func TestNewModerator_未設定はAllPass(t *testing.T) {
	m, err := vertex.NewModerator(context.Background(), config.LLMConfig{})
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if _, ok := m.(moderation.AllPass); !ok {
		t.Errorf("AllPass を期待: %T", m)
	}
}

// 未対応の provider は、実クライアント生成に進む前に明確なエラーにする。
func TestNewModerator_未対応providerはエラー(t *testing.T) {
	if _, err := vertex.NewModerator(context.Background(), config.LLMConfig{Provider: "anthropic"}); err == nil {
		t.Error("未対応 provider でエラーを期待")
	}
}
