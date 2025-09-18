package link

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

const baseSearchArticleLinksSQL = `
SELECT url, title, pub_date, source
FROM article_links
WHERE
    (CAST(:link_pattern AS TEXT) IS NULL OR url ILIKE '%' || CAST(:link_pattern AS TEXT) || '%')
    AND (CAST(:pub_date_from AS TIMESTAMPTZ) IS NULL OR pub_date >= CAST(:pub_date_from AS TIMESTAMPTZ))
    AND (CAST(:pub_date_to AS TIMESTAMPTZ) IS NULL OR pub_date <= CAST(:pub_date_to AS TIMESTAMPTZ))
ORDER BY pub_date DESC
`

type articleLinkSearchParams struct {
	LinkPattern sql.NullString `db:"link_pattern"`
	PubDateFrom sql.NullTime   `db:"pub_date_from"`
	PubDateTo   sql.NullTime   `db:"pub_date_to"`
}

// SearchArticleLinks はarticle_linksテーブルを検索し、条件に合致する記事リンクを降順で返す。
func SearchArticleLinks(ctx context.Context, db *sql.DB, query *ArticleLinkQuery) (links []ArticleLink, err error) {
	if db == nil {
		return nil, errors.New("データベース接続が未初期化です")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	params := buildArticleLinkSearchParams(query)

	namedQuery, args, err := sqlx.Named(baseSearchArticleLinksSQL, params)
	if err != nil {
		return nil, fmt.Errorf("検索条件の準備に失敗しました: %w", err)
	}
	queryWithBinding := sqlx.Rebind(sqlx.DOLLAR, namedQuery)

	sqlxdb := sqlx.NewDb(db, "postgres")
	links = make([]ArticleLink, 0, 16)
	if err := sqlxdb.SelectContext(ctx, &links, queryWithBinding, args...); err != nil {
		return nil, fmt.Errorf("記事リンクの検索に失敗しました: %w", err)
	}

	return links, nil
}

func buildArticleLinkSearchParams(query *ArticleLinkQuery) articleLinkSearchParams {
	params := articleLinkSearchParams{}
	if query == nil {
		return params
	}

	if query.LinkPattern != nil {
		pattern := strings.TrimSpace(*query.LinkPattern)
		if pattern != "" {
			params.LinkPattern = sql.NullString{String: pattern, Valid: true}
		}
	}

	if query.PubDateFrom != nil {
		from := query.PubDateFrom.UTC()
		params.PubDateFrom = sql.NullTime{Time: from, Valid: true}
	}

	if query.PubDateTo != nil {
		to := query.PubDateTo.UTC()
		params.PubDateTo = sql.NullTime{Time: to, Valid: true}
	}

	return params
}
