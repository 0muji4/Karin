-- 七十二候のメタを取り除く（候参照テーブル自体は記録のスキーマ側で落とす）。
DELETE FROM ko_reference;
