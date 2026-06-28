-- name: CreateAuthToken :one
-- Bearer トークンの hash を保存する（平文は保存しない）。
INSERT INTO auth_token (user_id, token_hash)
VALUES ($1, $2)
RETURNING id, user_id, created_at;

-- name: GetActiveUserByTokenHash :one
-- 失効していないトークン hash から所有ユーザーを引く（認証 middleware が使う）。
SELECT u.id, u.role, u.suspended_at
FROM auth_token t
JOIN users u ON u.id = t.user_id
WHERE t.token_hash = $1 AND t.revoked_at IS NULL;

-- name: TouchAuthToken :exec
UPDATE auth_token
SET last_used_at = now()
WHERE token_hash = $1 AND revoked_at IS NULL;

-- name: RevokeAuthToken :exec
UPDATE auth_token
SET revoked_at = now()
WHERE token_hash = $1;
