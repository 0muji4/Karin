-- name: CreateReport :one
-- 受け手の通報を記録する（同じ一枚への二重通報は UNIQUE が弾く）。
INSERT INTO report (tanzaku_id, reporter_id, reason, note)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: ResolveReport :exec
-- 再判定の結果で通報を決着させる。
UPDATE report
SET resolution = $2, resolved_at = now()
WHERE id = $1;
