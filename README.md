# new-nfl-library

NFL の選手・ロスター・スターター情報を管理する API サーバー（Go + Gin + PostgreSQL）。

技術スタック・レイヤー構成・コーディング規約は [`CLAUDE.md`](./CLAUDE.md) を、各ファイルの役割は
[`docs/project-structure.md`](./docs/project-structure.md) を参照。

## 前提条件

- Go 1.25 以上
- Docker / Docker Compose（PostgreSQL をコンテナで動かす）
- `make tools` で以下を `$(go env GOPATH)/bin` にインストール
  - `air`（ホットリロード）
  - `sqlc`（SQL からの Go コード生成）
  - `oapi-codegen`（OpenAPI からの型生成）
  - `migrate`（golang-migrate CLI）
  - `golangci-lint`

```sh
make tools
```

## 初回セットアップ

```sh
cp .env.example .env   # 必要に応じて値を編集する
make up                # API + PostgreSQL を docker compose で起動（ホットリロード）
make migrate-up        # マイグレーション適用
make seed              # マスタデータ投入（teams / positions / formations など）
make import-data       # db/data の CSV（roster/starter/madden）を一括投入
```

`make seed` で root ユーザーを作成する場合は `.env` に `SEED_ROOT_EMAIL` / `SEED_ROOT_PASSWORD` を設定しておく（未設定でも他のマスタデータは投入される）。

## マイグレーションをフレッシュにしてデータを投入し直す手順

スキーマを変更した後や、DB を完全に作り直したいときの手順。**`migrate-down-all` は全テーブルを削除するので、残しておきたいデータがある場合は事前に確認すること。**

1. PostgreSQL が起動していることを確認する

   ```sh
   make up            # 未起動なら起動（docker compose up -d --build）
   ```

2. マイグレーションを全て巻き戻してから、最新まで適用し直す

   ```sh
   "$(go env GOPATH)/bin/migrate" -path db/migrations \
     -database "postgres://app:app@localhost:5432/app?sslmode=disable" down -all
   make migrate-up
   ```

   `migrate` コマンドは `make tools` で `$(go env GOPATH)/bin` にインストールされるが、シェルの `PATH` には
   入っていないことが多いので、上記のようにフルパスで呼ぶ（`PATH` に追加済みなら `migrate ...` でよい）。
   `DATABASE_URL` を `.env` や環境変数で変えている場合は、1行目の接続文字列もそれに合わせる。

3. マスタデータを投入する（`force=1` で既存データを TRUNCATE してから入れ直す）

   ```sh
   SEED_ROOT_EMAIL=root@example.com SEED_ROOT_PASSWORD=password make seed force=1
   ```

   投入されるテーブル: `teams`（32件）、`positions`、`formations`、`pf_relations`、`users`（root）、`settings`、`foul_rules`。

4. `db/data` の CSV から players / rosters / starters / Madden評価値 を一括投入する

   ```sh
   make import-data
   ```

   2012〜2025年の全シーズン分を `rosters → starters → madden` の順に投入する。特定のシーズンや
   種類だけに絞りたい場合:

   ```sh
   make import-data type=rosters                 # rosters だけ（全シーズン）
   make import-data type=starters season=2024     # starters だけ、2024年のみ
   make import-data type=madden season=2024       # Madden評価値だけ、2024年のみ
   ```

   `type=all`（既定）のときは starters・madden が対応する roster を検索するため、必ず rosters を
   先に投入する。`db/data` に含まれていない季節や個別の CSV ファイルを取り込みたい場合は、単体コマンド
   `make import-rosters csv=path` / `make import-madden [season=2024] [dir=path]` も使える
   （`make import-madden` は season 省略時に `db/data/madden` 配下の全シーズンを単独でも一括処理できる）。

5. 投入結果を確認する

   ```sh
   docker compose exec db psql -U app -d app -c \
     "SELECT (SELECT count(*) FROM players) players, (SELECT count(*) FROM rosters) rosters, (SELECT count(*) FROM starters) starters, (SELECT count(*) FILTER (WHERE madden_attributes IS NOT NULL) FROM rosters) with_madden;"
   ```

## よく使うコマンド

| コマンド | 内容 |
|---|---|
| `make up` / `make down` / `make logs` | docker compose の起動・停止・API ログ |
| `make run` / `make dev` | ローカル起動（`.env` を読む）。`dev` は air によるホットリロード |
| `make test` / `make test-integration` | ユニットテスト / 統合テスト（Postgres が必要） |
| `make lint` / `make fmt` | golangci-lint による静的解析 / フォーマット |
| `make gen` / `make sqlc` | OpenAPI から型生成 / SQL から Go コード生成 |
| `make migrate-up` / `make migrate-down` | マイグレーション適用 / 1つ戻す |
| `make migrate-new name=xxx` | マイグレーションファイル作成 |
| `make seed [force=1]` | マスタデータ投入 |
| `make import-data [type=rosters\|starters\|madden\|all] [season=2024] [strict=1]` | `db/data` の CSV を一括投入 |
| `make import-rosters csv=path [strict=1]` | 単一の roster CSV から players / rosters を生成 |
| `make import-madden [season=2024] [dir=path] [strict=1]` | Madden NFL 評価CSVで rosters を更新（season 省略時は全シーズン一括） |
| `make scrape-rosters season=2024 [teams=buf,nwe]` | pro-football-reference.com からロスターをスクレイプ |
| `make scrape-starters season=2024 [teams=buf,nwe]` | 同上、スターター |

各コマンドの詳しい挙動は [`docs/project-structure.md`](./docs/project-structure.md) を参照。
