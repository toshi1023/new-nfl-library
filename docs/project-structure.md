# プロジェクト構成とファイルの役割

このドキュメントはリポジトリ内の各ファイルが何をしているかを説明する。設計方針そのもの（レイヤーの依存方向・コーディング規約）は `CLAUDE.md` を参照。

## 全体像

```
.
├── cmd/api/main.go                  エントリポイント（DI・起動・graceful shutdown）
├── internal/                        アプリ本体（外部から import 不可）
│   ├── config/                      環境変数の読み込みと検証
│   ├── domain/                      ドメインエラー（今後エンティティ・リポジトリ IF を追加）
│   ├── logging/                     slog ロガーの生成と ctx への格納
│   ├── middleware/                  リクエスト ID・アクセスログ・リカバリ・JWT 認証
│   ├── handler/                     Gin ハンドラ・エラー→HTTP 変換
│   ├── router/                      ルート定義の集約
│   ├── usecase/                     ビジネスロジック（未実装）
│   ├── repository/sqlc/             sqlc 生成コード（手編集禁止）
│   ├── infra/postgres/              pgx 接続プール
│   └── oapi/                        oapi-codegen 生成コード（手編集禁止）
├── api/                             OpenAPI 仕様と生成設定
├── db/migrations/                   golang-migrate 用 SQL
├── db/queries/                      sqlc 用 SQL
├── docs/                            ドキュメント
├── Makefile / Dockerfile / docker-compose.yml
├── .air.toml / .golangci.yml / sqlc.yaml / .env.example
└── go.mod / go.sum
```

## リクエストの流れ

```
HTTP → router (Gin)
        ├─ middleware.RequestID   : X-Request-ID を発行し ctx に格納
        ├─ middleware.Logger      : request_id 付きロガーを ctx に格納、完了時にアクセスログ
        ├─ middleware.Recovery    : panic を捕捉して 500 JSON
        └─ (middleware.Auth)      : /api/v1 配下の認証必須ルートのみ。JWT を検証し subject を ctx に格納
      → handler                   : バインド・バリデーション → usecase 呼び出し → レスポンス整形
      → usecase                   : ビジネスロジック（domain のインターフェースにのみ依存）
      → repository                : domain のインターフェース実装。sqlc 生成コードをラップ
      → infra/postgres            : pgxpool
```

エラーは `domain` のセンチネルエラーを `%w` でラップして上に返し、`handler/response.go` の `classify` が 1 箇所で HTTP ステータスに変換する。

---

## ルート直下

| ファイル | 役割 |
|---|---|
| `go.mod` / `go.sum` | モジュール定義と依存のハッシュ。モジュール名は `github.com/toshi1023/new-nfl-library`、Go 1.25.0 以上。 |
| `Makefile` | 開発コマンドの集約。ツールは `$(GOBIN)/<tool>` の絶対パスで呼ぶ（macOS の GNU Make 3.81 では Makefile 内の PATH 変更が効かないため）。`make lint GOLANGCI_LINT=golangci-lint` のように上書き可能。 |
| `Dockerfile` | 3 ステージ構成。`dev`（air ホットリロード用、compose が使用）、`build`（静的バイナリ作成）、`prod`（distroless 実行イメージ、nonroot）。`GO_VERSION` と `AIR_VERSION` は build-arg で変更できる。 |
| `docker-compose.yml` | `api`（dev ターゲット、ソースをバインドマウント、モジュール/ビルドキャッシュを volume 化）、`db`（PostgreSQL 16、healthcheck 付き）、`migrate`（`tools` プロファイル、`docker compose run --rm migrate up` で使用）。 |
| `.air.toml` | air の設定。`go` ファイルの変更を検知して `./tmp/api` に再ビルド。`db/`、`api/`、生成コードディレクトリ、`_test.go` は監視対象外。 |
| `.golangci.yml` | golangci-lint v2 設定。標準セットに加えて `errorlint`・`forbidigo`（`fmt.Print*` / `log.Print*` を禁止し slog を強制）・`misspell`・`unconvert`・`bodyclose`・`noctx` を有効化。生成コードのディレクトリは除外。フォーマッタは gofmt + goimports（ローカル import をグループ分け）。 |
| `sqlc.yaml` | sqlc 設定。`db/migrations` をスキーマ、`db/queries` をクエリとして読み、`internal/repository/sqlc` に pgx/v5 用コードを生成。`Querier` インターフェースも生成する（usecase のテストでモックしやすい）。 |
| `.env.example` | `make run`（ローカル起動）用の環境変数テンプレート。`.env` にコピーして使う。docker compose では compose 側の `environment` が使われるので不要。 |
| `.gitignore` | `tmp/`（air のビルド出力）、`bin/`、`.env`、`coverage.out` を除外。 |
| `CLAUDE.md` | 技術スタック・レイヤー構成・コーディング規約・新機能追加手順。設計判断の一次情報。 |

---

## `cmd/api/main.go`

プロセスのエントリポイント。ここでやるのは次の 4 つだけで、ビジネスロジックは書かない。

