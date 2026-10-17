# プロジェクト構成とファイルの役割

このドキュメントはリポジトリ内の各ファイルが何をしているかを説明する。設計方針そのもの（レイヤーの依存方向・コーディング規約）は `CLAUDE.md` を参照。

## 全体像

```
.
├── cmd/api/main.go                  API サーバーのエントリポイント（DI・起動・graceful shutdown）
├── cmd/scraperosters/main.go        ロスタースクレイプ CLI のエントリポイント
├── cmd/scrapestarters/main.go       スタータースクレイプ CLI のエントリポイント
├── cmd/seed/main.go                 マスタデータ投入 CLI のエントリポイント
├── cmd/importrosters/main.go        CSV → players/rosters 反映 CLI のエントリポイント
├── cmd/importmadden/main.go         Madden NFL 評価CSV → rosters 反映 CLI のエントリポイント
├── cmd/importdata/main.go           db/data の roster/starter CSV 一括投入 CLI のエントリポイント
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
│   ├── rosterimport/                CSV から players / rosters を生成するロジック
│   ├── maddenimport/                Madden NFL 評価CSVで rosters を更新するロジック
│   ├── starterimport/               starter CSV から starters テーブルを更新するロジック
│   ├── infra/postgres/              pgx 接続プール
│   ├── infra/pfr/                   pro-football-reference.com のページ取得（chromedp）と解析（goquery）。ロスター表・スターター表の両方を扱う
│   ├── infra/csvwriter/             ロスターの CSV 出力
│   └── oapi/                        oapi-codegen 生成コード（手編集禁止）
├── api/                             OpenAPI 仕様と生成設定
├── db/migrations/                   golang-migrate 用 SQL
├── db/queries/                      sqlc 用 SQL
├── db/data/roster_starters/         cmd/importdata が一括投入する roster/starter CSV（2012〜2025年）
├── db/data/madden/                  cmd/importdata が一括投入する Madden NFL 評価値CSV（年ごとのサブディレクトリ）
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

## `cmd/importrosters/main.go`

`cmd/scraperosters` が出力した CSV を読み込み、players / rosters テーブルへ反映する CLI。Laravel の `ScrapeRoster`（DB 書き込み部分。スクレイピング部分は既に `cmd/scraperosters` へ移植済み）に相当する。

```
go run ./cmd/importrosters -csv outputs/csv/2024_rosters.csv
make import-rosters csv=outputs/csv/2024_rosters.csv
make import-rosters csv=outputs/csv/2024_rosters.csv strict=1
```

事前に `make seed` で `teams` / `positions` が投入されている必要がある（無ければ `positions table is empty` 等のエラーになる）。

処理は `internal/rosterimport.Importer.Run` が担う。1 トランザクションで全行を処理し、途中で回復不能なエラーが起きた場合は何も反映されない。

1. `positions` を最初に全件読み込み、名前 → id のマップを作る（行ごとに問い合わせない）
2. 行ごとに `firstname + lastname + birthday` で既存 player を検索。無ければ作成（既存が見つかった場合は CSV の値で上書きしない）
3. `rosters` を `season, team_id, player_id` の一意制約を使って `INSERT ... ON CONFLICT DO UPDATE` で upsert する

**未知のポジションの扱い**: 既定ではその行を `WARN` ログに残してスキップし、続行する（`Result.Skipped` に集計）。`-strict` を付けると Laravel と同じように即時中断する。実データでは pro-football-reference.com 特有の複合表記（例 `FB/DL`）が正規のポジション名に含まれないことがあるため、`internal/domain/position_data.go` にマスタを追加して解決するのが本来の直し方で、スキップは最終手段。

`internal/rosterimport` は `internal/usecase` を経由しない。CLAUDE.md の「usecase は domain のインターフェースにのみ依存する」という制約は HTTP API のビジネスロジック向けで、`internal/seed` と同様、CLI 専用の import ツールとして sqlc を直接使う（domain 側の repository インターフェースは用意しない）。

## `cmd/importmadden/main.go`

Madden NFL の評価値CSVを読み込み、`rosters` テーブルの `position_id` / `rating` / `number` /
`madden_attributes` を更新する CLI。Laravel の `CsvImport` コマンドに相当する。CSV は
`db/data/madden/`（年ごとのサブディレクトリ構成）に2012〜2025年分が同梱されている。

```
go run ./cmd/importmadden                          # 全シーズン一括（既定 db/data/madden 配下）
go run ./cmd/importmadden -season 2024 -dir path/to/2024フォルダ
go run ./cmd/importmadden -season 2012 -csv a.csv -csv b.csv
make import-madden                                 # 全シーズン一括
make import-madden season=2024 dir=path/to/2024
```

`-season` を省略すると全シーズン一括モードになる。`-dir`（既定 `db/data/madden`）直下の、名前が
4桁の年であるサブディレクトリをそれぞれ1シーズンとして扱い、`internal/maddenimport.Importer.RunSeasons`
（`FindSeasonDirs` でサブディレクトリを列挙し、`CollectPaths`/`LoadFiles`/`Run` をシーズンごとに呼ぶ）
で順に処理する。このモードでは `-csv` は使えない（どのシーズンのファイルか判断できないため）。
`-season` を指定すると、その1シーズンだけを対象にする従来の挙動になる。

`cmd/importdata -type madden` も内部で同じ `RunSeasons` を呼んでいるため、両コマンドの一括処理ロジックは
共通（ログの内容も同じ）。`cmd/importmadden` 単体で Madden データだけを一括更新したい場合と、
`cmd/importdata` で rosters/starters と合わせて一括処理したい場合のどちらでも使える。

事前に `make seed`（teams / positions）と `cmd/importrosters`（対象シーズンの roster）が投入済みであること。
該当する roster が無い行はスキップする。

**列名の判定は固定の別名リストではなく正規化＋キーワード判定**（`internal/maddenimport/parser.go` の
`detectColumns`）。2012〜2025年の実データを調べたところ、年ごとに列名の表記がかなり異なっていた
（`Jersey Number` / `JerseyNumber` / `jerseyNum` / `JERSEY` など）ため、別名を列挙するより頑丈。
各フィールドは最初に一致した列だけを採用する（先勝ち）。2024年分には主レーティング列 `Overall` の他に
`stats/overall/value` のような入れ子形式の列があり、後者を rating 列と誤検出しないよう `/` を含む列は
jersey・rating の部分一致判定から除外している。

**チーム解決**: 行に `Team` 列があれば `teams.name` または `city+' '+name` に一致するチームを探す。
`NY Jets` や `Oakland Raiders`、`Washington Redskins` のような略記・旧名称は `teamNameAliases`
（`internal/maddenimport/importer.go`）で現在のチーム名に対応付ける。`Team` 列を持たない年
（2014・2015年分）は `TeamHintFromFilename` がファイル名の先頭トークン（`atl_madden_nfl_15.csv` → `atl`）
から `domain.FindTeam` で解決する。

**氏名の一致判定**: `firstname`/`lastname` は `cmd/importrosters` で作成済みの `players` テーブルと
完全一致で突き合わせる。ただし次の2点を正規化している。
- 姓の末尾の世代呼称（`Jr.`, `Sr.`, `II`, `III`, `IV`, `V`）を取り除く（players 側には含まれないため）
- イニシャル名（`A.J.` 等）はピリオドの有無が players 側で選手ごとに不統一なため、`findRoster` が
  そのままの表記で見つからなければピリオドを除いた表記でも1回だけ再検索する

**未知のポジションの扱い**は `cmd/importrosters` と同様、既定でスキップ（`-strict` で中断）。

## `cmd/importdata/main.go`

`db/data/`（2012〜2025年の14シーズン分がリポジトリに同梱されている）を一括で DB へ反映する CLI。
`db/data/roster_starters/`（`cmd/scraperosters`・`cmd/scrapestarters` の出力形式）と
`db/data/madden/`（Madden NFL の評価値CSV。年ごとのサブディレクトリ構成）の両方を対象にできる。

```
go run ./cmd/importdata                        # rosters・starters・madden とも全シーズン
go run ./cmd/importdata -type rosters          # rosters だけ
go run ./cmd/importdata -type starters -season 2024
go run ./cmd/importdata -type madden -season 2024
make import-data type=madden season=2024
```

| フラグ | 既定 | 説明 |
|---|---|---|
| `-dir` | `db/data/roster_starters` | rosters/starters CSV のあるディレクトリ |
| `-madden-dir` | `db/data/madden` | Madden 評価値CSV のあるディレクトリ（年ごとのサブディレクトリ構成） |
| `-type` | `all` | `rosters` / `starters` / `madden` / `all` |
| `-season` | 指定なし | このシーズンだけに絞る |
| `-strict` | false | 未知のポジションで中断するか |

`-type all`（既定）は **rosters → starters → madden の順に全シーズン投入**する。starters・madden は
season + team_id + firstname + lastname で対応する roster を探すため、rosters が先に無いと
全行スキップになる。

内部では既存の `internal/rosterimport.Importer`・`internal/starterimport.Importer`・
`internal/maddenimport.Importer` をファイルごと（シーズンごと）に呼び出すだけで、CSV の探索・
season 抽出（`findSeasonFiles`／`findSeasonDirs`）以外のロジックは持たない。1ファイル（または
1シーズン分のファイル群）が1トランザクションなので、どこかのシーズンで失敗してもそれ以前に
投入済みのシーズンは残る。`findSeasonDirs` は `db/data/madden/<season>/` のようにサブディレクトリ名が
そのままシーズンになる構成を前提にしている（`cmd/importmadden` が扱う `-dir`/`-csv` と同じ読み込み
ロジックを `internal/maddenimport.CollectPaths`/`LoadFiles` として共通化し、両コマンドから使っている）。

---

## `internal/config/`

| ファイル | 役割 |
|---|---|
| `config.go` | `Config` 構造体と `Load()`。envconfig で環境変数を読み、`validate()` で値を検証する。検証内容: `APP_ENV` は development / staging / production のいずれか、`APP_PORT` は 1〜65535、`DATABASE_URL` は空不可、`JWT_SECRET` は development 以外で必須、`LOG_LEVEL` は slog が解釈できる値。ヘルパーとして `IsDevelopment()`、`Addr()`（`:8080` 形式）、`SlogLevel()` を持つ。`AssetBaseURL`（`ASSET_BASE_URL`、既定空文字）は選手画像の配信元 URL。未設定なら `/api/v1/rosters` の `image_url` は常に `null` になる（検証対象ではない）。 |
| `config_test.go` | 上記の検証ルールのテスト。envconfig は「空文字で設定された変数」を未設定扱いにしないため、`unsetenv` ヘルパーで本当に未設定にしてテストしている。 |

---

## `internal/domain/`

| ファイル | 役割 |
|---|---|
| `errors.go` | センチネルエラー `ErrNotFound` / `ErrConflict` / `ErrUnauthorized` / `ErrForbidden` / `ErrValidation`。usecase / repository はこれらを `fmt.Errorf("...: %w", domain.ErrNotFound)` のようにラップして返す。HTTP ステータスは意識しない。 |
| `team.go` | `Team`（ID と URL 用略称）。`AllTeams()` が全 32 チームを既存 DB の teams.id 順に返し、`FindTeam(abbr)` で略称から引く。Laravel の `config/const.php` の `UrlTeams` / `TeamIds` に相当。 |
| `roster.go` | `RawRosterRow`（表 1 行の文字列そのまま）、`RosterEntry`（単位換算済みの CSV 1 行）、`RosterFetcher` / `RosterWriter` インターフェース。usecase はこの 2 つのインターフェースにのみ依存する。 |
| `position.go` | ポジションのカテゴリ（`CategoryQB` など）と抽象カテゴリ（`AbstractCategoryQB` など）の定数。Laravel の `config/const.php` の `Category` / `AbstCate` と同じ値。 |
| `position_data.go` | ポジションマスタ本体（136 件）。`AllPositions()`、名前からカテゴリを引く `PositionCategory(name)`、攻撃側かどうかを返す `IsOffenceCategory(category)`。`internal/seed`（DB 投入）・`cmd/scrapestarters`（odflg 判定）・`cmd/importrosters`/`cmd/importdata`（position_id 解決）から参照される唯一の定義元。Laravel の元データに無かった複合ポジション（`FB/DL`、`LEDG`/`REDG`/`MIKE`/`SAM`/`WILL`、および `db/data` の14シーズン分を一括投入した際に見つかった46件）を追加している。先頭のポジションのカテゴリを採用する既存の慣例に合わせているが、`LH`（2023年の Elgton Jenkins のスターター表記。PFR 側の表記ゆれで、本来はオフェンスラインの選手）だけは単独表記のため HB 系に割り当てる判断をしている。 |
| `foul_rule.go` | 反則ルールの攻守ステータス定数（`FoulStatusCommon` など）。 |
| `starter.go` | `RawStarterRow`（スターター表 1 行）、`StarterEntry`（CSV 1 行）、`StarterFetcher` / `StarterWriter` インターフェース。 |
| `position_data.go` の追記 | `LEDG`/`REDG`/`MIKE`/`SAM`/`WILL`（Madden 26 の `Position ID` に現れる新表記）を追加済み。`LEDG`→`LDE`、`REDG`→`RDE`、`MIKE`→`MLB` は一意に対応。`SAM`（strong-side LB）→`RLB`、`WILL`（weak-side LB）→`LLB` は既存の `ROLB`/`LOLB` の慣例に合わせた判断。 |
| `player.go` | `Player`（`/api/v1/rosters` の読み取り用エンティティ）。氏名・生年月日・身長体重（`*float64`、NULL 許容）・大学・ドラフト情報・画像ファイル名。 |
| `roster_detail.go` | `RosterDetail`（roster 1 件 + player + position + madden_attributes を束ねた読み取り用エンティティ）と `RosterRepository` インターフェース（`ListBySeasonAndTeam(ctx, season, teamID)`）。`PositionDetail`（`position.go` に定義。`position_data.go` のマスタ用 `Position` と名前が衝突するため別名にしている）に `ID` を含めて持つ。 |

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
| `response.go` | 統一レスポンスの一元化。成功時は `SuccessResponse{message, data}`、エラー時は `ErrorResponse{message, code}`、バリデーションエラー（422）時は `ValidationErrorResponse{message, code, errors}`。`code` はアプリ独自のエラーコード（`CodeValidation`=E0001 〜 `CodeUnavailable`=E0007。`api/openapi.yaml` の `ErrorCode` enum と一致させる。HTTP ステータスとは別軸で、Flutter 側の `ApiResult.error`/`validationError` が参照する）。`respondSuccess(c, message, data)`、`respondError(c, err)`（`classify` でドメインエラー → HTTP ステータス・コードに変換。未知のエラーは 500 で詳細を隠し、ログにだけ残す）、`respondValidationError(c, message, errs)`、`respondBindError(c, err)`（`ShouldBindQuery` 等の失敗を 422 に変換。`validator.ValidationErrors` からフィールド別にタグを集める）。`RegisterValidatorTagName()`（`main` 起動時に一度だけ呼ぶ。バリデーションエラーの `FieldError.Field()` を Go の構造体フィールド名でなく `form`/`json` タグ、つまりクライアントが実際に送るキー名にする）。 |
| `response_test.go` | `classify` の対応表テストと、`respondSuccess` / `respondError` / `respondValidationError` / `respondBindError` を Gin 経由で通すテスト。 |
| `health.go` | `HealthHandler`。`Healthz`（liveness、常に 200）と `Readyz`（readiness、`Pinger` インターフェース経由で DB に 2 秒タイムアウトで ping し、失敗なら 503）。`Pinger` を interface にしているのは、テストでスタブに差し替えるためと、handler が pgx の具象型を知らずに済むようにするため。 |
| `health_test.go` | `Healthz` / `Readyz` のテスト。`stubPinger` で DB の成功・失敗を再現。 |
| `roster.go` | `RosterHandler`（`GET /api/v1/rosters`）。`RosterLister` インターフェース（`usecase.ListRosters` 相当）に依存。クエリ `season`/`team_id`（`binding:"required,min=..."`）を bind し、結果を Flutter 側の `RosterResponse`/`RosterPlayerResponse`/`RosterPositionResponse` と同じ形に整形して返す。Laravel の `Player` モデルのアクセサ相当（`birthday_year`・`birthday_date`・`image_url`）はここで算出する（保存はしない）。`image_url` は `AssetBaseURL` が空なら常に `null`、画像ファイル名が無ければ `{AssetBaseURL}/no-image.jpg`、あれば `{AssetBaseURL}/players/{id}/{image_file}`。`madden_attributes`（JSONB、英語の CSV 見出しがキー）は `{key, label, value}` の配列に変換して返す。`label` は `maddenimport.JapaneseLabel` で日本語化し、未知キーは生のキーをそのまま表示名にする。キー順はソートして決定的にする。 |
| `roster_test.go` | `stubRosterLister` を使ったテスト。正常系（`birthday_year`/`birthday_date`/`odflg`/日本語ラベル/未知キーのフォールバック/`image_url`）、空配列（`null` でなく `[]`）、`image_url` の組み立て、バリデーションエラー（422・`code=E0001`・snake_case のフィールド名）、usecase エラー（500・`code=E0006`）。 |

新しいハンドラを追加するときは、リクエスト / レスポンス用の構造体をこのパッケージ内に定義し、domain のエンティティを直接 JSON にしない。

---

## `internal/router/router.go`

すべてのルート定義を置く唯一の場所。`Deps` 構造体で必要な依存（ロガー・JWT 秘密鍵・各ハンドラ）を受け取り、Gin エンジンを組み立てる。

- グローバルミドルウェア: `RequestID` → `Logger` → `Recovery` の順
- `/healthz`、`/readyz` は認証なし
- `/api/v1` グループ。認証不要のルートは `v1` に、認証必須のルートは `authed`（`middleware.Auth` 適用済み）に登録する。`v1.GET("/rosters", d.Roster.Index)` を登録済み（認証不要側）。`authed` は現時点では空。

---

## `internal/usecase/`

| ファイル | 役割 |
|---|---|
| `scrape_rosters.go` | `ScrapeRosters.Run(ctx, season, teams)`。シーズンの下限（2012）を検証し、チームごとに `RosterFetcher` を呼んで結果をスライスに蓄積、最後に `RosterWriter` へ一括で渡す。チーム間に待機を入れ、ctx のキャンセルにも応じる。途中で失敗しても取得済み分は書き出す。体重（lb × 0.45）と身長（feet × 30.48 + inch × 2.54）の換算はここでのみ行い、換算値は Laravel の `config/const.php` と同じ。 |
| `scrape_rosters_test.go` | Fetcher / Writer をスタブに差し替えたテスト。正常系、途中失敗時の部分出力、キャンセル、換算関数。 |
| `scrape_starters.go` | `ScrapeStarters.Run(ctx, season, teams)`。構造は `ScrapeRosters` と同じ（チーム毎に蓄積 → 最後に一括書き出し、途中失敗時も部分出力）。`toStarterEntry` が `domain.PositionCategory` / `IsOffenceCategory` で odflg を決める。 |
| `scrape_starters_test.go` | Fetcher / Writer をスタブに差し替えたテスト。odflg 判定（既知・未知ポジション）を含む。 |
| `list_rosters.go` | `ListRosters.Run(ctx, season, teamID)`。`season`（`MinSeason`=2012 未満は不可）と `teamID`（1〜`MaxTeamID`=32）を検証し、`domain.RosterRepository` へ委譲する。範囲外は `domain.ErrValidation` でラップ。 |
| `list_rosters_test.go` | `domain.RosterRepository` のスタブを使ったテスト。委譲・エラー伝播・season/team_id の範囲外チェック。 |

`domain` のインターフェースにのみ依存し、`repository` や `infra` の具象型を知らない。

---

## `internal/seed/`

| ファイル | 役割 |
|---|---|
| `seeder.go` | `Seeder.Run(ctx, Options)`。テーブルごとに「既に行があればスキップ」を判定し、sqlc の `:copyfrom`（pgx の COPY）で一括投入する。formations は投入後に id を読み直して Key → id の対応を作り、pf_relations の外部キーに使う。root ユーザーのパスワードは bcrypt でハッシュする（Laravel の `Hash::make` と互換）。 |
| `teams_data.go` | 32 チーム。順序は `domain.AllTeams` の ID と一致。 |
| `positions_data.go` | `domain.AllPositions()` を sqlc の投入用パラメータに変換するだけの薄いファイル。実データは `internal/domain/position_data.go` にある（重複を避けるため）。 |
| `data_test.go` | 件数（136 件）などの検証。|
| `formations_data.go` | フォーメーション 36 件。`Key` は Laravel の `Formationid` のキーで、テーブルには保存しない。Laravel 側で 21 件目の名前が `Nickel(LCB)` と重複していたのを `Nickel(RCB)` に修正している。 |
| `pf_relations_data.go` | フォーメーションとポジションカテゴリの関係 374 件。フォーメーションは `Key` で参照。 |
| `foul_rules_data.go` | 反則ルール 34 件。 |
| `data_test.go` | 件数・重複・参照整合性のユニットテスト（DB 不要）。 |
| `seeder_integration_test.go` | `-tags integration` で実際の Postgres に対して force 投入 → 再実行でスキップされることを検証。 |

`*_data.go` は Laravel のシーダーからスクリプトで機械変換したもの。以降は手で編集してよい。

---

## `internal/starterimport/`

| ファイル | 役割 |
|---|---|
| `importer.go` | `Importer.Run(ctx, entries, Options)`。season + team_id + firstname + lastname で対応する roster を検索し、見つかれば `starters` テーブルへ `(season, odflg, roster_id)` を upsert する。見つからない、またはポジションがマスタに無い行はスキップ（`-strict` で中断）。Laravel の `ScrapeStarters` と異なり `rosters.position_id` は更新しない（`cmd/importrosters`/`cmd/importmadden` の2経路が既に更新しており、3つ目を増やさないための判断）。 |
| `importer_integration_test.go` | `-tags integration` で実際の Postgres に対して新規作成・再実行時の冪等性（`starters_roster_id_key` の UNIQUE 制約による upsert）・未知ポジションの扱いを検証。 |

---

## `internal/maddenimport/`

| ファイル | 役割 |
|---|---|
| `parser.go` | `ParseCSV(io.Reader) (rows []Row, hasTeamColumn bool, err error)`。ヘッダーを正規化してキーワード判定し、`Team`/`Position`(`Position ID`優先)/氏名(結合列優先)/`Jersey`系/`Overall`系を抽出、残りは `Row.Extra` に集める。`TeamHintFromFilename` はファイル名からチーム略称を推定する。ブラウザや DB に依存しない純粋関数なので単体テストできる。 |
| `loader.go` | `CollectPaths(dir, explicit)` / `LoadFiles(paths, logger)`。ディレクトリ直下の `*.csv` 列挙とファイル読み込み（Team 列が無ければ `TeamHintFromFilename` で補う）を1箇所にまとめたもの。`FindSeasonDirs(rootDir, seasonFilter)` は年名のサブディレクトリ（`db/data/madden/<season>/`）を列挙する。`Importer.RunSeasons(ctx, rootDir, seasonFilter, opts, logger)` はこの3つを組み合わせて複数シーズンを一括処理する高レベル API。`cmd/importmadden`（`-season` 省略時）と `cmd/importdata`（`-type madden`）の両方から使う。 |
| `parser_test.go` | 2012〜2025年の実データで確認した各形式（結合名/分割名混在、Team列なし、Position ID優先、入れ子stats列など）のテスト。 |
| `importer.go` | `Importer.Run(ctx, season, []File, Options)`。1トランザクションで全行を処理。`resolveTeamID` がチーム名解決とエイリアス変換を、`findRoster` が氏名のピリオド表記ゆれを吸収したroster検索を行う。 |
| `importer_integration_test.go` | `-tags integration` で実際の Postgres に対して更新・チームエイリアス・ファイル名フォールバックを検証。 |
| `japanese.go` | `JapaneseLabel(rawHeader string) (string, bool)`。`madden_attributes`（JSONB）のキーであるCSVヘッダーに日本語名を付けるための辞書と変換関数。14シーズン分の実データには同じ能力値でも8種類ほどの表記ゆれ（大文字/小文字、`Rating`接尾辞、2024年のみの`stats/<name>/value`形式、typo等）があり、`canonicalStatKey` で正規化したうえで `japaneseNames`（概念ごとの日本語名）を引く。正規化だけでは一致しない表記（"Block"と"Blocking"の違い、Madden側の仕様変更によるFootwork→Finesse・Strength→Powerの改名、typoなど）は `aliasKeys` で読み替える。データの保存形式自体は変えず、表示時にその場で変換する用途を想定。 |
| `japanese_test.go` / `japanese_testdata_test.go` | `db/data/madden` の全14シーズンで実際に観測された生ヘッダー370件（`allObservedMaddenHeaders`）を1件も取り逃さず翻訳できることを保証するテスト。 |

---

## `internal/rosterimport/`

| ファイル | 役割 |
|---|---|
| `importer.go` | `Importer.Run(ctx, entries, Options)`。CSV から読み込んだ `[]domain.RosterEntry` を 1 トランザクションで players / rosters に反映する。`positions` は最初に全件読み込んでキャッシュする。player は firstname+lastname+birthday で検索し、見つかれば身長・体重・大学・ドラフト情報を CSV の値で更新（`UpdatePlayer`。firstname/lastname/birthday は突き合わせキーなので更新しない）、無ければ新規作成する。roster は `season, team_id, player_id` の UNIQUE 制約を使って upsert する。未知のポジションは既定でスキップ（`Options.Strict` で即時中断に変更可）。 |
| `importer_integration_test.go` | `-tags integration` で実際の Postgres に対して新規作成・再実行時の冪等性（player 再利用・roster 更新）・未知ポジションの扱いを検証。 |

---

## `internal/repository/`

| ファイル | 役割 |
|---|---|
| `pgtype.go` | `numericToFloat(pgtype.Numeric) (*float64, error)`。身長・体重列（`NUMERIC`）を `*float64` に変換する（`NULL` は `nil`）。 |
| `roster_repository.go` | `RosterRepository`（`domain.RosterRepository` の実装）。`ListBySeasonAndTeam` は `sqlc` の `ListRostersBySeasonTeam` を呼び、各行を `domain.RosterDetail` に変換する（`madden_attributes` の JSON デコード、身長体重の `pgtype.Numeric`→`*float64` 変換を含む）。 |
| `roster_repository_integration_test.go` | `-tags integration` で実際の Postgres に対して検証。実データ（2024シーズン・Josh Allen）での取得内容、未知の season/team は空配列（エラーにしない）。 |

## `internal/repository/sqlc/`（生成コード・手編集禁止）

`make sqlc` の出力。`db/queries/*.sql` を変更したら再生成する。

| ファイル | 役割 |
|---|---|
| `db.go` | `DBTX` インターフェース（`*pgxpool.Pool` と `pgx.Tx` の両方を満たす）と `Queries` 構造体、`New(db)`、`WithTx(tx)`。トランザクション境界を usecase 側で決めるための土台。 |
| `models.go` | 全テーブルに対応する構造体（`User`, `Player`, `Team`, `Roster` など 12 個）。 |
| `querier.go` | 全クエリのインターフェース `Querier`。repository のテストでモック対象にできる。 |
| `users.sql.go` | `db/queries/users.sql` から生成されたメソッド（`GetUserByID` / `GetUserByEmail` / `CreateUser`）。 |
| `seed.sql.go` / `copyfrom.go` | `db/queries/seed.sql` から生成。存在チェック（`TeamsExist` など）、`TruncateSeedTables`、`ListFormations`、および `:copyfrom` による一括 INSERT（`InsertTeams` など。pgx の COPY プロトコルを使う）。 |
| `rosterimport.sql.go` | `db/queries/rosterimport.sql` から生成。`GetPlayerByIdentity`、`CreatePlayer`、`ListPositions`、`UpsertRoster`（`ON CONFLICT ... DO UPDATE`）。 |
| `rosters_api.sql.go` | `db/queries/rosters_api.sql` から生成。`ListRostersBySeasonTeam`（rosters/players/positions の 3 テーブル JOIN。列名衝突を避けるため `player_table_id`/`position_table_id`/`position_name` 等の別名を付けている）。 |

`repository` パッケージ直下には、domain のリポジトリインターフェースを実装するラッパーを置く（`roster_repository.go` など）。

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

`decoder.go` / `starters_decoder.go` は逆方向。`ReadRosterEntriesFile` / `ReadStarterEntriesFile` がそれぞれ `Encode` / `EncodeStarters` の書式（ヘッダ行 + 列順固定）を読み戻し `[]domain.RosterEntry` / `[]domain.StarterEntry` にする。`cmd/importrosters` と `cmd/importdata` が使う。

---

## `internal/infra/postgres/postgres.go`

`NewPool(ctx, databaseURL)`: pgxpool の接続プールを生成する。接続数上限 10・最小 1、接続の最大寿命 1 時間、アイドル 30 分、ヘルスチェック 1 分。生成後に 5 秒タイムアウトで ping し、失敗したらプールを閉じてエラーを返す。

---

## `internal/oapi/types.gen.go`（生成コード・手編集禁止）

`make gen` の出力。`api/openapi.yaml` の `components.schemas` から Go の型を生成している。仕様を変更したら再生成する。

---

## `api/`

| ファイル | 役割 |
|---|---|
| `openapi.yaml` | API 仕様（OpenAPI 3.0.3）。これが正で、エンドポイントの追加・変更はまずここを更新する。`/healthz`、`/readyz`、`/api/v1/rosters`（`season`/`team_id` クエリ）を定義。共通スキーマ: `SuccessEnvelope`（`message`,`data`）、`ErrorEnvelope`（`message`,`code`）、`ValidationErrorEnvelope`（`message`,`code`,`errors`）、`ErrorCode`（`E0001`〜`E0007` の enum）、`RosterListData`/`Roster`/`RosterPlayer`/`RosterPosition`/`MaddenAttribute`。`bearerAuth` セキュリティスキームも定義。 |
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
| `000006_create_rosters` | rosters | シーズン × チーム × 選手。`(season, team_id, player_id)` に UNIQUE。`madden_attributes`（JSONB）に Madden NFL 評価値CSVのうち他列に対応しない項目一式を保持 |
| `000007_create_starters` | starters | スターター。rosters を参照。`roster_id` に UNIQUE（`cmd/importdata` の再実行を upsert で冪等にするため） |
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
| `rosterimport.sql` | `cmd/importrosters` 専用。player の検索/作成、positions 全件取得、roster の upsert。 |
| `starterimport.sql` | `cmd/importdata`（starters 側）専用。`UpsertStarter`（roster の検索は maddenimport.sql の `GetRosterByPlayerIdentity` を共用）。 |
| `maddenimport.sql` | `cmd/importmadden` 専用。チーム名検索（`name` または `city+' '+name`）、roster の氏名突合検索、`madden_attributes` を含む roster 更新。 |
| `rosters_api.sql` | `GET /api/v1/rosters` 専用。`ListRostersBySeasonTeam`（rosters/players/positions の JOIN。`season`・`team_id` で絞り込み、`players.lastname, players.firstname` でソート）。 |

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
| `make import-rosters csv=x.csv [strict=1]` | CSV から players / rosters を生成 |
| `make import-madden season=2024 dir=path [strict=1]` | Madden NFL 評価CSVで rosters を更新 |
| `make import-data [type=rosters\|starters\|madden\|all] [season=2024] [strict=1]` | `db/data` の CSV を一括投入 |

## 環境変数

| 変数 | 既定値 | 説明 |
|---|---|---|
| `APP_ENV` | `development` | `development` / `staging` / `production` |
| `APP_PORT` | `8080` | リッスンポート |
| `DATABASE_URL` | なし（必須） | PostgreSQL 接続文字列 |
| `JWT_SECRET` | なし | JWT 署名鍵。`development` 以外では必須 |
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `SEED_ROOT_EMAIL` / `SEED_ROOT_PASSWORD` | なし | `make seed` の root ユーザー。未設定なら users / settings をスキップ |
| `ASSET_BASE_URL` | 空文字 | 選手画像の配信元 URL。未設定なら `/api/v1/rosters` の `image_url` は常に `null` |
