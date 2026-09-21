# プロジェクト構成とファイルの役割

このドキュメントはリポジトリ内の各ファイルが何をしているかを説明する。設計方針そのもの（レイヤーの依存方向・コーディング規約）は `CLAUDE.md` を参照。

## 全体像

```
.
├── cmd/api/main.go                  API サーバーのエントリポイント（DI・起動・graceful shutdown）
├── cmd/scraperosters/main.go        ロスタースクレイプ CLI のエントリポイント
├── cmd/scrapestarters/main.go       スタータースクレイプ CLI のエントリポイント
├── cmd/seed/main.go                 マスタデータ投入 CLI のエントリポイント
├── internal/                        アプリ本体（外部から import 不可）
│   ├── config/                      環境変数の読み込みと検証
│   ├── domain/                      ドメインエラー・チーム定義・ポジション/反則の定数・ロスター型とインターフェース
│   ├── logging/                     slog ロガーの生成と ctx への格納
│   ├── middleware/                  リクエスト ID・アクセスログ・リカバリ・JWT 認証
│   ├── handler/                     Gin ハンドラ・エラー→HTTP 変換
│   ├── router/                      ルート定義の集約
│   ├── usecase/                     ビジネスロジック（ロスタースクレイプ）
│   ├── repository/sqlc/             sqlc 生成コード（手編集禁止）
│   ├── seed/                        マスタデータの定義と投入ロジック（Laravel の database/seeders 相当）
│   ├── infra/postgres/              pgx 接続プール
│   ├── infra/pfr/                   pro-football-reference.com のページ取得（chromedp）と解析（goquery）。ロスター表・スターター表の両方を扱う
│   ├── infra/csvwriter/             ロスターの CSV 出力
│   └── oapi/                        oapi-codegen 生成コード（手編集禁止）
├── api/                             OpenAPI 仕様と生成設定
├── db/migrations/                   golang-migrate 用 SQL
├── db/queries/                      sqlc 用 SQL
├── docs/                            ドキュメント
├── outputs/csv/                     スクレイプ結果の CSV（git 管理外、実行時に自動作成）
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
| `.gitignore` | `tmp/`（air のビルド出力）、`bin/`、`.env`、`coverage.out`、`outputs/`（スクレイプ出力）を除外。 |
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

## `cmd/scraperosters/main.go`

NFL 全チームのロスターを pro-football-reference.com からスクレイプし、CSV に一括出力する CLI。Laravel 側の `scrape:rosters` コマンドを Go に移植したもので、DB 保存の代わりにメモリ上に蓄積して最後にまとめて書き出す。

```
go run ./cmd/scraperosters -season 2024                 # 全 32 チーム → outputs/csv/2024_rosters.csv
go run ./cmd/scraperosters -season 2024 -teams buf,nwe  # チームを絞る
make scrape-rosters season=2024 teams=buf out=buf.csv   # → outputs/csv/buf.csv
```

CSV は常に `outputs/csv/` 配下に出力される（無ければ自動作成、git 管理外）。

| フラグ | 既定 | 説明 |
|---|---|---|
| `-season` | 2020 | 対象シーズン。2012 以降 |
| `-teams` | 全チーム | 略称のカンマ区切り |
| `-out` | `<season>_rosters.csv` | 出力ファイル名（`outputs/csv/` 配下） |
| `-delay` | 3s | チーム間の待機。連続アクセスによる遮断を避ける |
| `-timeout` | 30s | 1 ページの取得タイムアウト |
| `-headless` | false | 後述の理由で通常は false のまま |
| `-dump-dir` | なし | 取得 HTML を保存（デバッグ・テストデータ作成用） |

環境変数 `CHROME_WS_URL` を設定するとリモート Chrome（DevTools エンドポイント）に接続する。未設定ならローカルの Chrome を起動する。

**headless で動かない理由**: スクレイプ先は Cloudflare の bot 対策（Turnstile）を使っており、headless Chrome や自動化フラグ付きの Chrome ではチャレンジ画面から先に進めないことを確認済み。画面ありで `disable-blink-features=AutomationControlled` を付けて起動すると数秒で自動通過する。実行中は Chrome のウィンドウを閉じないこと。

処理の流れは main → `usecase.ScrapeRosters.Run` → チームごとに `pfr.Fetcher.FetchRoster` → 全チーム完了後に `csvwriter.Writer.Write`。途中でエラーや Ctrl+C があっても、それまでに取得した分は CSV に書き出してから終了コード 1 で終わる。

---

## `cmd/seed/main.go`

マスタデータを投入する CLI。Laravel の `php artisan db:seed` に相当する。

```
make seed              # データが無いテーブルにだけ投入
make seed force=1      # 対象テーブルを TRUNCATE してから投入し直す
go run ./cmd/seed -force
```

環境変数は `DATABASE_URL`（必須）、`SEED_ROOT_EMAIL` / `SEED_ROOT_PASSWORD`（root ユーザー。未設定なら users と settings をスキップ）、`LOG_LEVEL`。`make seed` は `.env` を読み込む。

投入対象と順序: teams → positions → formations → pf_relations → users → settings → foul_rules。全体を 1 トランザクションで実行するので、途中で失敗しても中途半端な状態にはならない。既定ではテーブルごとに「既に行があればスキップ」する冪等な動作で、`-force` のときだけ `TRUNCATE ... RESTART IDENTITY CASCADE` で消してから入れ直す（rosters / starters / tf_relations / origin_rosters も CASCADE で消えるので注意）。

## `cmd/scrapestarters/main.go`

NFL 各チームのスターター（各ポジションの主力選手）をスクレイプし、CSV に一括出力する CLI。Laravel 側の `scrape:starters` コマンドを Go に移植したもので、ロスターと同じ考え方（DB 保存の代わりにメモリ上に蓄積して最後にまとめて書き出す）を踏襲している。

```
go run ./cmd/scrapestarters -season 2024                 # 全 32 チーム → outputs/csv/2024_starters.csv
go run ./cmd/scrapestarters -season 2024 -teams buf,nwe  # チームを絞る
make scrape-starters season=2024 teams=buf out=buf.csv
```

フラグ構成は `cmd/scraperosters` と同じ（`-season` `-teams` `-out` `-delay` `-timeout` `-headless` `-dump-dir`）。CSV は `outputs/csv/` 配下、既定ファイル名は `<season>_starters.csv`。headless が使えない理由（Cloudflare の bot 対策）も scraperosters と同じ。

**スターター表はロスター表と同じ 1 ページ**（`https://www.pro-football-reference.com/teams/<team>/<season>_roster.htm`）にある。ページ内の `#starters` テーブルにオフェンス 11 人・ディフェンス 11 人が並ぶ。そのため `pfr.Fetcher` は `FetchRoster` と `FetchStarters` の 2 つのメソッドを持ち、内部のページ取得処理（`fetchRosterPageHTML`。チャレンジ突破・ポーリング・リトライを含む）を共有している。

