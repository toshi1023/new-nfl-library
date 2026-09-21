// Package csvwriter はスクレイプ結果を CSV ファイルに出力する。
package csvwriter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// writeFile は path に write の内容を書き出す。出力先ディレクトリが無ければ作成し、既存ファイルは上書きする。
// ロスター用・スターター用の Writer で共通のファイル I/O（ディレクトリ作成・生成・クローズ）をまとめたもの。
func writeFile(path string, write func(io.Writer) error) (err error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dir, err)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close %s: %w", path, cerr)
		}
	}()

	if err := write(f); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
