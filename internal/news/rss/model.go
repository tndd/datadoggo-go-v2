package rss

// RssLink はRSSフィードのリンクとその分類情報を表す。
type RssLink struct {
	Group string
	Name  string
	URL   string
}

// RssLinkQuery はRSSリンクを検索するときのフィルター条件。
type RssLinkQuery struct {
	Group *string
	Name  *string
}
