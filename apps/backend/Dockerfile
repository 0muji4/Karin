# syntax=docker/dockerfile:1
#
# 夏鈴バックエンドのコンテナイメージ。1 イメージに server / matcher / migrate を同梱し、
# 起動コマンドの違いで使い分ける（Web サービス=server、日次 cron=matcher、デプロイ前=migrate up）。
# クラウド非依存: アプリが知るのは DATABASE_URL ほかの環境変数のみ。

# ---- build stage ----
FROM golang:1.26-bookworm AS build
WORKDIR /src
# 依存を先に解決してレイヤキャッシュを効かせる。
COPY go.mod go.sum ./
RUN go mod download
# ソース一式（db/migrations を含む＝//go:embed のため）。
COPY . .
# pgx は純 Go。CGO 無効で完全静的にし、distroless/static に載せる。
ENV CGO_ENABLED=0
RUN go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && go build -trimpath -ldflags="-s -w" -o /out/matcher ./cmd/matcher \
 && go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---- runtime stage ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/server /out/matcher /out/migrate /usr/local/bin/
USER nonroot:nonroot
EXPOSE 8080
# 既定は API サーバ。ENTRYPOINT ではなく CMD にして、用途で command を完全に置き換えられるようにする
# （compose の command や Render の dockerCommand で matcher/migrate に切り替えるため）:
#   /usr/local/bin/migrate up   … デプロイ前のマイグレーション
#   /usr/local/bin/matcher      … 日次マッチング（cron）
CMD ["/usr/local/bin/server"]
