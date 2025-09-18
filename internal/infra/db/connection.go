// Package db はデータベース接続およびマイグレーション関連の共通処理を提供する。
package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	// PostgreSQLドライバを利用するためのブランクインポート。
	_ "github.com/lib/pq"
)

// Option はデータベース接続のパラメータを調整するための関数。
type Option func(db *sql.DB)

// WithMaxOpenConns は最大同時接続数を指定する。
func WithMaxOpenConns(limit int) Option {
	return func(db *sql.DB) {
		db.SetMaxOpenConns(limit)
	}
}

// WithMaxIdleConns はアイドル状態の接続数を指定する。
func WithMaxIdleConns(limit int) Option {
	return func(db *sql.DB) {
		db.SetMaxIdleConns(limit)
	}
}

// WithConnMaxLifetime は接続の最大ライフタイムを指定する。
func WithConnMaxLifetime(d time.Duration) Option {
	return func(db *sql.DB) {
		db.SetConnMaxLifetime(d)
	}
}

// Connect はポストグレス用の*sql.DBを生成し、Pingで接続確認を行う。
func Connect(ctx context.Context, dsn string, opts ...Option) (*sql.DB, error) {
	if dsn == "" {
		return nil, errors.New("データベース接続文字列が指定されていません")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("データベース接続の初期化に失敗しました: %w", err)
	}

	for _, opt := range opts {
		opt(db)
	}

	if err := db.PingContext(ctx); err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			return nil, fmt.Errorf("データベースへの接続確認に失敗しました: %v (close error: %v)", err, closeErr)
		}
		return nil, fmt.Errorf("データベースへの接続確認に失敗しました: %w", err)
	}

	return db, nil
}
