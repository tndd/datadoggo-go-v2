package link

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// SearchArticleLinks はarticle_linksテーブルを検索し、条件に合致する記事リンクを降順で返す。
func SearchArticleLinks(ctx context.Context, db *sql.DB, query *ArticleLinkQuery) (links []ArticleLink, err error) {
	if db == nil {
		return nil, errors.New("データベース接続が未初期化です")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	builder := strings.Builder{}
	builder.WriteString(baseSearchArticleLinksSQL)
	builder.WriteString(" WHERE 1=1")

	args := make([]any, 0, 3)

	if query != nil {
		if query.LinkPattern != nil && *query.LinkPattern != "" {
			builder.WriteString(fmt.Sprintf(" AND url ILIKE '%%' || $%d || '%%'", len(args)+1))
			args = append(args, *query.LinkPattern)
		}
		if query.PubDateFrom != nil {
			builder.WriteString(fmt.Sprintf(" AND pub_date >= $%d", len(args)+1))
			args = append(args, query.PubDateFrom.UTC())
		}
		if query.PubDateTo != nil {
			builder.WriteString(fmt.Sprintf(" AND pub_date <= $%d", len(args)+1))
			args = append(args, query.PubDateTo.UTC())
		}
	}

	builder.WriteString(" ORDER BY pub_date DESC")

	rows, err := db.QueryContext(ctx, builder.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("記事リンクの検索に失敗しました: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("記事リンクの読み出し後処理に失敗しました: %w", closeErr)
		}
	}()

	links = make([]ArticleLink, 0, 16)
	for rows.Next() {
		var link ArticleLink
		if scanErr := rows.Scan(&link.URL, &link.Title, &link.PubDate, &link.Source); scanErr != nil {
			err = fmt.Errorf("記事リンクのスキャンに失敗しました: %w", scanErr)
			return nil, err
		}
		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("記事リンク読み出し中にエラーが発生しました: %w", err)
	}

	return links, nil
}
