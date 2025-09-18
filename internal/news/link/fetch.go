package link

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	newrss "datadoggo-go-v2/internal/news/rss"
)

// HTTPFetcher はRSSフィード取得に必要な最小限のインターフェース。
type HTTPFetcher interface {
	Fetch(ctx context.Context, url string, timeout time.Duration) ([]byte, error)
}

const defaultFetchTimeout = 30 * time.Second

// FetchArticleLinksUsingFeed はRSSリンクの情報を元にフィードを取得し、<item>に含まれる記事を抽出する。
func FetchArticleLinksUsingFeed(ctx context.Context, client HTTPFetcher, rssLink *newrss.Link) ([]ArticleLink, error) {
	if client == nil {
		return nil, errors.New("HTTPクライアントが指定されていません")
	}
	if rssLink == nil {
		return nil, errors.New("RSSリンクが指定されていません")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	body, err := client.Fetch(ctx, rssLink.URL, defaultFetchTimeout)
	if err != nil {
		return nil, fmt.Errorf("RSSフィードの取得に失敗しました (%s): %w", rssLink.URL, err)
	}

	items, err := extractItems(body)
	if err != nil {
		return nil, err
	}

	links := make([]ArticleLink, 0, len(items))
	for _, item := range items {
		if item.Link == "" {
			continue
		}
		published, err := parsePubDate(item.PubDate)
		if err != nil {
			continue
		}
		title := item.Title
		if title == "" {
			title = "タイトルなし"
		}
		links = append(links, ArticleLink{
			URL:     item.Link,
			Title:   title,
			PubDate: published,
			Source:  "rss",
		})
	}

	return links, nil
}

type rssDocument struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
}

func extractItems(data []byte) ([]rssItem, error) {
	var doc rssDocument
	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("RSSフィードの解析に失敗しました: %w", err)
	}
	return doc.Channel.Items, nil
}

var dateLayouts = []string{
	time.RFC1123Z,
	time.RFC1123,
	time.RFC3339,
	time.RFC3339Nano,
}

func parsePubDate(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, trimmed); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("サポートしていない日付形式です: %s", value)
}
