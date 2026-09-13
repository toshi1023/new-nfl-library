// Command api は HTTP API サーバーのエントリポイント。DI（手組み）・起動・graceful shutdown のみを行う。
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/toshi1023/new-nfl-library/internal/config"
	"github.com/toshi1023/new-nfl-library/internal/handler"
	"github.com/toshi1023/new-nfl-library/internal/infra/postgres"
	"github.com/toshi1023/new-nfl-library/internal/logging"
	"github.com/toshi1023/new-nfl-library/internal/router"
)

// shutdownTimeout は停止シグナル受信後、処理中のリクエストの完了を待つ上限時間。
// これを超えると未完了のリクエストがあっても強制的に終了する。
const shutdownTimeout = 10 * time.Second

// main は run を呼び出すだけの薄いラッパー。
// run がエラーを返した場合はログに残して終了コード 1 でプロセスを終える。
// os.Exit は defer を実行しないため、リソース解放を伴う処理は run 側に置く。
func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("err", err))
		os.Exit(1)
	}
}

// run はサーバーの起動から停止までの一連の流れを担う。
// 1. 設定読み込み → 2. ロガー初期化 → 3. インフラ（DB）接続 → 4. 依存の組み立て
// → 5. HTTP サーバー起動 → 6. シグナル待機 → 7. graceful shutdown
func run() error {
	// ---- 1. 設定読み込み ----
	// 環境変数から Config を生成し、必須項目や値の範囲を検証する。
	// ここで失敗した場合は起動を続行しない（設定不備で中途半端に動くことを防ぐ）。
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// ---- 2. ロガー初期化 ----
	// LOG_LEVEL を slog.Level に変換し、JSON 形式の構造化ロガーを生成する。
	// slog.SetDefault で全体の既定ロガーにしておくことで、ctx にロガーが無い箇所
	// （logging.FromContext のフォールバック）でも同じ出力形式になる。
	level, err := cfg.SlogLevel()
	if err != nil {
		return err
	}
	logger := logging.New(os.Stdout, level)
	slog.SetDefault(logger)

	// 開発環境以外では Gin のデバッグ出力（ルート一覧など）を抑止する。
	if !cfg.IsDevelopment() {
		gin.SetMode(gin.ReleaseMode)
	}

	// SIGINT（Ctrl+C）/ SIGTERM（docker stop 等）を受け取るとキャンセルされる ctx を作る。
	// この ctx は DB 接続などの初期化にも渡し、起動途中でもシグナルで中断できるようにする。
	// stop() を defer しておくことで、run 終了時にシグナルハンドラを解除する。
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---- 3. インフラ（DB）接続 ----
	// pgx の接続プールを生成し、ping で疎通を確認する。
	// DB に到達できない状態で起動しても意味がないため、失敗時はエラーで終了する。
	// defer pool.Close() により、run を抜けるときに接続を確実に閉じる。
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	// ---- 4. 依存の組み立て（手組み DI） ----
	// 下位レイヤーから順に生成し、コンストラクタ経由で上位へ渡す。
	// 今後は repository → usecase → handler の順にここへ追加していく。
	// pool は handler.Pinger インターフェースを満たすため、readiness チェック用にそのまま渡す。
	healthHandler := handler.NewHealthHandler(pool)

	// ---- ルーター組み立て ----
	// ミドルウェアと全ルートの登録は router パッケージに集約されている。
	// main はハンドラを Deps に詰めて渡すだけで、ルーティングの詳細を知らない。
	engine := router.New(router.Deps{
		Logger:    logger,
		JWTSecret: cfg.JWTSecret,
		Health:    healthHandler,
	})

	// ---- 5. HTTP サーバー生成 ----
	// 各種タイムアウトを明示して、遅いクライアントによる接続の占有（Slowloris 等）を防ぐ。
	//   ReadHeaderTimeout: リクエストヘッダの読み取り上限
	//   ReadTimeout:       リクエスト全体（ボディ含む）の読み取り上限
	//   WriteTimeout:      レスポンス書き込みの上限
	//   IdleTimeout:       Keep-Alive 接続を維持する上限
	srv := &http.Server{
		Addr:              cfg.Addr(),
		Handler:           engine,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// サーバーは別 goroutine で起動し、異常終了したときだけ errCh にエラーを送る。
	// srv.Shutdown による正常停止では ListenAndServe が http.ErrServerClosed を返すため、
	// それはエラー扱いにしない。
	// バッファ 1 のチャネルにしておくことで、受信側がいなくても goroutine がブロックしない。
	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting server", slog.String("addr", cfg.Addr()), slog.String("env", cfg.AppEnv))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	// ---- 6. 停止要因の待機 ----
	// 「シグナルを受信した」か「サーバーが異常終了した」のどちらか早い方で先に進む。
	// 後者でエラーが nil の場合（正常に close されたケース）はそのまま shutdown 処理へ進む。
	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil {
			return err
		}
	}

	// ---- 7. graceful shutdown ----
	// 新規接続の受付を止め、処理中のリクエストが完了するのを shutdownTimeout まで待つ。
	// ここでは signal 由来の ctx は既にキャンセル済みのため、新たに context.Background() から
	// タイムアウト付き ctx を作る必要がある。
	// この後 defer により pool.Close() と stop() が実行され、リソースが解放される。
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	logger.Info("server stopped")
	return nil
}
