package infra

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestLoadFile はファイル読み込みユーティリティの基本動作を検証する。
// 目的: 既存の設定ファイルが正常に読み込めることと、存在しないパスでエラーになることを担保する。
// 観点:
// - 正常系でバイト列が取得できる
// - 異常系でエラーが返される
func TestLoadFile(t *testing.T) {
	t.Parallel()

	t.Run("正常系", func(t *testing.T) {
		t.Parallel()
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			t.Fatal("runtime.Caller に失敗しました")
		}
		baseDir := filepath.Dir(filename)
		path := filepath.Join(baseDir, "..", "news", "rss", "link.yml")
		data, err := LoadFile(path)
		if err != nil {
			t.Fatalf("LoadFile(%s) でエラー: %v", path, err)
		}
		if len(data) == 0 {
			t.Fatal("読み込んだデータが空です")
		}
	})

	t.Run("存在しないファイル", func(t *testing.T) {
		t.Parallel()
		_, err := LoadFile("internal/infra/does_not_exist.yml")
		if err == nil {
			t.Fatal("存在しないファイルなのにエラーが返されませんでした")
		}
	})
}
