package link

import (
	"context"
	"errors"
	"testing"
	"time"

	newrss "datadoggo-go-v2/internal/news/rss"
)

type stubHTTPClient struct {
	body []byte
	err  error
}

func (s *stubHTTPClient) Fetch(_ context.Context, _ string, _ time.Duration) ([]byte, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.body, nil
}

// TestFetchArticleLinksUsingFeed はRSS取得から記事抽出までの成功パスとエラー伝搬を検証する。
// 目的: 正常ケースでitemがスキップなく抽出され、HTTPエラー時には適切に失敗することを確認する。
// 観点:
// - 正しいRSSを渡すと<item>がArticleLinkへ変換される
// - HTTP取得エラーが発生した場合にそのままエラーが伝播する
func TestFetchArticleLinksUsingFeed(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("正常なRSS", func(t *testing.T) {
		t.Parallel()

		xml := []byte(`<?xml version="1.0" encoding="UTF-8"?>
            <rss version="2.0">
                <channel>
                    <title>Example</title>
                    <item>
                        <title>記事A</title>
                        <link>https://example.com/a</link>
                        <pubDate>Mon, 15 Sep 2025 12:00:00 +0000</pubDate>
                    </item>
                    <item>
                        <title></title>
                        <link>https://example.com/b</link>
                        <pubDate>Mon, 15 Sep 2025 13:00:00 +0000</pubDate>
                    </item>
                    <item>
                        <title>無効な日付</title>
                        <link>https://example.com/c</link>
                        <pubDate>invalid</pubDate>
                    </item>
                </channel>
            </rss>`) // invalid itemはスキップされる

		client := &stubHTTPClient{body: xml}
		rssLink := &newrss.Link{Group: "test", Name: "feed", URL: "https://example.com/rss"}

		links, err := FetchArticleLinksUsingFeed(ctx, client, rssLink)
		if err != nil {
			t.Fatalf("FetchArticleLinksUsingFeed が失敗しました: %v", err)
		}
		if len(links) != 2 {
			t.Fatalf("抽出件数が期待と異なります: %d", len(links))
		}
		if links[0].Title != "記事A" {
			t.Fatalf("最初のタイトルが期待と異なります: %s", links[0].Title)
		}
		if links[1].Title != "タイトルなし" {
			t.Fatalf("空タイトルはデフォルト化されるべきです: %s", links[1].Title)
		}
	})

	t.Run("HTTP取得エラー", func(t *testing.T) {
		t.Parallel()

		client := &stubHTTPClient{err: errors.New("timeout")}
		rssLink := &newrss.Link{Group: "test", Name: "feed", URL: "https://example.com/rss"}

		if _, err := FetchArticleLinksUsingFeed(ctx, client, rssLink); err == nil {
			t.Fatal("エラーが期待されますが成功扱いになりました")
		}
	})
}