処理の流れは main → `usecase.ScrapeStarters.Run` → チームごとに `pfr.Fetcher.FetchStarters` → `pfr.ParseStarters`（`#starters .full_table` 行から `th[data-stat=pos]` のテキストと `td[data-stat=player]` の `csk` 属性を抽出。見出し行「Offensive/Defensive Starters」は `full_table` クラスが無いので除外）→ 全チーム完了後に `csvwriter.StartersWriter.Write`。

**odflg（攻守フラグ）の決め方**: スクレイプしたポジション名（例 `QB`, `LDE`）を `domain.PositionCategory` でマスタと照合してカテゴリを取得し、`domain.IsOffenceCategory`（カテゴリ 11 未満が攻撃）で判定する。Laravel の `ScrapeStarters` が `positions.category < 11` で判定していたのと同じ境界値。ポジションがマスタに無い場合は false（Laravel が `position_model` が null のとき odflg を false とするのと同じ挙動）。

## Laravel から意図的に変える点（scrapestarters）

- DB 依存を削除。Laravel は roster_id を突き合わせて `starters` テーブルへ保存し、`rosters.position_id` も更新していたが、CSV 出力では roster との紐付けを season/team/firstname/lastname で表現し、DB の更新は行わない。
- 1 チームごとに 2 回ページを取得していた `scrape:rosters` と `scrape:starters` の Laravel 実装と異なり、Go 側では roster 用と starters 用で fetch のたびにページを取り直す設計は同じだが、内部の取得ロジック自体は共通化されている（無駄なコード重複を避けるため）。

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
| `team.go` | `Team`（ID と URL 用略称）。`AllTeams()` が全 32 チームを既存 DB の teams.id 順に返し、`FindTeam(abbr)` で略称から引く。Laravel の `config/const.php` の `UrlTeams` / `TeamIds` に相当。 |
| `roster.go` | `RawRosterRow`（表 1 行の文字列そのまま）、`RosterEntry`（単位換算済みの CSV 1 行）、`RosterFetcher` / `RosterWriter` インターフェース。usecase はこの 2 つのインターフェースにのみ依存する。 |
| `position.go` | ポジションのカテゴリ（`CategoryQB` など）と抽象カテゴリ（`AbstractCategoryQB` など）の定数。Laravel の `config/const.php` の `Category` / `AbstCate` と同じ値。 |
| `position_data.go` | ポジションマスタ本体（84 件）。`AllPositions()`、名前からカテゴリを引く `PositionCategory(name)`、攻撃側かどうかを返す `IsOffenceCategory(category)`。`internal/seed`（DB 投入）と `cmd/scrapestarters`（odflg 判定）の両方から参照される唯一の定義元。 |
| `foul_rule.go` | 反則ルールの攻守ステータス定数（`FoulStatusCommon` など）。 |
| `starter.go` | `RawStarterRow`（スターター表 1 行）、`StarterEntry`（CSV 1 行）、`StarterFetcher` / `StarterWriter` インターフェース。 |

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

