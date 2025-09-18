package link

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

// assumeDockerAvailable はDockerデーモンにアクセスできない場合はテストをスキップする。
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
		t.Skipf("Dockerが利用できないためテストをスキップします: %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Dockerクライアントのクローズに失敗しました: %v", err)
	}
}
