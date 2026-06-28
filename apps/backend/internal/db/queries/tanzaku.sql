-- name: PoolTanzaku :one
-- 記録の複製を未配信プールへ投入する（風に乗せる）。
INSERT INTO tanzaku (author_id, body, ko_written, is_official)
VALUES ($1, $2, $3, $4)
RETURNING id, author_id, body, ko_written, status, pooled_at, is_official;

-- name: ExpirePooledBefore :execrows
-- 指定時刻より前にプールされた未配信を expired にする（候TTL）。
UPDATE tanzaku
SET status = 'expired'
WHERE status = 'pooled' AND pooled_at < $1;

-- name: PickOldestPooledForRecipient :one
-- 受け手が著者でない最古の未配信を 1 枚、行ロックして引く（同一バッチ内の取り合いを防ぐ）。
SELECT id
FROM tanzaku
WHERE status = 'pooled' AND author_id IS DISTINCT FROM $1
ORDER BY pooled_at ASC, id
FOR UPDATE SKIP LOCKED
LIMIT 1;

-- name: MarkDelivered :exec
-- 配信済みにし、著者を剥がして匿名化する。
UPDATE tanzaku
SET status = 'delivered', author_id = NULL
WHERE id = $1;
