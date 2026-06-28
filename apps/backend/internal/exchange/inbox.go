package exchange

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrNotReceived は、その一枚が本人宛に配信されていないときに返る。
var ErrNotReceived = errors.New("受信した一枚が見つからない")

// ReceivedCard は受け手が受信した一枚（風だより）。送り主は含めない（匿名）。
type ReceivedCard struct {
	TanzakuID   uuid.UUID
	Body        string
	Ko          int
	IsOfficial  bool
	DeliveredOn time.Time
	Kept        bool
}

// Inbox は受信（風だより）の読み取りと「文箱にしまう」のポート。
type Inbox interface {
	// ListReceived は本人が受信した一枚を新しい順に返す。
	ListReceived(ctx context.Context, recipientID uuid.UUID) ([]ReceivedCard, error)
	// Keep は受信した一枚を文箱に複製してしまう（冪等）。本人宛でなければ ErrNotReceived。
	Keep(ctx context.Context, recipientID, tanzakuID uuid.UUID) error
}
