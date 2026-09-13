# CLAUDE.md

このリポジトリは Go（Gin）製の API サーバーです。コードを書く・変更する際は本ドキュメントの方針に従ってください。

## 技術スタック

| 役割 | 採用 | 備考 |
|---|---|---|
| 言語 | Go 1.23+ | `go.mod` の `go` ディレクティブに従う |
| HTTP フレームワーク | Gin (`github.com/gin-gonic/gin`) | ルーティング・バインディング・ミドルウェア |
| DB | PostgreSQL 16 | |
| DB アクセス | sqlc (`github.com/sqlc-dev/sqlc`) + pgx (`github.com/jackc/pgx/v5`) | SQL を `db/queries/` に書き、コード生成する。ORM は使わない |
| マイグレーション | golang-migrate (`github.com/golang-migrate/migrate/v4`) | `db/migrations/` |
| バリデーション | go-playground/validator | Gin の `binding` タグで利用 |
| 設定 | `github.com/kelseyhightower/envconfig` | 環境変数 → 構造体。`.env` は開発用のみ |
| ログ | `log/slog`（標準ライブラリ） | JSON 出力の構造化ログ。`fmt.Println` / `log.Println` は禁止 |
| 認証 | `github.com/golang-jwt/jwt/v5` | Bearer JWT |
| API 仕様 | OpenAPI 3 (`api/openapi.yaml`) + oapi-codegen | スキーマファースト。仕様を先に更新してから実装する |
| テスト | `testing` + `github.com/stretchr/testify` + `net/http/httptest` | |
| Lint | golangci-lint | CI で必須 |
| ホットリロード | air (`github.com/air-verse/air`) | 開発環境のみ |

## ディレクトリ構成

```
.
├── cmd/
│   └── api/main.go            # エントリポイント。DI（手組み）・起動・graceful shutdown のみ
├── internal/                  # アプリ本体。外部パッケージから import 不可
│   ├── config/                # 環境変数の読み込みと検証
│   ├── handler/               # Gin ハンドラ（Controller 相当）。HTTP <-> usecase の変換だけを行う
│   │   └── response.go        # 共通レスポンス整形・エラー→HTTP ステータス変換
│   ├── middleware/            # 認証・リクエストログ・リカバリ・CORS・リクエスト ID
│   ├── router/router.go       # 全ルート定義を 1 箇所に集約
│   ├── usecase/               # ビジネスロジック（Service 相当）
│   ├── domain/                # エンティティ・値オブジェクト・リポジトリインターフェース・ドメインエラー
│   ├── repository/            # domain のリポジトリインターフェースの実装。sqlc 生成コードをラップする
│   │   └── sqlc/              # sqlc 生成コード（手編集禁止）
│   └── infra/                 # DB 接続プール・外部 API クライアント・キャッシュ
├── db/
│   ├── migrations/            # golang-migrate 形式（NNNNNN_name.up.sql / .down.sql）
│   └── queries/               # sqlc 用 SQL
├── api/
│   └── openapi.yaml
├── pkg/                       # 外部公開可能な汎用コード。必要になるまで作らない
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── sqlc.yaml
├── .air.toml
└── go.mod
```

## レイヤーと依存の向き

```
handler → usecase → domain ← repository
                       ↑
                     infra
```

- **handler**: リクエストのバインド・バリデーション、usecase 呼び出し、レスポンス整形。ビジネスロジックを書かない。
- **usecase**: ビジネスロジック。`domain` のインターフェースにのみ依存し、`repository` や `infra` の具象型を知らない。
- **domain**: エンティティとリポジトリインターフェース、ドメインエラー。他の internal パッケージを import しない。
- **repository / infra**: domain のインターフェースを実装する。上位レイヤーを import しない。
- 依存の注入は `cmd/api/main.go` で手組みで行う。wire / fx などの DI ライブラリは導入しない。

## コーディング規約

