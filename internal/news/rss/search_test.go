package rss

import "testing"

// TestSearchRssLinks はRSSリンク検索ロジックの正常系シナリオを検証する。
// 目的: link.ymlを読み込んだ際にフィルター条件が正しく適用されることを確認する。
// 観点:
// - フィルターなしで全件が取得できる
// - group指定で該当グループのみが取得できる
// - groupとname両方を指定した場合に単一のエントリに絞り込まれる
func TestSearchRssLinks(t *testing.T) {
	t.Parallel()

	t.Run("フィルターなし", func(t *testing.T) {
		t.Parallel()
		links, err := SearchRssLinks(nil)
		if err != nil {
			t.Fatalf("SearchRssLinks(nil) でエラー: %v", err)
		}
		if len(links) == 0 {
			t.Fatal("リンクが1件も取得できませんでした")
		}
	})

	t.Run("groupのみ指定", func(t *testing.T) {
		t.Parallel()
		group := "bbc"
		query := RssLinkQuery{Group: &group}
		links, err := SearchRssLinks(&query)
		if err != nil {
			t.Fatalf("SearchRssLinks(group) でエラー: %v", err)
		}
		if len(links) == 0 {
			t.Fatal("指定グループのリンクが取得できませんでした")
		}
		for _, link := range links {
			if link.Group != group {
				t.Fatalf("期待しないグループが含まれています: %+v", link)
			}
		}
	})

	t.Run("groupとnameで絞り込む", func(t *testing.T) {
		t.Parallel()
		group := "bbc"
		name := "world"
		query := RssLinkQuery{Group: &group, Name: &name}
		links, err := SearchRssLinks(&query)
		if err != nil {
			t.Fatalf("SearchRssLinks(group+name) でエラー: %v", err)
		}
		if len(links) != 1 {
			t.Fatalf("期待件数と異なります。期待=1 実際=%d", len(links))
		}
		link := links[0]
		if link.Group != group || link.Name != name {
			t.Fatalf("取得結果が期待と異なります: %+v", link)
		}
	})
}

// TestLoadRssLinks はlink.ymlの読み込み処理を詳細に検証する。
// 目的: 手動パーサーが単純なグループ→名称→URL構造を正しく読み取れることを担保する。
// 観点:
// - 既存ファイルの読み込みでエラーにならない
// - 一部グループの名称とURLが想定通りに格納される
func TestLoadRssLinks(t *testing.T) {
	t.Parallel()

	rssLinks, err := LoadRssLinks(rssLinkFilePath)
	if err != nil {
		t.Fatalf("LoadRssLinks でエラー: %v", err)
	}
	if len(rssLinks) == 0 {
		t.Fatal("グループが1件も読み込めませんでした")
	}

	var found bool
	want := RssLink{
		Group: "bbc",
		Name:  "world",
		URL:   "https://feeds.bbci.co.uk/news/world/rss.xml",
	}

	for _, link := range rssLinks {
		if link.Group == want.Group && link.Name == want.Name {
			if link.URL != want.URL {
				t.Fatalf("bbc/world のURLが想定と異なります。期待=%s 実際=%s", want.URL, link.URL)
			}
			found = true
			break
		}
	}

	if !found {
		t.Fatal("bbc/world が見つかりません")
	}
}
