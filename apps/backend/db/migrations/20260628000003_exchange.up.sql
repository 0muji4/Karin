-- M2 交換コア: 未配信プール・配信・互酬レジャー。
-- 受信1日1回・ユニキャスト・候の範囲・クレジット非負を、アプリ型ではなく DB の CHECK / UNIQUE で守る。

-- 交換に出された一枚（記録 L1 から風に乗せた複製）。
-- author_id は配信時に NULL へ更新して匿名化する。is_official は種まきを受け手に明示する（偽装しない）。
CREATE TABLE tanzaku (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id   uuid REFERENCES users(id),
    body        text NOT NULL,
    ko_written  smallint NOT NULL CHECK (ko_written BETWEEN 1 AND 72),
    status      text NOT NULL DEFAULT 'pooled'
                CHECK (status IN ('pooled', 'delivered', 'expired')),
    pooled_at   timestamptz NOT NULL DEFAULT now(),
    is_official boolean NOT NULL DEFAULT false
);
-- 最古の未配信を引くための部分索引。
CREATE INDEX idx_tanzaku_pooled ON tanzaku (pooled_at) WHERE status = 'pooled';

-- 配信（割り当て）の記録。著者向けの既読列は持たない（著者は配信を追跡しない）。
CREATE TABLE delivery (
    tanzaku_id   uuid NOT NULL UNIQUE REFERENCES tanzaku(id),  -- 一枚は高々一人（ユニキャスト）
    recipient_id uuid NOT NULL REFERENCES users(id),
    delivered_on date NOT NULL,
    kept_at      timestamptz,                                  -- 受け手が文箱にしまった時刻
    UNIQUE (recipient_id, delivered_on)                        -- 受信は1日1回
);

-- 互酬クレジット。関門通過で +1、受信で −1。非負を CHECK で守る。
CREATE TABLE exchange_ledger (
    user_id         uuid PRIMARY KEY REFERENCES users(id),
    receive_credits integer NOT NULL DEFAULT 0 CHECK (receive_credits >= 0)
);