| ファイル | 役割 |
|---|---|
| `scrape_rosters.go` | `ScrapeRosters.Run(ctx, season, teams)`。シーズンの下限（2012）を検証し、チームごとに `RosterFetcher` を呼んで結果をスライスに蓄積、最後に `RosterWriter` へ一括で渡す。チーム間に待機を入れ、ctx のキャンセルにも応じる。途中で失敗しても取得済み分は書き出す。体重（lb × 0.45）と身長（feet × 30.48 + inch × 2.54）の換算はここでのみ行い、換算値は Laravel の `config/const.php` と同じ。 |
| `scrape_rosters_test.go` | Fetcher / Writer をスタブに差し替えたテスト。正常系、途中失敗時の部分出力、キャンセル、換算関数。 |
| `scrape_starters.go` | `ScrapeStarters.Run(ctx, season, teams)`。構造は `ScrapeRosters` と同じ（チーム毎に蓄積 → 最後に一括書き出し、途中失敗時も部分出力）。`toStarterEntry` が `domain.PositionCategory` / `IsOffenceCategory` で odflg を決める。 |
| `scrape_starters_test.go` | Fetcher / Writer をスタブに差し替えたテスト。odflg 判定（既知・未知ポジション）を含む。 |

`domain` のインターフェースにのみ依存し、`repository` や `infra` の具象型を知らない。

---

## `internal/seed/`

| ファイル | 役割 |
|---|---|
| `seeder.go` | `Seeder.Run(ctx, Options)`。テーブルごとに「既に行があればスキップ」を判定し、sqlc の `:copyfrom`（pgx の COPY）で一括投入する。formations は投入後に id を読み直して Key → id の対応を作り、pf_relations の外部キーに使う。root ユーザーのパスワードは bcrypt でハッシュする（Laravel の `Hash::make` と互換）。 |
| `teams_data.go` | 32 チーム。順序は `domain.AllTeams` の ID と一致。 |
| `positions_data.go` | `domain.AllPositions()` を sqlc の投入用パラメータに変換するだけの薄いファイル。実データは `internal/domain/position_data.go` にある（重複を避けるため）。 |
| `formations_data.go` | フォーメーション 36 件。`Key` は Laravel の `Formationid` のキーで、テーブルには保存しない。Laravel 側で 21 件目の名前が `Nickel(LCB)` と重複していたのを `Nickel(RCB)` に修正している。 |
| `pf_relations_data.go` | フォーメーションとポジションカテゴリの関係 374 件。フォーメーションは `Key` で参照。 |
| `foul_rules_data.go` | 反則ルール 34 件。 |
| `data_test.go` | 件数・重複・参照整合性のユニットテスト（DB 不要）。 |
| `seeder_integration_test.go` | `-tags integration` で実際の Postgres に対して force 投入 → 再実行でスキップされることを検証。 |

