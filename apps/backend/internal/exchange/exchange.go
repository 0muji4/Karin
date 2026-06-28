// Package exchange は交換(L2)のドメインとユースケースを担う。
// 未配信プールへの投入・日次マッチャによる配信・互酬クレジットを扱う。
// 永続化は Pool / MatchStore ポートに委ね、この層は具体 DB を知らない。
package exchange

import (
	"time"

	"github.com/google/uuid"
)

// Status は短冊の状態。DB の CHECK と一致させる。
type Status string

const (
	StatusPooled    Status = "pooled"
	StatusDelivered Status = "delivered"
	StatusExpired   Status = "expired"
)

// Tanzaku は交換に出された一枚（記録 L1 から風に乗せた複製）。
// AuthorID は配信時に剥がされる（受け手には渡さない）。
type Tanzaku struct {
	ID         uuid.UUID
	AuthorID   uuid.UUID
	Body       string
	Ko         int
	Status     Status
	PooledAt   time.Time
	IsOfficial bool
}
