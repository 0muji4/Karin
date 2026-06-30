# デプロイ手順（Phase 1・無料／記録中心）

Neon（マネージド PostgreSQL・無料枠）＋ Render（Web サービス・無料）にバックエンドを上げる。
交換の自動配信（matcher cron）と関門（Vertex AI）は有料・追加設定のため、ここでは無効。
記録・今日の候・文箱、および交換の「風に乗せる（プール投入）」までが本番で動く。

構成は `render.yaml`（Blueprint）と `Dockerfile`（server/matcher/migrate を 1 イメージに同梱）で定義済み。

## 前提
- Neon アカウント、Render アカウント（どちらも GitHub ログイン可・無料）。
- ローカルに Go（マイグレーションを手で流すため）。

## 1. Neon（DB）
1. Neon でプロジェクトを作成（リージョンは Render に近い Singapore 推奨）。
2. 接続文字列（`DATABASE_URL`）を控える。末尾に `?sslmode=require` を付ける。
   例: `postgresql://USER:PASS@ep-xxx.ap-southeast-1.aws.neon.tech/karin?sslmode=require`

## 2. マイグレーション（手動・無料プランは preDeploy が使えないため）
スキーマと七十二候 72 件の seed まで一度に入る（migration 4 本）。
```sh
cd apps/backend
DATABASE_URL='<Neon の接続文字列>' go run ./cmd/migrate up
```
スキーマ変更時は、その都度デプロイ前に同じコマンドを流す。

## 3. Render（Web サービス）
1. Render に GitHub リポジトリ（`0muji4/Karin`）を接続する。
2. New → Blueprint → リポジトリを選ぶ。`render.yaml` を読み、`karin-server`（無料 Web）が作られる。
   （cron `karin-matcher` は `render.yaml` でコメントアウト済みなので作られない。）
3. `karin-server` の環境変数で `DATABASE_URL`（Secret）に Neon の接続文字列を設定する。
   `KARIN_LLM_*` は空のまま＝関門は AllPass。`KARIN_HTTP_ADDR` は `:10000`（render.yaml の既定）。
4. デプロイ。ヘルスチェックは `/healthz`。

## 4. 動作確認
```sh
BASE=https://<サービス名>.onrender.com
curl -s "$BASE/healthz"                          # 200
curl -s -XPOST "$BASE/accounts"                  # 匿名アカウント発行（token が返る）
```
無料プランは無通信が続くとスリープし、初回アクセスでコールドスタート（数十秒）する。

## 5. Android を本番に向ける
配布ビルドは `apps/mobile/local.properties` の接続先を Render の HTTPS にする：
```
karinBaseUrl=https://<サービス名>.onrender.com/
```
（開発時の `adb reverse`＋`localhost` は不要になる。HTTPS なので平文許可も不要。）

## 交換・関門を本番で開くとき（後日）
- `render.yaml` の `karin-matcher`（cron）のコメントを外し、Render の有料プランで再適用する。
  日次でマッチングし、プールの短冊を受信者へ配信する。
- 関門を Vertex AI で有効化する場合は、GCP サービスアカウント JSON を Render の Secret File として
  置き、`KARIN_LLM_PROVIDER=vertex` ほか `KARIN_LLM_*` と `GOOGLE_APPLICATION_CREDENTIALS` を設定する
  （server・matcher 共通）。詳細は `render.yaml` 冒頭のコメント参照。
- 有料プランに上げたら `render.yaml` の `preDeployCommand`（migrate up）も有効化でき、
  デプロイ前に自動でマイグレーションが走る。