1. `config.Load()` で環境変数を読み、`logging.New` で JSON ロガーを作って `slog.SetDefault`
2. `postgres.NewPool` で DB 接続プールを作成（疎通確認込み。失敗すれば終了コード 1）
3. handler → router の順に手組みで依存を注入（DI ライブラリは使わない）
4. `http.Server` を起動し、SIGINT / SIGTERM を受けたら 10 秒のタイムアウトで graceful shutdown

`APP_ENV` が `development` 以外なら Gin をリリースモードにする。

---

## `internal/config/`

| ファイル | 役割 |
|---|---|
| `config.go` | `Config` 構造体と `Load()`。envconfig で環境変数を読み、`validate()` で値を検証する。検証内容: `APP_ENV` は development / staging / production のいずれか、`APP_PORT` は 1〜65535、`DATABASE_URL` は空不可、`JWT_SECRET` は development 以外で必須、`LOG_LEVEL` は slog が解釈できる値。ヘルパーとして `IsDevelopment()`、`Addr()`（`:8080` 形式）、`SlogLevel()` を持つ。 |
| `config_test.go` | 上記の検証ルールのテスト。envconfig は「空文字で設定された変数」を未設定扱いにしないため、`unsetenv` ヘルパーで本当に未設定にしてテストしている。 |

---

## `internal/domain/`

| ファイル | 役割 |
|---|---|
| `errors.go` | センチネルエラー `ErrNotFound` / `ErrConflict` / `ErrUnauthorized` / `ErrForbidden` / `ErrValidation`。usecase / repository はこれらを `fmt.Errorf("...: %w", domain.ErrNotFound)` のようにラップして返す。HTTP ステータスは意識しない。 |

今後、エンティティ・値オブジェクト・リポジトリインターフェースをこのパッケージに追加する。他の `internal` パッケージを import してはならない。

---

## `internal/logging/`

| ファイル | 役割 |
|---|---|
| `logging.go` | `New(w, level)` で JSON ハンドラの slog ロガーを生成。`WithLogger(ctx, l)` / `FromContext(ctx)` でロガーを ctx 経由で受け渡す。`FromContext` は未格納なら `slog.Default()` を返すので、どの層でも安全に呼べる。 |

`middleware` でも `usecase` でも同じロガーを使うために、`middleware` から独立した小さなパッケージにしている（usecase が middleware を import しなくて済む）。

---

## `internal/middleware/`

| ファイル | 役割 |
|---|---|
| `request_id.go` | `RequestID()`: リクエストヘッダ `X-Request-ID` があれば引き継ぎ、なければ UUID を発行。レスポンスヘッダと ctx に設定する。`RequestIDFromContext(ctx)` で取り出せる。 |
| `logger.go` | `Logger(base)`: `request_id` を付与したロガーを ctx に格納し、ハンドラ完了後にメソッド・パス・ステータス・レイテンシ・クライアント IP・バイト数をアクセスログとして出力。ステータスが 5xx なら Error、4xx なら Warn、それ以外は Info。`RequestID` の後に登録すること。 |
| `recovery.go` | `Recovery()`: panic を捕捉し、スタックトレースをログに残したうえで統一形式の 500 JSON を返す。Gin 標準の `gin.Recovery()` はレスポンス形式が揃わないため独自実装。 |
| `auth.go` | `Auth(secret)`: `Authorization: Bearer <JWT>` を HS256 で検証し、`sub` クレームを ctx に格納。ヘッダ欠落・署名不一致・期限切れ・`sub` 空はすべて 401 の統一 JSON。`SubjectFromContext(ctx)` でユーザー識別子を取り出す。 |
| `auth_test.go` | `Auth` のテスト（有効トークン・ヘッダなし・別鍵署名・期限切れ）。 |

---

## `internal/handler/`

| ファイル | 役割 |
|---|---|
| `response.go` | エラーレスポンスの一元化。エラーコード定数（`VALIDATION_ERROR` など、`api/openapi.yaml` の enum と一致させる）、`ErrorResponse` 構造体、`respondError(c, err)`（`classify` でドメインエラー → HTTP ステータスに変換。未知のエラーは 500 で詳細を隠し、ログにだけ残す）、`respondBindError(c, err)`（`ShouldBindJSON` 等の失敗を 400 に変換）。 |
| `response_test.go` | `classify` の対応表テストと、`respondError` / `respondBindError` を Gin 経由で通すテスト。 |
| `health.go` | `HealthHandler`。`Healthz`（liveness、常に 200）と `Readyz`（readiness、`Pinger` インターフェース経由で DB に 2 秒タイムアウトで ping し、失敗なら 503）。`Pinger` を interface にしているのは、テストでスタブに差し替えるためと、handler が pgx の具象型を知らずに済むようにするため。 |
| `health_test.go` | `Healthz` / `Readyz` のテスト。`stubPinger` で DB の成功・失敗を再現。 |

新しいハンドラを追加するときは、リクエスト / レスポンス用の構造体をこのパッケージ内に定義し、domain のエンティティを直接 JSON にしない。

