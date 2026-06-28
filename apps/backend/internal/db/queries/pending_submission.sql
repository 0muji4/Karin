-- name: CreatePendingSubmission :one
-- 判定が安全と確定しなかった投入を保留する（fail-closed）。body は投入時スナップショット。
INSERT INTO pending_submission (author_id, source_record_id, body, ko_written, last_error)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: ListAwaitingSubmissions :many
-- 復旧後に再判定する保留を古い順に返す。
SELECT id, author_id, source_record_id, body, ko_written, attempts
FROM pending_submission
WHERE status = 'awaiting'
ORDER BY created_at ASC
LIMIT $1;

-- name: MarkPendingPooled :exec
-- 再判定で安全になり昇格した保留を pooled にする。
UPDATE pending_submission
SET status = 'pooled', promoted_tanzaku_id = $2, resolved_at = now()
WHERE id = $1;

-- name: MarkPendingRejected :exec
-- 再判定で不適切と確定した保留を rejected にする。
UPDATE pending_submission
SET status = 'rejected', last_error = $2, resolved_at = now()
WHERE id = $1;

-- name: IncrementPendingAttempt :exec
-- 判定が曖昧/エラーのまま据え置く（試行回数を進め、原因を残す）。
UPDATE pending_submission
SET attempts = attempts + 1, last_error = $2
WHERE id = $1;
