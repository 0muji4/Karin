-- name: CreateDelivery :exec
-- 配信を記録する。tanzaku_id UNIQUE がユニキャストを、(recipient_id,delivered_on) UNIQUE が1日1回を守る。
INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on)
VALUES ($1, $2, $3);

-- name: ListReceivedByRecipient :many
-- 受け手が受信した一枚を新しい順に返す。author は渡さない（匿名・配信時に NULL 化済み）。
SELECT t.id, t.body, t.ko_written, t.is_official, d.delivered_on, d.kept_at
FROM delivery d
JOIN tanzaku t ON t.id = d.tanzaku_id
WHERE d.recipient_id = $1
ORDER BY d.delivered_on DESC, t.id;

-- name: GetReceivedForKeep :one
-- 文箱にしまう前に、その一枚が本人宛に配信されたものか確認し、本文・候・既存のしまい時刻を引く。
SELECT t.body, t.ko_written, d.kept_at
FROM delivery d
JOIN tanzaku t ON t.id = d.tanzaku_id
WHERE d.tanzaku_id = $1 AND d.recipient_id = $2;

-- name: MarkKept :exec
-- 受信した一枚を文箱にしまった時刻を記録する（未しまいのときだけ）。
UPDATE delivery SET kept_at = now()
WHERE tanzaku_id = $1 AND recipient_id = $2 AND kept_at IS NULL;
