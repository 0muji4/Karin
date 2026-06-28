-- name: IncrementCredit :exec
-- 関門通過で互酬クレジットを +1（無ければ 1 で作る）。
INSERT INTO exchange_ledger (user_id, receive_credits)
VALUES ($1, 1)
ON CONFLICT (user_id) DO UPDATE SET receive_credits = exchange_ledger.receive_credits + 1;

-- name: DecrementCredit :exec
-- 受信でクレジットを −1（非負は CHECK が守る）。
UPDATE exchange_ledger
SET receive_credits = receive_credits - 1
WHERE user_id = $1;

-- name: GetCredits :one
SELECT receive_credits FROM exchange_ledger WHERE user_id = $1;

-- name: ListEligibleRecipients :many
-- 受信資格者（credits>0・当日未配信・未停止）を、割当可能枚数の少ない順に返す。
-- 選択肢の少ない人を先に満たすことで、薄いプールでも公平に配る。
SELECT el.user_id,
       (SELECT count(*) FROM tanzaku t
          WHERE t.status = 'pooled' AND t.author_id IS DISTINCT FROM el.user_id) AS assignable
FROM exchange_ledger el
JOIN users u ON u.id = el.user_id
WHERE el.receive_credits > 0
  AND u.suspended_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM delivery d
      WHERE d.recipient_id = el.user_id AND d.delivered_on = $1
  )
ORDER BY assignable ASC, el.user_id;
