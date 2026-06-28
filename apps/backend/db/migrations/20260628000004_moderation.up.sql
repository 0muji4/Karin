-- M3 モデレーション関門: 出口の安全判定とその副産物を保持する。
-- 設計の軸:
--   * 匿名性と保全/評判の両立 …… 著者特定は tanzaku_origin にだけ残し、受け手向けクエリには JOIN しない。
--     配信時に tanzaku.author_id を NULL 化しても、通報→評判・児童保全に必要な著者はここから辿れる。
--   * fail-closed …… 判定が安全と確定しない限り配信しない。エラー/曖昧は pending_submission に保留する。
--   * 著者に判定を見せない …… 判定の監査は gate_verdict にのみ残し、著者応答には出さない。

-- 著者特定の私的リンク。プール時に作成し、運営/モデレーションだけが参照する（受け手には決して開かない）。
CREATE TABLE tanzaku_origin (
    tanzaku_id       uuid PRIMARY KEY REFERENCES tanzaku(id),
    author_id        uuid NOT NULL REFERENCES users(id),
    source_record_id uuid REFERENCES record(id),  -- 風に乗せた元の記録（削除されうるので NULL 許容）
    created_at       timestamptz NOT NULL DEFAULT now()
);

-- fail-closed の保留。安全と確定するまで配信せず、ここで待たせる。
-- body は投入時のスナップショット（記録の編集/削除と切り離し、判定対象を固定する）。
CREATE TABLE pending_submission (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id           uuid NOT NULL REFERENCES users(id),
    source_record_id    uuid REFERENCES record(id),
    body                text NOT NULL,
    ko_written          smallint NOT NULL CHECK (ko_written BETWEEN 1 AND 72),
    status              text NOT NULL DEFAULT 'awaiting'
                        CHECK (status IN ('awaiting', 'pooled', 'rejected')),
    attempts            integer NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    last_error          text,
    created_at          timestamptz NOT NULL DEFAULT now(),
    resolved_at         timestamptz,
    promoted_tanzaku_id uuid REFERENCES tanzaku(id)  -- safe で昇格したときに作られた一枚
);
-- 復旧後に再判定する対象を引くための部分索引。
CREATE INDEX idx_pending_awaiting ON pending_submission (created_at) WHERE status = 'awaiting';

-- 受け手による通報の不変ログ。自動削除はせず、再判定と評判反映の根拠にする。
CREATE TABLE report (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tanzaku_id  uuid NOT NULL REFERENCES tanzaku(id),
    reporter_id uuid NOT NULL REFERENCES users(id),
    reason      text NOT NULL
                CHECK (reason IN ('harassment', 'sexual', 'self_harm', 'child_safety', 'spam', 'other')),
    note        text,
    created_at  timestamptz NOT NULL DEFAULT now(),
    resolved_at timestamptz,
    resolution  text CHECK (resolution IN ('upheld', 'dismissed', 'escalated')),
    UNIQUE (tanzaku_id, reporter_id)  -- 同じ一枚を同じ人が二重に通報しない
);

-- 児童保全のホールド。関門で弾かれて tanzaku が存在しない場合もあるため tanzaku_id は NULL 許容。
-- 匿名化後も判断材料を失わないよう本文をスナップショットで保持する。
CREATE TABLE child_safety_alert (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tanzaku_id       uuid REFERENCES tanzaku(id),
    author_id        uuid NOT NULL REFERENCES users(id),
    body_snapshot    text NOT NULL,
    source_report_id uuid REFERENCES report(id),  -- 通報起点なら参照、関門起点なら NULL
    created_at       timestamptz NOT NULL DEFAULT now(),
    operator_ack_at  timestamptz,
    operator_id      uuid REFERENCES users(id)
);

-- 判定の監査。著者には露出しない（fail-closed の検証と運営の振り返り用）。
CREATE TABLE gate_verdict (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_kind text NOT NULL CHECK (subject_kind IN ('pending', 'tanzaku')),
    subject_id   uuid NOT NULL,
    verdict      text NOT NULL
                 CHECK (verdict IN ('safe', 'harm', 'crisis', 'child', 'ambiguous', 'llm_error')),
    model        text,
    raw          jsonb,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_gate_verdict_subject ON gate_verdict (subject_kind, subject_id);
