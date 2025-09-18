package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// assumeDockerAvailable はDockerデーモンにアクセスできるかを判定し、不可ならテストをスキップする。
// 目的: TestContainerを利用するテストが環境非依存で動作するための前提条件チェック。
// 観点:
// - Dockerクライアント生成が成功する
// - 失敗した場合はエラー内容を添えてSkipする
func assumeDockerAvailable(t *testing.T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer func() {
		if r := recover(); r != nil {
			t.Skipf("Dockerが利用できないためテストをスキップします: %v", r)
		}
	}()

	client, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		if errors.Is(err, ErrDockerUnavailable) {
			t.Skipf("Dockerが利用できないためテストをスキップします: %v", err)
		}
		t.Skipf("Dockerクライアントを初期化できないためテストをスキップします: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Dockerクライアントのクローズに失敗しました: %v", err)
	}
}
