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

// NewRssLinkQueryFromGroup はgroupだけを指定した検索条件を生成する。
func NewRssLinkQueryFromGroup(group string) RssLinkQuery {
	g := group
	return RssLinkQuery{Group: &g}
}

// RssLinkMap はYAML構造に対応するグループ→名称→URLのマップ。
type RssLinkMap map[string]map[string]string
