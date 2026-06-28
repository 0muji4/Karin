// Package config loads runtime configuration from the environment.
//
// 設計方針: 設定は小さく平坦なので、ライブラリを使わず標準の os だけで読む。
// 必須項目が欠けたら起動を止める（DB を指さないままサーバを上げない）。
// テスト容易性のため、環境取得は lookup 関数で注入できるようにしている。
package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Config はバックエンド全体の実行時設定。
type Config struct {
	// DatabaseURL は PostgreSQL の接続文字列（必須）。
	DatabaseURL string
	// HTTPAddr は API サーバの待受アドレス。既定は ":8080"。
	HTTPAddr string
	// KoTTL は未配信の短冊を expired にするまでの候数（N候）。1 以上。
	KoTTL int
	// LLM は出口の関門（モデレーション）で使う外部 LLM 設定。未設定でも記録・交換は動く。
	LLM LLMConfig
}

// LLMConfig は関門が呼ぶ外部 LLM プロバイダの設定。
type LLMConfig struct {
	Provider string
	APIKey   string
	Model    string
}

const (
	defaultHTTPAddr = ":8080"
	defaultKoTTL    = 6 // 約30日（候は約5日）。運用で調整する。
)

// Load はプロセス環境から設定を読む。必須項目が欠けるとエラーを返す。
func Load(lookup func(string) (string, bool)) (Config, error) {
	var cfg Config
	var missing []string

	if v, ok := lookup("DATABASE_URL"); ok && v != "" {
		cfg.DatabaseURL = v
	} else {
		missing = append(missing, "DATABASE_URL")
	}

	cfg.HTTPAddr = defaultHTTPAddr
	if v, ok := lookup("KARIN_HTTP_ADDR"); ok && v != "" {
		cfg.HTTPAddr = v
	}

	cfg.KoTTL = defaultKoTTL
	if v, ok := lookup("KARIN_KO_TTL"); ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return Config{}, fmt.Errorf("KARIN_KO_TTL は整数で指定する: %q: %w", v, err)
		}
		if n < 1 {
			return Config{}, fmt.Errorf("KARIN_KO_TTL は 1 以上にする: %d", n)
		}
		cfg.KoTTL = n
	}

	// LLM 設定は任意。provider を指定したときは key/model も要る。
	cfg.LLM.Provider, _ = lookupTrim(lookup, "KARIN_LLM_PROVIDER")
	cfg.LLM.APIKey, _ = lookupTrim(lookup, "KARIN_LLM_API_KEY")
	cfg.LLM.Model, _ = lookupTrim(lookup, "KARIN_LLM_MODEL")
	if cfg.LLM.Provider != "" {
		if cfg.LLM.APIKey == "" {
			missing = append(missing, "KARIN_LLM_API_KEY")
		}
		if cfg.LLM.Model == "" {
			missing = append(missing, "KARIN_LLM_MODEL")
		}
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("%w: %s", ErrMissing, strings.Join(missing, ", "))
	}
	return cfg, nil
}

// ErrMissing は必須項目欠落を表す番兵エラー（呼び出し側の分岐用）。
var ErrMissing = errors.New("必須の環境変数が未設定")

func lookupTrim(lookup func(string) (string, bool), key string) (string, bool) {
	v, ok := lookup(key)
	return strings.TrimSpace(v), ok
}
