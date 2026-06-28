-- name: CreateUser :one
-- 匿名アカウントを 1 件作る（role は既定の member）。
INSERT INTO users DEFAULT VALUES
RETURNING id, created_at, role, reputation, suspended_at;

-- name: GetUserByID :one
SELECT id, created_at, role, reputation, suspended_at
FROM users
WHERE id = $1;
