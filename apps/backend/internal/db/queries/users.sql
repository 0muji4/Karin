-- name: CreateUser :one
-- 匿名アカウントを 1 件作る（role は既定の member）。
INSERT INTO users DEFAULT VALUES
RETURNING id, created_at, role, reputation, suspended_at;

-- name: GetUserByID :one
SELECT id, created_at, role, reputation, suspended_at
FROM users
WHERE id = $1;

-- name: AdjustReputation :exec
-- 通報の決着に応じて評判を増減する（別表にせず集約列で持つ）。
UPDATE users
SET reputation = reputation + $2
WHERE id = $1;
