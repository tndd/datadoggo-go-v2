package file

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
)

// GroupedStringEntry は単純なグループ→名称→値の三つ組を表す。
type GroupedStringEntry struct {
	Group string
	Name  string
	Value string
}

// LoadGroupedStringEntries はlink.ymlのようなシンプルな入れ子構造を読み込み、三つ組へ変換する。
// 現状はスペース2個インデントの形式のみをサポートする。
func LoadGroupedStringEntries(path string) ([]GroupedStringEntry, error) {
	content, err := LoadFile(path)
	if err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	var entries []GroupedStringEntry
	var currentGroup string
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := countLeadingSpaces(rawLine)
		if indent == 0 {
			if !strings.HasSuffix(trimmed, ":") {
				return nil, fmt.Errorf("グループ行の形式が不正です (行番号:%d 行内容:%s)", lineNumber, rawLine)
			}
			group := strings.TrimSuffix(trimmed, ":")
			group = strings.TrimSpace(group)
			if group == "" {
				return nil, fmt.Errorf("グループ名が空です (行番号:%d)", lineNumber)
			}
			currentGroup = group
			continue
		}

		if indent != 2 {
			return nil, fmt.Errorf("サポートしていないインデント幅です (行番号:%d インデント:%d)", lineNumber, indent)
		}

		if currentGroup == "" {
			return nil, fmt.Errorf("グループ定義前に子要素があります (行番号:%d)", lineNumber)
		}

		if !strings.Contains(trimmed, ":") {
			return nil, fmt.Errorf("名称と値の区切りが見つかりません (行番号:%d 行内容:%s)", lineNumber, rawLine)
		}

		parts := strings.SplitN(trimmed, ":", 2)
		name := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if name == "" {
			return nil, fmt.Errorf("名称が空です (行番号:%d)", lineNumber)
		}
		if value == "" {
			return nil, fmt.Errorf("値が空です (行番号:%d)", lineNumber)
		}

		entries = append(entries, GroupedStringEntry{Group: currentGroup, Name: name, Value: value})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("設定ファイルの読み込み中にエラーが発生しました: %w", err)
	}

	return entries, nil
}

func countLeadingSpaces(line string) int {
	trimmed := strings.TrimLeft(line, " ")
	return len(line) - len(trimmed)
}
