// Package link は記事リンクに関するドメインロジックをまとめる。
package link

import "time"

// ArticleLink はニュース記事のリンク情報を表す。
type ArticleLink struct {
	URL     string    `db:"url" json:"url"`
	Title   string    `db:"title" json:"title"`
	PubDate time.Time `db:"pub_date" json:"pub_date"`
	Source  string    `db:"source" json:"source"`
}

// ArticleLinkQuery はリンク検索時のフィルター条件を格納する。
type ArticleLinkQuery struct {
	LinkPattern *string    `json:"link_pattern"`
	PubDateFrom *time.Time `json:"pub_date_from"`
	PubDateTo   *time.Time `json:"pub_date_to"`
}
