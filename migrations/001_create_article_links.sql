-- 記事リンクを保持するテーブル。Rust版のarticle_linksと同じ構造。
CREATE TABLE IF NOT EXISTS article_links (
    url TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    pub_date TIMESTAMPTZ NOT NULL,
    source TEXT NOT NULL
);
