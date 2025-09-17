// Package file は設定ファイルの読み込み機能を提供する。
package file

import (
	"encoding/json"
	"fmt"
	"sort"
)

// LoadGroupedStringEntriesJSON は YAML 版と同じ三つ組モデルで JSON ファイルを読み込む。
// 想定フォーマット: { "group": { "name": "value", ... }, ... }
func LoadGroupedStringEntriesJSON(path string) ([]GroupedStringEntry, error) {
	if path == "" {
		return nil, fmt.Errorf("ファイルパスが空です")
	}

	b, err := LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	var raw map[string]map[string]string
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, fmt.Errorf("JSONのパースに失敗しました: %w", err)
	}

	entries := make([]GroupedStringEntry, 0, 16)
	for g, m := range raw {
		for n, v := range m {
			if g == "" || n == "" || v == "" {
				return nil, fmt.Errorf("空のキーまたは値が含まれています (group=%q name=%q)", g, n)
			}
			entries = append(entries, GroupedStringEntry{Group: g, Name: n, Value: v})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Group == entries[j].Group {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Group < entries[j].Group
	})

	return entries, nil
}
