package pfr

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"

	"github.com/toshi1023/new-nfl-library/internal/domain"
	"github.com/toshi1023/new-nfl-library/internal/logging"
)

// baseURL はロスターページの URL テンプレート。
const baseURL = "https://www.pro-football-reference.com/teams/%s/%d_roster.htm"

// Options は Fetcher の設定。
type Options struct {
	// RemoteURL はリモート Chrome の DevTools エンドポイント（例: ws://localhost:9222）。空ならローカルの Chrome を起動する。
	RemoteURL string
	// Headless はローカル起動時に headless で起動するか。
	// スクレイプ先は Cloudflare の bot 対策（Turnstile）を使っており、headless では通過できないことを確認済み。
	// 画面ありで起動すると自動的に通過する。通常は false にすること。
	Headless bool
	// UserAgent は空なら Chrome 既定の UA を使う。
	// 実際の Chrome バージョンと異なる UA を名乗るとクライアントヒントとの不一致で bot 判定されやすくなるため、通常は上書きしない。
	UserAgent string
	// PageTimeout は 1 ページの取得（遷移 + 表の表示待ち）に許す時間。
	PageTimeout time.Duration
	// Retries はブラウザ起動・接続のリトライ回数。
	Retries int
	// RetryInterval はリトライ間隔。
	RetryInterval time.Duration
	// DumpDir を指定すると取得した HTML を <DumpDir>/<abbr>_<season>_roster.html に保存する（デバッグ・テストデータ用）。
	DumpDir string
}

func (o *Options) applyDefaults() {
	if o.PageTimeout <= 0 {
		o.PageTimeout = 30 * time.Second
	}
	if o.Retries <= 0 {
		o.Retries = 3
	}
	if o.RetryInterval <= 0 {
		o.RetryInterval = time.Second
	}
}

// Fetcher は chromedp でロスターページを開き、HTML を取得して解析する。domain.RosterFetcher を実装する。
type Fetcher struct {
	browserCtx context.Context
	cancel     context.CancelFunc
	opts       Options
	logger     *slog.Logger
}

// NewFetcher はブラウザを起動（またはリモートに接続）して Fetcher を返す。使い終わったら Close を呼ぶこと。
func NewFetcher(ctx context.Context, opts Options) (*Fetcher, error) {
	opts.applyDefaults()
	logger := logging.FromContext(ctx)

	var lastErr error
	for attempt := 1; attempt <= opts.Retries; attempt++ {
		browserCtx, cancel, err := startBrowser(ctx, opts)
		if err == nil {
			return &Fetcher{browserCtx: browserCtx, cancel: cancel, opts: opts, logger: logger}, nil
		}
		lastErr = err
		logger.Warn("browser start failed", slog.Int("attempt", attempt), slog.Any("err", err))

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("start browser: %w", ctx.Err())
		case <-time.After(opts.RetryInterval):
		}
	}
	return nil, fmt.Errorf("start browser after %d attempts: %w", opts.Retries, lastErr)
}

// startBrowser はアロケータとブラウザコンテキストを作り、実際にブラウザを立ち上げる。
// 失敗したコンテキストは再利用できないため、リトライのたびに作り直す。
func startBrowser(ctx context.Context, opts Options) (context.Context, context.CancelFunc, error) {
	var (
		allocCtx    context.Context
		allocCancel context.CancelFunc
	)
	if opts.RemoteURL != "" {
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(ctx, opts.RemoteURL)
	} else {
		execOpts := []chromedp.ExecAllocatorOption{
			chromedp.NoFirstRun,
			chromedp.NoDefaultBrowserCheck,
			chromedp.NoSandbox,
			chromedp.DisableGPU,
			chromedp.WindowSize(1920, 1600),
			chromedp.Flag("disable-infobars", true),
			// navigator.webdriver 等の自動化シグナルを出さない。これが無いと Cloudflare の bot 判定に引っかかる
			chromedp.Flag("disable-blink-features", "AutomationControlled"),
		}
		if opts.UserAgent != "" {
			execOpts = append(execOpts, chromedp.UserAgent(opts.UserAgent))
		}
		if opts.Headless {
			execOpts = append(execOpts, chromedp.Headless)
		}
		allocCtx, allocCancel = chromedp.NewExecAllocator(ctx, execOpts...)
	}

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	cancel := func() {
		browserCancel()
		allocCancel()
	}

	// 空の Run でブラウザプロセスの起動（または接続）だけを行う
	if err := chromedp.Run(browserCtx); err != nil {
		cancel()
		return nil, nil, err
	}
	return browserCtx, cancel, nil
}

// Close はブラウザを終了する。
func (f *Fetcher) Close() {
	f.cancel()
}

// FetchRoster は指定シーズン・チームのロスターページを開き、ロスター表（#roster）の行を返す。
func (f *Fetcher) FetchRoster(ctx context.Context, season int, team domain.Team) ([]domain.RawRosterRow, error) {
	html, url, err := f.fetchRosterPageHTML(ctx, season, team)
	if err != nil {
		return nil, err
	}
	rows, err := ParseRoster(html)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", url, err)
	}
	return rows, nil
}

