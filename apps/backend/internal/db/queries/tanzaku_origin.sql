-- name: CreateTanzakuOrigin :exec
-- 著者特定の私的リンクを作る（プール時・運営/モデレーションのみ参照、受け手向けに JOIN しない）。
INSERT INTO tanzaku_origin (tanzaku_id, author_id, source_record_id)
VALUES ($1, $2, $3);

-- name: GetOriginForReview :one
-- 再判定/保全のために一枚の本文と著者を引く（受け手向けには使わない運営専用クエリ）。
SELECT t.id, t.body, t.status, o.author_id, o.source_record_id
FROM tanzaku t
JOIN tanzaku_origin o ON o.tanzaku_id = t.id
WHERE t.id = $1;
