-- name: CreateChildSafetyAlert :exec
-- 児童保全のホールドを作る（匿名化後も判断材料を失わないよう本文をスナップショット）。
INSERT INTO child_safety_alert (tanzaku_id, author_id, body_snapshot, source_report_id)
VALUES ($1, $2, $3, $4);
