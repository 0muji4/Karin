-- 記録: 匿名アカウント・認証トークン・文箱・七十二候の静的メタ。
-- 不変条件はアプリ型ではなく DB の CHECK / UNIQUE / NOT NULL に置く。

-- 匿名アカウント。個人情報を紐づけない。
-- role: official/system は種まき著者・運営（交換・関門で使う）。reputation は通報の反映で更新する。
CREATE TABLE users (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   timestamptz NOT NULL DEFAULT now(),
    role         text NOT NULL DEFAULT 'member'
                 CHECK (role IN ('member', 'official', 'system')),
    reputation   integer NOT NULL DEFAULT 0,
    suspended_at timestamptz
);
CREATE INDEX idx_users_special ON users (id) WHERE role IN ('official', 'system');

-- Bearer トークン。平文は保存せず hash のみを持つ。複数端末で多対一。
CREATE TABLE auth_token (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash   bytea NOT NULL UNIQUE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz,
    revoked_at   timestamptz
);
CREATE INDEX idx_auth_token_user ON auth_token (user_id);

-- 記録。文箱に収める本人だけの私的な日記。テキストのみ（不変条件6）。
-- owner-only はアプリ層の WHERE owner_id = $current で守る（DB は行単位認可を持たない・不変条件7）。
CREATE TABLE record (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id   uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body       text NOT NULL,
    ko_written smallint NOT NULL CHECK (ko_written BETWEEN 1 AND 72),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
CREATE INDEX idx_record_owner ON record (owner_id, created_at DESC)
    WHERE deleted_at IS NULL;

-- 七十二候の静的メタ（年非依存）。日付境界は持たず Go の天文計算で解決する。
CREATE TABLE ko_reference (
    ko      smallint PRIMARY KEY CHECK (ko BETWEEN 1 AND 72),
    name    text NOT NULL,
    kana    text NOT NULL,
    meaning text NOT NULL,
    sekki   smallint NOT NULL CHECK (sekki BETWEEN 1 AND 24),
    season  text NOT NULL CHECK (season IN ('spring', 'summer', 'autumn', 'winter'))
);
