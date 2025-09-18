// Package db はデータベース接続およびマイグレーション関連の共通処理を提供する。
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// RunMigrations は指定ディレクトリ配下の .sql ファイルを名前順に実行する。
// 単純なファイルベースのマイグレーションであり、1ファイルを1トランザクションで流す。
func RunMigrations(ctx context.Context, db *sql.DB, dir string) error {
	if db == nil {
		return errors.New("接続プールが初期化されていません")
	}
	if dir == "" {
		return errors.New("マイグレーションディレクトリが指定されていません")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("マイグレーションディレクトリの読み込みに失敗しました: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}

	sort.Strings(files)

	for _, file := range files {
		if err := runSingleMigration(ctx, db, file); err != nil {
			return err
		}
	}

	return nil
}

func runSingleMigration(ctx context.Context, db *sql.DB, filePath string) (err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("マイグレーションファイルのオープンに失敗しました (%s): %w", filePath, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("マイグレーションファイルのクローズに失敗しました (%s): %w", filePath, closeErr)
		}
	}()

	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("マイグレーションファイルの読み込みに失敗しました (%s): %w", filePath, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("マイグレーション用トランザクションの開始に失敗しました (%s): %w", filePath, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	statements := strings.Split(string(content), ";")
	for _, statement := range statements {
		trimmed := strings.TrimSpace(statement)
		if trimmed == "" {
			continue
		}
		if _, execErr := tx.ExecContext(ctx, trimmed); execErr != nil {
			err = fmt.Errorf("マイグレーションの実行に失敗しました (%s): %w", filePath, execErr)
			return err
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return fmt.Errorf("マイグレーションのコミットに失敗しました (%s): %w", filePath, commitErr)
	}

	return nil
}
