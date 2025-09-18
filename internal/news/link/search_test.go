package link

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	infradb "datadoggo-go-v2/internal/infra/db"
)

func setupSearchLinkContainer(ctx context.Context, t *testing.T, migrationsDir string, seedArticles []ArticleLink) *infradb.TestContainer {
	t.Helper()

	container, err := infradb.SetupTestContainer(ctx, migrationsDir)
	if err != nil {
		if errors.Is(err, infradb.ErrDockerUnavailable) {
			t.Skipf("Dockerが利用できないためスキップします: %v", err)
		}
		t.Fatalf("テスト用DBの準備に失敗しました: %v", err)
	}
	if err := StoreArticleLinks(ctx, container.DB, seedArticles); err != nil {
		t.Fatalf("シードデータ投入に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		if closeErr := container.Close(ctx); closeErr != nil {
			t.Fatalf("テスト用DBのクローズに失敗しました: %v", closeErr)
		}
	})
	return container
}

func articleSeeds() []ArticleLink {
	return []ArticleLink{
		{
			URL:     "https://example.com/articles/alpha",
			Title:   "アルファ",
			PubDate: time.Date(2025, time.January, 15, 12, 0, 0, 0, time.UTC),
			Source:  "rss",
		},
		{
			URL:     "https://example.com/articles/beta",
			Title:   "ベータ",
			PubDate: time.Date(2025, time.February, 1, 9, 0, 0, 0, time.UTC),
			Source:  "rss",
		},
		{
			URL:     "https://another.example.com/articles/gamma",
			Title:   "ガンマ",
			PubDate: time.Date(2024, time.December, 25, 20, 30, 0, 0, time.UTC),
			Source:  "rss",
		},
		{
			URL:     "https://news.example.com/articles/delta",
			Title:   "デルタ",
			PubDate: time.Date(2025, time.February, 10, 7, 15, 0, 0, time.UTC),
			Source:  "rss",
		},
	}
}

// TestSearchArticleLinksAllRecords はクエリなしで全件取得されるかを検証する。
// 目的: 検索結果が公開日時の降順で並び、件数が一致することを確認する。
// 観点:
// - 4件取得される
// - 公開日時が降順ソートされている
func TestSearchArticleLinksAllRecords(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	migrationsDir, err := infradb.DefaultMigrationsDir()
	if err != nil {
		t.Fatalf("migrationsディレクトリの解決に失敗しました: %v", err)
	}

	seeds := articleSeeds()
	container := setupSearchLinkContainer(ctx, t, migrationsDir, seeds)

	results, err := SearchArticleLinks(ctx, container.DB, nil)
	if err != nil {
		t.Fatalf("SearchArticleLinksの実行に失敗しました: %v", err)
	}
	if len(results) != len(seeds) {
		t.Fatalf("期待件数と異なります。期待=%d 実際=%d", len(seeds), len(results))
	}

	for i := 1; i < len(results); i++ {
		if results[i-1].PubDate.Before(results[i].PubDate) {
			t.Fatalf("公開日時が降順になっていません: %v < %v", results[i-1].PubDate, results[i].PubDate)
		}
	}
}

// TestSearchArticleLinksFilterByPattern はURL部分一致検索を検証する。
// 目的: パターンに一致する2件のみが抽出されることを確認する。
// 観点:
// - 取得件数が2件
// - 取得結果が期待するURLパターンを含む
func TestSearchArticleLinksFilterByPattern(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	migrationsDir, err := infradb.DefaultMigrationsDir()
	if err != nil {
		t.Fatalf("migrationsディレクトリの解決に失敗しました: %v", err)
	}

	seeds := articleSeeds()
	container := setupSearchLinkContainer(ctx, t, migrationsDir, seeds)

	pattern := "https://example.com/articles/"
	query := &ArticleLinkQuery{LinkPattern: &pattern}
	results, err := SearchArticleLinks(ctx, container.DB, query)
	if err != nil {
		t.Fatalf("SearchArticleLinksの実行に失敗しました: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("部分一致の件数が期待と異なります: %d", len(results))
	}
	for _, link := range results {
		if !strings.Contains(link.URL, "example.com/articles") {
			t.Fatalf("期待外のURLが含まれています: %s", link.URL)
		}
	}
}

// TestSearchArticleLinksFilterByRange は日時範囲指定の検索を検証する。
// 目的: 境界値を含む範囲指定で該当レコードのみ抽出されることを確認する。
// 観点:
// - 取得件数が2件
// - 取得した記事が範囲内の日時である
func TestSearchArticleLinksFilterByRange(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	migrationsDir, err := infradb.DefaultMigrationsDir()
	if err != nil {
		t.Fatalf("migrationsディレクトリの解決に失敗しました: %v", err)
	}

	seeds := articleSeeds()
	container := setupSearchLinkContainer(ctx, t, migrationsDir, seeds)

	from := time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2025, time.February, 10, 23, 59, 59, 0, time.UTC)
	query := &ArticleLinkQuery{PubDateFrom: &from, PubDateTo: &to}

	results, err := SearchArticleLinks(ctx, container.DB, query)
	if err != nil {
		t.Fatalf("SearchArticleLinksの実行に失敗しました: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("日時フィルターの件数が期待と異なります: %d", len(results))
	}
	for _, link := range results {
		if link.PubDate.Before(from) || link.PubDate.After(to) {
			t.Fatalf("範囲外のデータが含まれています: %s", link.PubDate)
		}
	}
}
