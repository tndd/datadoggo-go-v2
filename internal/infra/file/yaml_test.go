package file

import "testing"

// TestLoadGroupedStringEntries はシンプルなYAML風ファイルの読み込みを検証する。
// 目的: link.yml形式のグループ・名称・値の対応が正しく三つ組に変換されることを担保する。
// 観点:
// - 正常に読み込める
// - 特定エントリの値が期待通りである
// - 解析対象は testdata 配下の仮データを使用し、実運用データに依存しない
func TestLoadGroupedStringEntries(t *testing.T) {
	t.Parallel()

	entries, err := LoadGroupedStringEntries("internal/infra/file/testdata/mock.yml")
	if err != nil {
		t.Fatalf("LoadGroupedStringEntries でエラー: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("エントリが1件も読み込めませんでした")
	}

	var found bool
	for _, entry := range entries {
		if entry.Group == "bbc" && entry.Name == "world" {
			if entry.Value != "https://feeds.bbci.co.uk/news/world/rss.xml" {
				t.Fatalf("bbc/world の値が想定と異なります: %s", entry.Value)
			}
			found = true
			break
		}
	}

	if !found {
		t.Fatal("bbc/world エントリが見つかりません")
	}
}
