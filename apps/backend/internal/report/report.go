// Package report は、受け手が配信された一枚を通報したときの再判定ユースケースを担う。
// 関門(moderation)で本文を再判定し、結果を Store ポートに反映する（通報の決着・著者評判・児童保全）。
// 受け手に著者を一切開かない匿名性は、Store の実装が tanzaku_origin を受け手向けに JOIN しないことで守る。
package report

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/0muji4/Karin/apps/backend/internal/moderation"
)

// ErrNotReceived は、その一枚を受け取っていない者が通報しようとしたとき返る（任意の id への通報を防ぐ）。
var ErrNotReceived = errors.New("report: その一枚を受け取っていない")

// ErrAlreadyReported は、同じ受け手が同じ一枚を二重に通報したとき返る。
var ErrAlreadyReported = errors.New("report: すでに通報済み")

// Subject は再判定の対象。Body は判定にかける本文、AuthorID は評判・保全に使う著者（運営内のみ）。
type Subject struct {
	TanzakuID uuid.UUID
	AuthorID  uuid.UUID
	Body      string
}

// Outcome は再判定の結果。Store がこれを見て通報の決着・評判・保全を反映する。
type Outcome struct {
	ReportID uuid.UUID
	Subject  Subject
	Verdict  moderation.Verdict
	Reason   string
}

// Store は通報の永続化と決着のポート。匿名性を守る責務（受け手向けに著者を開かない）も実装側にある。
type Store interface {
	// Submit は通報を記録し、再判定に要る対象（本文・著者）を返す。受信していなければ ErrNotReceived、
	// 二重通報なら ErrAlreadyReported を返す。
	Submit(ctx context.Context, tanzakuID, reporterID uuid.UUID, reason, note string) (uuid.UUID, Subject, error)
	// Resolve は再判定の結果を不可分に反映する: 通報の決着・著者評判の増減・（児童なら）児童保全ホールド。
	Resolve(ctx context.Context, out Outcome) error
}

// Service は通報の再判定ユースケース。
type Service struct {
	gate  moderation.Moderator
	store Store
}

// NewService は依存ポートから Service を作る。
func NewService(gate moderation.Moderator, store Store) *Service {
	return &Service{gate: gate, store: store}
}

// Report は受け手の通報を受け、本文を再判定して結果を反映する。
// 再判定できなかった場合は決着させず（通報自体は記録済み・運営/再試行に委ねる）、応答は一律にする。
// 判定そのものは通報者に見せない。
func (s *Service) Report(ctx context.Context, tanzakuID, reporterID uuid.UUID, reason, note string) error {
	reportID, subj, err := s.store.Submit(ctx, tanzakuID, reporterID, reason, note)
	if err != nil {
		return err
	}
	dec, err := s.gate.Review(ctx, subj.Body)
	if err != nil {
		// fail-closed: 確定できないものは決着させない。通報は残り、後で再判定される。
		return nil
	}
	return s.store.Resolve(ctx, Outcome{
		ReportID: reportID,
		Subject:  subj,
		Verdict:  dec.Verdict,
		Reason:   dec.Reason,
	})
}
