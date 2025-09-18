// Package db はデータベース接続およびマイグレーション関連の共通処理を提供する。
package db

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// DefaultMigrationsDir はリポジトリルート配下のmigrationsディレクトリへの絶対パスを返す。
func DefaultMigrationsDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("呼び出し元のファイルパス取得に失敗しました")
	}

	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	dir := filepath.Join(root, "migrations")
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("migrationsディレクトリの検出に失敗しました: %w", err)
	}

	return dir, nil
}
