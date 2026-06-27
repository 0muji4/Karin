# Contributing to Karin

## ブランチ・マージ方針

`main` の履歴は常に直線（linear history）に保つ。マージコミットは作らない。

### なぜ

- 履歴が一本道になり、変更の順序と各コミットの差分を追いやすい。
- `git bisect` や `git log` が直感的に働き、問題の混入箇所を特定しやすい。
- リバート・チェリーピックの単位が明確になる。

### どうやって

- PR のマージは **Squash and merge** または **Rebase and merge** のみを使う。リポジトリ設定で merge commit は無効化する。
- `main` には「直線履歴必須（require linear history）」の ruleset を適用し、マージコミットを含む更新はサーバー側で拒否する。
- ローカルでも `git pull` / `git merge` が意図せずマージコミットを作らないよう、clone したマシンには次を設定する。

  ```sh
  git config merge.ff only
  git config pull.ff only
  ```

## Issue / PR / コミットの対応規約

1 つの Issue → 1 つの PR → 1 つのコミットを一対一で対応させ、タイトルを一貫させる。CI がこれを検証し、外れた PR はマージできない。

### ルール

- **1 PR = 1 commit**: PR は必ず 1 コミットにまとめる（複数になったら squash する）。
- **コミット件名 = PR タイトル**: コミットの件名は PR タイトルで始め、末尾はピリオドにする。
  - 例: PR タイトル `[App] 今日の候を表示する画面を追加` → コミット件名 `[App] 今日の候を表示する画面を追加.`
- **カテゴリ接頭辞**: Issue・PR のタイトルは `[カテゴリ]` で始める。許可カテゴリは [hack/prefix.yaml](hack/prefix.yaml) を唯一の出どころとして管理する（`App` / `Server` / `Infra` / `Data` / `Docs` / `CI/CD` / `Chore`。複合は `[App/Server]`）。
- **Issue との紐付け**: PR 本文で `Closes #<番号>` 等の closing キーワードを使い、対応 Issue を必ず参照する（マージで Issue が閉じる）。

### カテゴリの使い分け

- `App` … モバイルクライアント（ユーザー向けの体験。Flutter / 各 OS ネイティブ）。
- `Server` … Go バックエンド。API・匿名認証・記録の保存・未配信プール・日次マッチャ・関門の制御・DB スキーマはここに含める。
- `Infra` … クラウド・デプロイ・マネージド PostgreSQL・スケジュール実行など、アプリケーションのロジックの外側の土台。
- `Data` … 七十二候データセット（候の読み・意味・季節の読み物などのコンテンツ）。第0段階の主な作業。
- `Docs` … PRD / DD / README などのドキュメント。
- `CI/CD` … GitHub Actions 等のパイプライン。
- `Chore` … ツール・設定・雑務（`.gitignore`・formatter・依存更新など）。

> モデレーション関門は第2段階（交換と安全機能の実装）まで存在しないため、`Server` に含めて扱い、専用カテゴリは置かない。関門が独立した関心事として育ったら、`hack/prefix.yaml` に 1 行追加する（下記ブートストラップ手順）。

### なぜ

- Issue・PR・コミットのタイトルが一致することで、`git log` だけで「どの Issue のどの対応か」を一目で追える。
- 1 PR = 1 commit により、`main` の各コミットがレビュー済みの変更単位と一致し、revert やリリースノート生成が単純になる。
- closing キーワードで Issue が自動的に閉じ、課題管理と実装履歴の乖離を防ぐ。

### 検証の仕組み

- [.github/workflows/validate_pr.yaml](.github/workflows/validate_pr.yaml) … 上記ルールを PR ごとに検証する（[hack/validate_pr.sh](hack/validate_pr.sh)）。ローカルでも、open な PR があるブランチ上で `sh hack/validate_pr.sh` を実行して事前確認できる。
- [.github/workflows/validate_issue_title.yaml](.github/workflows/validate_issue_title.yaml) … Issue タイトルの接頭辞を検証する（[hack/validate_issue_title.sh](hack/validate_issue_title.sh)）。

## カテゴリを増やすとき（ブートストラップ）

`hack/prefix.yaml` に新カテゴリを追加する PR では、その Issue 作成時点では新カテゴリがまだ `main` に無く、Issue タイトル検証 workflow（`main` を参照）が赤くなる。その PR は**既存カテゴリ**（例 `[CI/CD]` や `[Chore]`）のタイトルで出し、**最初にマージ**する。以降の PR で新カテゴリが使えるようになる。
