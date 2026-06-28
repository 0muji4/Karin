-- name: CreateRecord :one
-- 文箱に記録を 1 枚保存する。
INSERT INTO record (owner_id, body, ko_written)
VALUES ($1, $2, $3)
RETURNING id, owner_id, body, ko_written, created_at;

-- name: ListRecordsByOwner :many
-- 本人の文箱を候別・新しい順に並べて返す（owner-only: owner_id で必ず絞る）。
SELECT id, owner_id, body, ko_written, created_at
FROM record
WHERE owner_id = $1 AND deleted_at IS NULL
ORDER BY ko_written ASC, created_at DESC;

-- name: GetRecord :one
-- 本人の記録を 1 枚引く（owner_id を必ず条件に入れて他人のものを返さない）。
SELECT id, owner_id, body, ko_written, created_at
FROM record
WHERE id = $1 AND owner_id = $2 AND deleted_at IS NULL;
