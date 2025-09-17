package rss

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"datadoggo-go-v2/internal/infra"
)

var rssLinkFilePath = defaultRssLinkFilePath()

// SearchRssLinks はlink.ymlを読み込み、クエリ条件に一致するRSSリンクを返す。
func SearchRssLinks(query *RssLinkQuery) ([]RssLink, error) {
	links, err := LoadRssLinks(rssLinkFilePath)
	if err != nil {
		return nil, err
	}

	var filtered []RssLink
	for _, link := range links {
		if query != nil && query.Group != nil && link.Group != *query.Group {
			continue
		}
		if query != nil && query.Name != nil && link.Name != *query.Name {
			continue
		}
		filtered = append(filtered, link)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Group == filtered[j].Group {
			return filtered[i].Name < filtered[j].Name
		}
		return filtered[i].Group < filtered[j].Group
	})

	return filtered, nil
}

// LoadRssLinks はYAMLファイルを読み込み、RssLinkの配列として返す。
func LoadRssLinks(filePath string) ([]RssLink, error) {
	if filePath == "" {
		return nil, errors.New("ファイルパスが空です")
	}

	normalizedPath := filepath.Clean(filePath)
	content, err := infra.LoadFile(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("RSSリンクファイルの読み込みに失敗: %w", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(content))
	var links []RssLink
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
			return nil, fmt.Errorf("名称とURLの区切りが見つかりません (行番号:%d 行内容:%s)", lineNumber, rawLine)
		}

		parts := strings.SplitN(trimmed, ":", 2)
		name := strings.TrimSpace(parts[0])
		url := strings.TrimSpace(parts[1])
		if name == "" {
			return nil, fmt.Errorf("名称が空です (行番号:%d)", lineNumber)
		}
		if url == "" {
			return nil, fmt.Errorf("URLが空です (行番号:%d)", lineNumber)
		}

		links = append(links, RssLink{Group: currentGroup, Name: name, URL: url})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("RSSリンクファイルの読み込み中にエラーが発生しました: %w", err)
	}

	return links, nil
}

// countLeadingSpaces は行頭の半角スペース数を数えるヘルパー。
func countLeadingSpaces(line string) int {
	trimmed := strings.TrimLeft(line, " ")
	return len(line) - len(trimmed)
}

// defaultRssLinkFilePath はこのファイルと同一ディレクトリのlink.ymlへの絶対パスを返す。
func defaultRssLinkFilePath() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Join(".", "link.yml")
	}
	return filepath.Join(filepath.Dir(filename), "link.yml")
}