`*_data.go` は Laravel のシーダーからスクリプトで機械変換したもの。以降は手で編集してよい。

---

## `internal/repository/sqlc/`（生成コード・手編集禁止）

`make sqlc` の出力。`db/queries/*.sql` を変更したら再生成する。

| ファイル | 役割 |
|---|---|
| `db.go` | `DBTX` インターフェース（`*pgxpool.Pool` と `pgx.Tx` の両方を満たす）と `Queries` 構造体、`New(db)`、`WithTx(tx)`。トランザクション境界を usecase 側で決めるための土台。 |
| `models.go` | 全テーブルに対応する構造体（`User`, `Player`, `Team`, `Roster` など 12 個）。 |
| `querier.go` | 全クエリのインターフェース `Querier`。repository のテストでモック対象にできる。 |
| `users.sql.go` | `db/queries/users.sql` から生成されたメソッド（`GetUserByID` / `GetUserByEmail` / `CreateUser`）。 |
| `seed.sql.go` / `copyfrom.go` | `db/queries/seed.sql` から生成。存在チェック（`TeamsExist` など）、`TruncateSeedTables`、`ListFormations`、および `:copyfrom` による一括 INSERT（`InsertTeams` など。pgx の COPY プロトコルを使う）。 |

`repository` パッケージ直下には、今後 domain のリポジトリインターフェースを実装するラッパー（例: `user_repository.go`）を置く。

---

## `internal/infra/pfr/`

| ファイル | 役割 |
|---|---|
| `fetcher.go` | `Fetcher`。chromedp でローカル Chrome を起動（またはリモートに接続）し、`FetchRoster` でチームごとに新しいタブを開いてロスターページを取得する。起動は PHP と同じく 3 回リトライ。ページ遷移後は JS の `querySelector` で `.is_setup #roster` の出現を 0.5 秒ごとにポーリングし、`outerHTML` を取り出す。chromedp の `Navigate` / `WaitReady` を使わないのは、Cloudflare のチャレンジ画面から本ページへ遷移する過程でフレーム追跡が狂いハングするため。失敗時はページタイトルをエラーに含め、`DumpDir` 指定時は HTML を保存する。 |
| `parser.go` | `ParseRoster(html)`。goquery で `tbody tr` を走査し、`data-stat` 属性で列を判別して `RawRosterRow` にする。抽出規則は PHP と同一（名前は `csk` 属性を `,` で分割、生年月日は `csk` の `-` を除去、大学は最初の `<a>`、`Rook` は `0`、ドラフトはテキストを ` / ` で分割し年は `<a>` の先頭 4 文字、空ポジションは `No Data`）。ブラウザ非依存の純粋関数なので単体テストできる。 |
| `starters_parser.go` | `ParseStarters(html)`。同じロスターページ内の `#starters .full_table` 行を走査し、`th[data-stat=pos]` のテキストと `td[data-stat=player]` の `csk` 属性（無ければ表示名）から `RawStarterRow` を作る。オフェンス/ディフェンスの見出し行（`full_table` クラス無し）は自然に除外される。 |
| `parser_test.go` / `starters_parser_test.go` | 実ページから抽出した fixture を使ったテストと、`csk` 欠落・空セル・見出し行などのフォールバックのテスト。 |
| `testdata/buf_2024_roster.html` / `testdata/buf_2024_starters.html` | 2024 年 Buffalo Bills のロスター表・スターター表（実ページから該当部分のみ抽出）。 |

---

## `internal/infra/csvwriter/writer.go`

