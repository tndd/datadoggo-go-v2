package link

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

const upsertArticleLinkSQL = `
INSERT INTO article_links (url, title, pub_date, source)
VALUES ($1, $2, $3, $4)
ON CONFLICT (url) DO UPDATE SET
    title = EXCLUDED.title,
    pub_date = EXCLUDED.pub_date,
    source = EXCLUDED.source
WHERE (article_links.title, article_links.pub_date, article_links.source)
    IS DISTINCT FROM (EXCLUDED.title, EXCLUDED.pub_date, EXCLUDED.source)
`

// StoreArticleLinks は記事リンクの配列をバルクUPSERTする。
// 既存URLがあれば値を更新し、存在しなければ新規挿入する。
func StoreArticleLinks(ctx context.Context, db *sql.DB, links []ArticleLink) error {
	if db == nil {
		return errors.New("データベース接続が未初期化です")
	}
	if len(links) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("トランザクション開始に失敗しました: %w", err)
	}

	for _, link := range links {
		if _, err := tx.ExecContext(ctx, upsertArticleLinkSQL, link.URL, link.Title, link.PubDate.UTC(), link.Source); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("記事リンクの保存に失敗しました: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("トランザクションのコミットに失敗しました: %w", err)
	}

	return nil
}
