// Package link は記事リンクに関するドメインロジックをまとめる。
package link

import "time"

// ArticleLink はニュース記事のリンク情報を表す。
type ArticleLink struct {
	URL     string    `json:"url"`
	Title   string    `json:"title"`
	PubDate time.Time `json:"pub_date"`
	Source  string    `json:"source"`
}

// ArticleLinkQuery はリンク検索時のフィルター条件を格納する。
type ArticleLinkQuery struct {
	LinkPattern *string    `json:"link_pattern"`
	PubDateFrom *time.Time `json:"pub_date_from"`
	PubDateTo   *time.Time `json:"pub_date_to"`
}