### 全般
- すべてのレイヤーの関数は第一引数に `ctx context.Context` を取る。`context.Background()` を main 以外で生成しない。
- エラーは `fmt.Errorf("...: %w", err)` でラップして返す。握りつぶさない。`panic` はプログラムのバグ以外で使わない。
- 構造体は必ずコンストラクタ（`NewXxx`）経由で生成し、依存を明示的に渡す。グローバル変数・シングルトンは禁止。
- 公開シンボルには doc コメントを付ける。

### エラーハンドリング
- `domain` にセンチネルエラーを定義する（例: `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrValidation`）。
- usecase / repository はセンチネルエラーを `%w` でラップして返す。HTTP ステータスを意識しない。
- HTTP ステータスへの変換は `internal/handler/response.go` の 1 箇所でのみ行う（`errors.Is` で判定）。
- エラーレスポンスの JSON 形式は統一する:
  ```json
  { "error": { "code": "NOT_FOUND", "message": "user not found" } }
  ```

### ハンドラ
- Gin の `c.ShouldBindJSON` / `c.ShouldBindQuery` でバインドし、`binding` タグでバリデーションする。
- リクエスト / レスポンス用の構造体は handler 内に定義し、domain のエンティティを直接 JSON にしない。
- ルートは `internal/router/router.go` にのみ定義する。`/api/v1` でグループ化する。

### DB
- SQL は `db/queries/*.sql` に書き、`make sqlc` で生成する。生成コードは手編集しない。
- スキーマ変更は必ずマイグレーションファイル（up / down 両方）で行う。
- トランザクションは usecase 側で境界を決め、repository にはトランザクション付き `*pgx.Tx` または `pgx.Tx` を受け取れる手段を用意する。

### ログ
- `slog` を使い、`slog.String("key", value)` 形式で構造化する。
- リクエストごとに request ID を発行し、ミドルウェアで `ctx` に格納した logger を全ログに使う。
- パスワード・トークン・個人情報をログに出さない。

### テスト
- usecase のテストは domain インターフェースのモックで書く（DB 不要）。
- handler のテストは `httptest` + Gin のテストモードで書く。
- repository のテストは docker-compose の Postgres に接続する統合テストとし、`//go:build integration` タグを付ける。
- テストファイルは対象と同じパッケージに `xxx_test.go` として置く。

### API 仕様
- `api/openapi.yaml` が正。エンドポイントの追加・変更時はまず仕様を更新し、`make gen` で型を再生成してから実装する。

## 主要コマンド（Makefile）

| コマンド | 内容 |
|---|---|
| `make up` | docker compose で API + Postgres を起動（ホットリロード） |
| `make down` | 停止 |
| `make run` | ローカルで API を起動 |
| `make test` | ユニットテスト |
| `make test-integration` | 統合テスト（`-tags integration`） |
| `make lint` | golangci-lint |
| `make sqlc` | sqlc 生成 |
| `make gen` | oapi-codegen 生成 |
| `make migrate-up` / `make migrate-down` | マイグレーション |
| `make migrate-new name=xxx` | マイグレーションファイル作成 |

## 環境変数

| 変数 | 例 | 説明 |
|---|---|---|
| `APP_ENV` | `development` | `development` / `staging` / `production` |
| `APP_PORT` | `8080` | リッスンポート |
| `DATABASE_URL` | `postgres://app:app@db:5432/app?sslmode=disable` | 接続文字列 |
| `JWT_SECRET` | | JWT 署名鍵。開発環境以外で必須 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |

## 新機能の追加手順

1. `api/openapi.yaml` にエンドポイントを追加し `make gen`
2. 必要なら `db/migrations/` にマイグレーション、`db/queries/` に SQL を追加し `make sqlc`
3. `domain` にエンティティ・リポジトリインターフェース・エラーを定義
4. `repository` に実装
5. `usecase` にロジックを実装し、モックでテスト
6. `handler` を実装し、`router` にルートを登録
7. `make lint && make test` を通す

## やらないこと

- ORM（GORM 等）の導入
- グローバル変数・`init()` での副作用
- handler へのビジネスロジックの記述
- `internal/` 内のレイヤー逆依存（usecase から handler、domain から repository 等）
- 生成コード（`internal/repository/sqlc/`、oapi-codegen 出力）の手編集