`Writer.Write(ctx, entries)` がロスターを指定パスの CSV に書き出す（親ディレクトリが無ければ作成、既存ファイルは上書き）。`Encode(io.Writer, entries)` に分けてあるのはテストのため。列は `season, team_id, team, number, position, firstname, lastname, birthday, weight_kg, height_cm, college, experience, drafted_team, drafted_round, drafted_rank, drafted_year`。換算できなかった体重・身長は空欄。

`StartersWriter.Write(ctx, entries)` がスターターを同様に書き出す。列は `season, team_id, team, position, firstname, lastname, odflg`。ファイル I/O（ディレクトリ作成・生成・クローズ）は `file.go` の `writeFile` に共通化してあり、両方の Writer から使われている。

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

### `db/migrations/`

Laravel 側（`nfl-library-backend/database/migrations`）の業務テーブルを PostgreSQL 向けに移植したもの。Laravel 専用テーブル（password_resets / failed_jobs / personal_access_tokens）は Go + JWT では不要なため作っていない。連番は Laravel の作成順と同じ。

| ファイル | テーブル | 補足 |
|---|---|---|
| `000001_create_users` | users | 名前・メール・パスワード・お気に入りシーズン/チーム |
| `000002_create_players` | players | 選手マスタ。身長・体重は `NUMERIC(5,1)` |
| `000003_create_teams` | teams | 32 チーム。作成後に `users.favorite_team` へ外部キーを張る |
| `000004_create_formations` | formations | フォーメーション。`odflg` で攻守を区別 |
| `000005_create_positions` | positions | ポジション。Laravel で後から追加された `odflg` を最初から含む |
| `000006_create_rosters` | rosters | シーズン × チーム × 選手。`(season, team_id, player_id)` に UNIQUE |
| `000007_create_starters` | starters | スターター。rosters を参照 |
| `000008_create_tf_relations` | tf_relations | チームとフォーメーションの関係 |
| `000009_create_pf_relations` | pf_relations | フォーメーションとポジションカテゴリの関係 |
| `000010_create_settings` | settings | ユーザー設定。`(user_id, setting_id)` が複合主キー |
| `000011_create_foul_rules` | foul_rules | 反則ルール |
| `000012_create_origin_rosters` | origin_rosters | スクレイプ結果の生データ置き場。`cmd/scraperosters` の CSV と同じ列構成 |

MySQL → PostgreSQL の主な対応: `increments` → `INTEGER GENERATED ALWAYS AS IDENTITY`、`string` → `VARCHAR(255)`、`double(5,1)` → `NUMERIC(5,1)`、`tinyInteger` → `SMALLINT`、`timestamps()` → `TIMESTAMPTZ NOT NULL DEFAULT now()`。列コメントは `COMMENT ON COLUMN` で全列に移植している。

Laravel には無かった外部キー制約と検索用インデックスを追加している。そのため投入順序は users / players / teams / formations / positions → rosters → starters の順を守る必要がある。

### `db/queries/`

| ファイル | 役割 |
|---|---|
| `users.sql` | ユーザーの取得・作成。`-- name: Xxx :one` 形式のコメントでメソッド名と戻り値の種類を指定する。 |
| `seed.sql` | シードコマンド専用。存在チェック、TRUNCATE、`:copyfrom` の一括 INSERT。 |

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
| `make scrape-rosters season=2024 [teams=buf,nwe] [out=x.csv]` | ロスターをスクレイプして `outputs/csv/` に CSV 出力 |
| `make scrape-starters season=2024 [teams=buf,nwe] [out=x.csv]` | スターターをスクレイプして `outputs/csv/` に CSV 出力 |
| `make seed [force=1]` | マスタデータを投入 |

## 環境変数

| 変数 | 既定値 | 説明 |
|---|---|---|
| `APP_ENV` | `development` | `development` / `staging` / `production` |
| `APP_PORT` | `8080` | リッスンポート |
| `DATABASE_URL` | なし（必須） | PostgreSQL 接続文字列 |
| `JWT_SECRET` | なし | JWT 署名鍵。`development` 以外では必須 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `SEED_ROOT_EMAIL` / `SEED_ROOT_PASSWORD` | なし | `make seed` の root ユーザー。未設定なら users / settings をスキップ |