---

## `internal/router/router.go`

すべてのルート定義を置く唯一の場所。`Deps` 構造体で必要な依存（ロガー・JWT 秘密鍵・各ハンドラ）を受け取り、Gin エンジンを組み立てる。

- グローバルミドルウェア: `RequestID` → `Logger` → `Recovery` の順
- `/healthz`、`/readyz` は認証なし
- `/api/v1` グループ。認証不要のルートは `v1` に、認証必須のルートは `authed`（`middleware.Auth` 適用済み）に登録する。現時点では空。

---

## `internal/usecase/`

ビジネスロジックを置く場所。まだ実装はない（`.gitkeep` のみ）。`domain` のインターフェースにのみ依存し、`repository` や `infra` の具象型を知らない。テストはモックで書く。

---

## `internal/repository/sqlc/`（生成コード・手編集禁止）

`make sqlc` の出力。`db/queries/*.sql` を変更したら再生成する。

| ファイル | 役割 |
|---|---|
| `db.go` | `DBTX` インターフェース（`*pgxpool.Pool` と `pgx.Tx` の両方を満たす）と `Queries` 構造体、`New(db)`、`WithTx(tx)`。トランザクション境界を usecase 側で決めるための土台。 |
| `models.go` | テーブルに対応する構造体（現在は `User`）。 |
| `querier.go` | 全クエリのインターフェース `Querier`。repository のテストでモック対象にできる。 |
| `users.sql.go` | `db/queries/users.sql` から生成されたメソッド（`GetUserByID` / `GetUserByEmail` / `CreateUser`）。 |

`repository` パッケージ直下には、今後 domain のリポジトリインターフェースを実装するラッパー（例: `user_repository.go`）を置く。

---

## `internal/infra/postgres/postgres.go`

`NewPool(ctx, databaseURL)`: pgxpool の接続プールを生成する。接続数上限 10・最小 1、接続の最大寿命 1 時間、アイドル 30 分、ヘルスチェック 1 分。生成後に 5 秒タイムアウトで ping し、失敗したらプールを閉じてエラーを返す。

---

## `internal/oapi/types.gen.go`（生成コード・手編集禁止）

`make gen` の出力。`api/openapi.yaml` の `components.schemas` から Go の型（`HealthResponse`、`ErrorResponse`、`ErrorDetail`）を生成している。仕様を変更したら再生成する。

---

## `api/`

| ファイル | 役割 |
|---|---|
| `openapi.yaml` | API 仕様（OpenAPI 3.0.3）。これが正で、エンドポイントの追加・変更はまずここを更新する。現在は `/healthz`、`/readyz`、共通スキーマ（`HealthResponse`、`ErrorResponse`、`ErrorDetail` とそのエラーコード enum）、`bearerAuth` セキュリティスキームを定義。 |
| `oapi-codegen.yaml` | oapi-codegen の設定。パッケージ名 `oapi`、出力先 `internal/oapi/types.gen.go`、モデル型のみ生成（サーバースタブは生成せず、ルーティングは Gin で手書きする）。 |

---

## `db/`

| ファイル | 役割 |
|---|---|
| `migrations/000001_create_users.up.sql` | `users` テーブル作成（UUID 主キー、`email` ユニーク、`password_hash`、作成・更新日時）。JWT 認証で最低限必要になるテーブルとして用意した最初のマイグレーション。 |
| `migrations/000001_create_users.down.sql` | 上記の取り消し。 |
| `queries/users.sql` | sqlc 用クエリ。`-- name: Xxx :one` 形式のコメントでメソッド名と戻り値の種類を指定する。 |

新しいマイグレーションは `make migrate-new name=xxx` で連番ファイルを作成する。up / down は必ずセットで書く。sqlc はマイグレーションファイルからスキーマを読むので、テーブル追加後に `make sqlc` を実行する。

---

## よく使うコマンド

| コマンド | 内容 |
|---|---|
| `make up` / `make down` / `make logs` | docker compose の起動・停止・API ログ |
| `make run` / `make dev` | ローカル起動（`.env` を読む）。`dev` は air によるホットリロード |
| `make test` / `make test-integration` | ユニットテスト / 統合テスト（`integration` タグ、Postgres が必要） |
| `make lint` / `make fmt` | golangci-lint による静的解析 / フォーマット |
| `make gen` / `make sqlc` | OpenAPI から型生成 / SQL から Go コード生成 |
| `make migrate-up` / `make migrate-down` / `make migrate-new name=xxx` | マイグレーション適用 / 1 つ戻す / ファイル作成 |
| `make tools` | 開発ツールを `$(go env GOPATH)/bin` にインストール |

## 環境変数

| 変数 | 既定値 | 説明 |
|---|---|---|
| `APP_ENV` | `development` | `development` / `staging` / `production` |
| `APP_PORT` | `8080` | リッスンポート |
| `DATABASE_URL` | なし（必須） | PostgreSQL 接続文字列 |
| `JWT_SECRET` | なし | JWT 署名鍵。`development` 以外では必須 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
