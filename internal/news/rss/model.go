// Package rss はRSSフィード関連の型定義を提供する。
package rss

// Link はRSSフィードのリンクとその分類情報を表す。
type Link struct {
	Group string
	Name  string
	URL   string
}

// LinkQuery はRSSリンクを検索するときのフィルター条件。
type LinkQuery struct {
	Group *string
	Name  *string
}
