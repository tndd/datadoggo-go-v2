package link

import (
	"context"
	"errors"
	"testing"
	"time"

	infradb "datadoggo-go-v2/internal/infra/db"
)

func setupStoreLinkContainer(ctx context.Context, t *testing.T, migrationsDir string) *infradb.TestContainer {
	t.Helper()

	container, err := infradb.SetupTestContainer(ctx, migrationsDir)
	if err != nil {
		if errors.Is(err, infradb.ErrDockerUnavailable) {
			t.Skipf("Dockerが利用できないためスキップします: %v", err)
		}
		t.Fatalf("テスト用DBの準備に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := container.Close(ctx); closeErr != nil {
			t.Fatalf("テスト用DBのクローズに失敗しました: %v", closeErr)
		}
	})
	return container
}

// TestStoreArticleLinksNewRecords は新規リンクの一括保存が成功するかを検証する。
// 目的: 3件の新規データを保存した際に件数が期待通り増えることを確認する。
// 観点:
// - StoreArticleLinksがエラーなく完了する
// - article_linksテーブルに3件が挿入される
func TestStoreArticleLinksNewRecords(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	migrationsDir, err := infradb.DefaultMigrationsDir()
	if err != nil {
		t.Fatalf("migrationsディレクトリの解決に失敗しました: %v", err)
	}

	container := setupStoreLinkContainer(ctx, t, migrationsDir)

	articles := []ArticleLink{
		{
			URL:     "https://example.com/article-1",
			Title:   "記事1",
			PubDate: time.Date(2025, time.September, 1, 10, 0, 0, 0, time.UTC),
			Source:  "rss",
		},
		{
			URL:     "https://example.com/article-2",
			Title:   "記事2",
			PubDate: time.Date(2025, time.September, 1, 11, 0, 0, 0, time.UTC),
			Source:  "rss",
		},
		{
			URL:     "https://example.com/article-3",
			Title:   "記事3",
			PubDate: time.Date(2025, time.September, 1, 12, 0, 0, 0, time.UTC),
			Source:  "rss",
		},
	}

	if err := StoreArticleLinks(ctx, container.DB, articles); err != nil {
		t.Fatalf("記事リンクの保存に失敗しました: %v", err)
	}

	var count int
	if err := container.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM article_links").Scan(&count); err != nil {
		t.Fatalf("件数取得に失敗しました: %v", err)
	}
	if count != len(articles) {
		t.Fatalf("期待件数と異なります。期待=%d 実際=%d", len(articles), count)
	}
}

// TestStoreArticleLinksUpsertExisting は既存URLに対するUPSERT挙動を検証する。
// 目的: 同一URLの再保存でタイトル・日時が上書きされ、新規URLは追加されることを確認する。
// 観点:
// - 既存URLのタイトル/日時が更新される
// - テーブル全体の件数が期待通りになる
func TestStoreArticleLinksUpsertExisting(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	migrationsDir, err := infradb.DefaultMigrationsDir()
	if err != nil {
		t.Fatalf("migrationsディレクトリの解決に失敗しました: %v", err)
	}

	container := setupStoreLinkContainer(ctx, t, migrationsDir)

	initial := []ArticleLink{
		{
			URL:     "https://example.com/unique",
			Title:   "初回タイトル",
			PubDate: time.Date(2025, time.August, 31, 8, 30, 0, 0, time.UTC),
			Source:  "rss",
		},
	}
	if err := StoreArticleLinks(ctx, container.DB, initial); err != nil {
		t.Fatalf("初回保存に失敗しました: %v", err)
	}

	updated := []ArticleLink{
		{
			URL:     "https://example.com/unique",
			Title:   "更新後タイトル",
			PubDate: time.Date(2025, time.August, 31, 9, 0, 0, 0, time.UTC),
			Source:  "rss",
		},
		{
			URL:     "https://example.com/new",
			Title:   "新規タイトル",
			PubDate: time.Date(2025, time.August, 31, 9, 30, 0, 0, time.UTC),
			Source:  "rss",
		},
	}

	if err := StoreArticleLinks(ctx, container.DB, updated); err != nil {
		t.Fatalf("更新処理に失敗しました: %v", err)
	}

	row := container.DB.QueryRowContext(ctx, "SELECT title, pub_date FROM article_links WHERE url = $1", "https://example.com/unique")
	var title string
	var pubDate time.Time
	if err := row.Scan(&title, &pubDate); err != nil {
		t.Fatalf("更新後データの取得に失敗しました: %v", err)
	}
	if title != "更新後タイトル" {
		t.Fatalf("タイトルが更新されていません: %s", title)
	}
	if !pubDate.Equal(updated[0].PubDate) {
		t.Fatalf("公開日時が更新されていません: %s", pubDate)
	}

	var count int
	if err := container.DB.QueryRowContext(ctx, "SELECT COUNT(*) FROM article_links").Scan(&count); err != nil {
		t.Fatalf("件数取得に失敗しました: %v", err)
	}
	if count != 2 {
		t.Fatalf("UPSERT後の件数が期待と異なります: %d", count)
	}
}
