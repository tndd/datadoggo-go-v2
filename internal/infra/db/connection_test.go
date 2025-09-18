package db

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// TestConnectEmptyDSN はConnect関数の入力バリデーションを検証する。
// 目的: DSN未指定時にエラーが返されることを確認する。
// 観点:
// - エラーが発生する
// - エラーメッセージが期待する文言を含む
func TestConnectEmptyDSN(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	_, err := Connect(ctx, "")
	if err == nil {
		t.Fatal("エラーが返されるべきです")
	}
	if got, want := err.Error(), "接続文字列"; !strings.Contains(got, want) {
		t.Fatalf("エラーメッセージが期待と異なります: %s", got)
	}
}

// TestConnectWithTestContainer はTestContainerで取得したDSNに対してConnectが成功するかを確認する。
// 目的: 実データベースに接続してクエリが実行できることを担保する。
// 観点:
// - SetupTestContainerでDSNが取得できる
// - Connectで返ったDBに対してSELECT 1が成功する
func TestConnectWithTestContainer(t *testing.T) {
	t.Parallel()

	assumeDockerAvailable(t)

	ctx := context.Background()
	container, err := SetupTestContainer(ctx, "")
	if err != nil {
		if errors.Is(err, ErrDockerUnavailable) {
			t.Skipf("Dockerが利用できないためスキップします: %v", err)
		}
		t.Fatalf("テスト用コンテナの起動に失敗しました: %v", err)
	}
	t.Cleanup(func() {
		_ = container.Close(ctx)
	})

	db, err := Connect(ctx, container.DSN())
	if err != nil {
		t.Fatalf("Connectに失敗しました: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var dummy int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&dummy); err != nil {
		t.Fatalf("SELECT 1 が失敗しました: %v", err)
	}
}
