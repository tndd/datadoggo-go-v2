package file

import "testing"

// TestLoadGroupedStringEntriesJSON は JSON 版の読み込み挙動を検証する。
// 目的: グループ→名称→値の3要素が正しく復元されること。
// 観点:
// - 正常に読み込める
// - 代表的な1件の値が一致する
// - 並び順はグループ→名称の安定ソートである
func TestLoadGroupedStringEntriesJSON(t *testing.T) {
	t.Parallel()

	entries, err := LoadGroupedStringEntriesJSON("internal/infra/file/testdata/mock.json")
	if err != nil {
		t.Fatalf("LoadGroupedStringEntriesJSON でエラー: %v", err)
	}
	if len(entries) < 2 {
		t.Fatalf("想定より少ない件数です: %d", len(entries))
	}

	// 具体値検証
	var found bool
	for _, e := range entries {
		if e.Group == "bbc" && e.Name == "world" {
			if e.Value != "https://feeds.bbci.co.uk/news/world/rss.xml" {
				t.Fatalf("bbc/world の値が想定と異なります: %s", e.Value)
			}
			found = true
			break
		}
	}
	if !found {
		t.Fatal("bbc/world エントリが見つかりません")
	}

	// 安定ソートの軽い検証（先頭と末尾の関係のみ）
	if !(entries[0].Group <= entries[len(entries)-1].Group) {
		t.Fatal("ソート順が想定外です")
	}
}