// FetchStarters は指定シーズン・チームのロスターページを開き、スターター表（#starters）の行を返す。
// ロスター表と同じ 1 ページに両方の表があるため、取得処理は FetchRoster と共通（fetchRosterPageHTML）。
func (f *Fetcher) FetchStarters(ctx context.Context, season int, team domain.Team) ([]domain.RawStarterRow, error) {
	html, url, err := f.fetchRosterPageHTML(ctx, season, team)
	if err != nil {
		return nil, err
	}
	rows, err := ParseStarters(html)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", url, err)
	}
	return rows, nil
}

// fetchRosterPageHTML はロスターページ（ロスター表とスターター表の両方を含む）を開き、HTML 全体を返す。
func (f *Fetcher) fetchRosterPageHTML(ctx context.Context, season int, team domain.Team) (html, url string, err error) {
	url = fmt.Sprintf(baseURL, team.Abbr, season)

	// チームごとに新しいタブを開き、終わったら閉じる
	tabCtx, cancelTab := chromedp.NewContext(f.browserCtx)
	defer cancelTab()
	// 呼び出し元の ctx（シグナル等）がキャンセルされたらタブも止める
	stop := context.AfterFunc(ctx, cancelTab)
	defer stop()

	f.logger.Debug("fetching roster page", slog.String("url", url))

	// ページ遷移と表の表示待ちにはタイムアウトを付ける。タブ自体（tabCtx）は生かしておき、
	// 失敗時に診断用の HTML を取り出せるようにする。
	waitCtx, cancelWait := context.WithTimeout(tabCtx, f.opts.PageTimeout)
	defer cancelWait()

	html, err = f.loadRosterHTML(waitCtx, url)
	if err != nil {
		title := f.diagnose(tabCtx, team.Abbr, season)
		if errors.Is(err, context.DeadlineExceeded) {
			return "", url, fmt.Errorf("fetch %s: roster table did not appear within %s (page title %q): %w", url, f.opts.PageTimeout, title, err)
		}
		return "", url, fmt.Errorf("fetch %s (page title %q): %w", url, title, err)
	}

	if f.opts.DumpDir != "" {
		if err := dumpHTML(f.opts.DumpDir, team.Abbr, season, html); err != nil {
			f.logger.Warn("dump html failed", slog.Any("err", err))
		}
	}
	return html, url, nil
}

// pollInterval は表の出現を確認する間隔。
const pollInterval = 500 * time.Millisecond

// loadRosterHTML は url に遷移し、ロスター表が現れるまでポーリングしてから HTML を返す。
//
// chromedp.Navigate / WaitReady / OuterHTML は使わない。これらは load イベントやフレーム状態に依存しており、
// スクレイプ先の bot 対策（Cloudflare）がチャレンジ画面から本ページへ遷移させる過程でハングすることがある。
// 代わりに遷移だけを発行し、JS の querySelector で表の存在を確認する。
func (f *Fetcher) loadRosterHTML(ctx context.Context, url string) (string, error) {
	navigate := chromedp.ActionFunc(func(ctx context.Context) error {
		_, _, _, _, err := page.Navigate(url).Do(ctx)
		return err
	})
	if err := chromedp.Run(ctx, navigate); err != nil {
		return "", fmt.Errorf("navigate: %w", err)
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ticker.C:
		}

		var ready bool
		if err := chromedp.Run(ctx, chromedp.Evaluate(`!!document.querySelector(`+strconv.Quote(rosterTableSelector)+`)`, &ready)); err != nil {
			// チャレンジ画面から本ページへ遷移中は実行コンテキストが破棄されて失敗することがあるので、次の tick で再試行する
			f.logger.Debug("poll roster table", slog.Any("err", err))
			continue
		}
		if !ready {
			continue
		}

		var html string
		if err := chromedp.Run(ctx, chromedp.Evaluate(`document.documentElement.outerHTML`, &html)); err != nil {
			return "", fmt.Errorf("get html: %w", err)
		}
		return html, nil
	}
}

// diagnose は取得失敗時にページタイトルを返し、DumpDir が設定されていれば現在の HTML を保存する。
// 取得済みのページ内容を見ることで、bot 対策のチャレンジ画面なのかセレクタ不一致なのかを切り分けられる。
func (f *Fetcher) diagnose(tabCtx context.Context, abbr string, season int) string {
	ctx, cancel := context.WithTimeout(tabCtx, 5*time.Second)
	defer cancel()

	var title, html string
	if err := chromedp.Run(ctx,
		chromedp.Evaluate(`document.title`, &title),
		chromedp.Evaluate(`document.documentElement.outerHTML`, &html),
	); err != nil {
		f.logger.Debug("diagnose failed", slog.Any("err", err))
		return ""
	}
	if f.opts.DumpDir != "" {
		if err := dumpHTML(f.opts.DumpDir, abbr+"_error", season, html); err != nil {
			f.logger.Warn("dump html failed", slog.Any("err", err))
		}
	}
	return title
}

func dumpHTML(dir, abbr string, season int, html string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%s_%d_roster.html", abbr, season))
	if err := os.WriteFile(path, []byte(html), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
