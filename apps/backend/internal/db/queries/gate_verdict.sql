-- name: RecordGateVerdict :exec
-- 判定を監査ログに残す（著者には露出しない）。
INSERT INTO gate_verdict (subject_kind, subject_id, verdict, model, raw)
VALUES ($1, $2, $3, $4, $5);
