package config

import (
	"errors"
	"testing"
)

// envOf は map ベースの lookup を作る。os.Environ を汚さずにテストできる。
func envOf(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoad_minimal(t *testing.T) {
	cfg, err := Load(envOf(map[string]string{
		"DATABASE_URL": "postgres://localhost/karin",
	}))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if cfg.DatabaseURL != "postgres://localhost/karin" {
		t.Errorf("DatabaseURL = %q", cfg.DatabaseURL)
	}
	if cfg.HTTPAddr != defaultHTTPAddr {
		t.Errorf("HTTPAddr 既定値 = %q, want %q", cfg.HTTPAddr, defaultHTTPAddr)
	}
	if cfg.KoTTL != defaultKoTTL {
		t.Errorf("KoTTL 既定値 = %d, want %d", cfg.KoTTL, defaultKoTTL)
	}
}

func TestLoad_missingDatabaseURL(t *testing.T) {
	_, err := Load(envOf(map[string]string{}))
	if !errors.Is(err, ErrMissing) {
		t.Fatalf("DATABASE_URL 欠落で ErrMissing を期待: %v", err)
	}
}

func TestLoad_koTTL(t *testing.T) {
	tests := []struct {
		name    string
		val     string
		want    int
		wantErr bool
	}{
		{"正の値", "10", 10, false},
		{"非整数", "abc", 0, true},
		{"ゼロは不可", "0", 0, true},
		{"負は不可", "-1", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(envOf(map[string]string{
				"DATABASE_URL": "postgres://localhost/karin",
				"KARIN_KO_TTL": tt.val,
			}))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("エラーを期待したが nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if cfg.KoTTL != tt.want {
				t.Errorf("KoTTL = %d, want %d", cfg.KoTTL, tt.want)
			}
		})
	}
}

func TestLoad_vertexRequiresProjectLocationModel(t *testing.T) {
	// provider=vertex を指定したら project/location/model も必須。
	_, err := Load(envOf(map[string]string{
		"DATABASE_URL":       "postgres://localhost/karin",
		"KARIN_LLM_PROVIDER": "vertex",
	}))
	if !errors.Is(err, ErrMissing) {
		t.Fatalf("vertex のみで ErrMissing を期待: %v", err)
	}

	cfg, err := Load(envOf(map[string]string{
		"DATABASE_URL":       "postgres://localhost/karin",
		"KARIN_LLM_PROVIDER": "vertex",
		"KARIN_LLM_MODEL":    "gemini-x",
		"KARIN_LLM_PROJECT":  "my-gcp-proj",
		"KARIN_LLM_LOCATION": "asia-northeast1",
	}))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}
	if cfg.LLM.Provider != "vertex" || cfg.LLM.Model != "gemini-x" ||
		cfg.LLM.Project != "my-gcp-proj" || cfg.LLM.Location != "asia-northeast1" {
		t.Errorf("LLM 設定が読めていない: %+v", cfg.LLM)
	}
}

func TestLoad_unknownLLMProviderErrors(t *testing.T) {
	// 未対応の provider は ErrMissing ではなく明確な設定エラーにする。
	_, err := Load(envOf(map[string]string{
		"DATABASE_URL":       "postgres://localhost/karin",
		"KARIN_LLM_PROVIDER": "anthropic",
	}))
	if err == nil || errors.Is(err, ErrMissing) {
		t.Fatalf("未対応 provider で設定エラーを期待: %v", err)
	}
}

func TestLoad_llmAbsentIsOK(t *testing.T) {
	// LLM 未設定でも記録・交換の機能は動くべき。
	cfg, err := Load(envOf(map[string]string{
		"DATABASE_URL": "postgres://localhost/karin",
	}))
	if err != nil {
		t.Fatalf("LLM 未設定でエラーになった: %v", err)
	}
	if cfg.LLM.Provider != "" {
		t.Errorf("LLM.Provider は空のはず: %q", cfg.LLM.Provider)
	}
}
