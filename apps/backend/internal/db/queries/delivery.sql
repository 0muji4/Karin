-- name: CreateDelivery :exec
-- 配信を記録する。tanzaku_id UNIQUE がユニキャストを、(recipient_id,delivered_on) UNIQUE が1日1回を守る。
INSERT INTO delivery (tanzaku_id, recipient_id, delivered_on)
VALUES ($1, $2, $3);
