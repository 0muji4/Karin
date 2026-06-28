// Package auth は匿名アカウントと Bearer トークン認証のユースケース・ドメインを担う。
// アカウントに個人情報を紐づけない。トークンは平文を保存せず hash だけを扱う。
// 永続化は Repository ポートに委ね、この層は具体 DB を知らない（依存性ルール）。
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// ErrUnauthorized はトークンが無効・未知のときに返る。
var ErrUnauthorized = errors.New("認証に失敗")

// tokenBytes は Bearer トークンの乱数長（バイト）。
const tokenBytes = 32

// Principal は認証済みの呼び出し主体。
type Principal struct {
	UserID    uuid.UUID
	Role      string
	Suspended bool
}

// Repository はアカウントとトークンの永続化ポート。
// CreateAccountWithToken はアカウント作成とトークン保存をひとつの不可分な操作として行う。
// FindByActiveToken は失効していないトークンの所有者を返し、無ければ ErrUnauthorized を返す。
type Repository interface {
	CreateAccountWithToken(ctx context.Context, tokenHash []byte) (uuid.UUID, error)
	FindByActiveToken(ctx context.Context, tokenHash []byte) (Principal, error)
	TouchToken(ctx context.Context, tokenHash []byte) error
}

// Service は匿名アカウントの発行と認証ユースケース。
type Service struct {
	repo Repository
}

// NewService は Repository ポートから Service を作る。
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CreateAccount は匿名アカウントを 1 つ作り、生のトークンを 1 度だけ返す。
func (s *Service) CreateAccount(ctx context.Context) (userID uuid.UUID, token string, err error) {
	token, hash, err := newToken()
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("トークン生成に失敗: %w", err)
	}
	userID, err = s.repo.CreateAccountWithToken(ctx, hash)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("アカウント作成に失敗: %w", err)
	}
	return userID, token, nil
}

// Authenticate は提示トークンから Principal を返す。未知・失効・空なら ErrUnauthorized。
func (s *Service) Authenticate(ctx context.Context, token string) (Principal, error) {
	if token == "" {
		return Principal{}, ErrUnauthorized
	}
	hash := hashToken(token)
	p, err := s.repo.FindByActiveToken(ctx, hash)
	if err != nil {
		return Principal{}, err
	}
	// 最終利用時刻の更新は best-effort。失敗しても認証自体は通す。
	_ = s.repo.TouchToken(ctx, hash)
	return p, nil
}

// newToken は乱数トークン（base64url 文字列）とその SHA-256 hash を返す。
func newToken() (token string, hash []byte, err error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}
	token = base64.RawURLEncoding.EncodeToString(b)
	return token, hashToken(token), nil
}

// hashToken は保存・照合に使うトークンの SHA-256 を返す。
func hashToken(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}
