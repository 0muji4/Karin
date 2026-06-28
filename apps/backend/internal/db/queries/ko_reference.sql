-- name: GetKoReference :one
-- 候番号からメタ（名称・読み・意味・節気・季節）を引く。
SELECT ko, name, kana, meaning, sekki, season
FROM ko_reference
WHERE ko = $1;

-- name: ListKoReference :many
-- 全 72 候のメタを候番号順に返す（文箱の候別表示でメタを引くため）。
SELECT ko, name, kana, meaning, sekki, season
FROM ko_reference
ORDER BY ko;
