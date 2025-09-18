package db

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestRunMigrationsCreatesTable はRunMigrationsがarticle_linksテーブルを生成するかを検証する。
// 目的: マイグレーションファイルを適用するとテーブルが作成されクエリが実行可能になることを確認する。
// 観点:
// - マイグレーション実行が成功する
// - テーブルにINSERT/SELECTが通る
func TestRunMigrationsCreatesTable(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := SetupTestContainer(ctx, "")
	if err != nil {
		if errors.Is(err, ErrDockerUnavailable) {
			t.Skipf("Dockerが利用できないためスキップします: %v", err)
		}
		t.Fatalf("テスト用コンテナの起動に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Close(ctx)
	})

	migrationsDir, err := DefaultMigrationsDir()
	if err != nil {
		t.Fatalf("migrationsディレクトリの解決に失敗しました: %v", err)
	}

	if migrationErr := RunMigrations(ctx, container.DB, migrationsDir); migrationErr != nil {
		t.Fatalf("マイグレーションに失敗しました: %v", migrationErr)
	}

	_, err = container.DB.ExecContext(ctx, `
		INSERT INTO article_links (url, title, pub_date, source)
		VALUES ('https://example.com/db-test', 'DBテスト', NOW(), 'test')
	`)
	if err != nil {
		t.Fatalf("INSERTが失敗しました: %v", err)
	}

	var count int
	if err := container.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM article_links").Scan(&count); err != nil {
		t.Fatalf("SELECT COUNT(*) が失敗しました: %v", err)
	}
	if count == 0 {
		t.Fatal("article_linksにデータが挿入されていません")
	}
}

// TestRunMigrationsInvalidDir は存在しないディレクトリを指定した場合のエラーを検証する。
// 目的: 入力検証が正しく動作し想定外の場所を参照しないことを担保する。
// 観点:
// - エラーが返る
// - エラーメッセージにディレクトリが含まれる
func TestRunMigrationsInvalidDir(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	container, err := SetupTestContainer(ctx, "")
	if err != nil {
		if errors.Is(err, ErrDockerUnavailable) {
			t.Skipf("Dockerが利用できないためスキップします: %v", err)
		}
		t.Fatalf("テスト用コンテナの起動に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Close(ctx)
	})

	err = RunMigrations(ctx, container.DB, "./not-found-dir")
	if err == nil {
		t.Fatal("エラーが返るべきです")
	}
	if got, want := err.Error(), "not-found-dir"; !strings.Contains(got, want) {
		t.Fatalf("エラーメッセージが期待と異なります: %s", got)
	}
}
