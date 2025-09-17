package rss

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

var rssLinkFilePath = defaultRssLinkFilePath()

// SearchRssLinks はlink.ymlを読み込み、クエリ条件に一致するRSSリンクを返す。
func SearchRssLinks(query *RssLinkQuery) ([]RssLink, error) {
	rssMap, err := loadRssLinkMap(rssLinkFilePath)
	if err != nil {
		return nil, err
	}

	var links []RssLink
	for group, nameMap := range rssMap {
		if query != nil && query.Group != nil && group != *query.Group {
			continue
		}
		for name, url := range nameMap {
			if query != nil && query.Name != nil && name != *query.Name {
				continue
			}
			links = append(links, RssLink{Group: group, Name: name, URL: url})
		}
	}

	sort.Slice(links, func(i, j int) bool {
		if links[i].Group == links[j].Group {
			return links[i].Name < links[j].Name
		}
		return links[i].Group < links[j].Group
	})

	return links, nil
}

// loadRssLinkMap はYAML形式(グループ→名称→URL)を手動で解析して返す。
func loadRssLinkMap(filePath string) (RssLinkMap, error) {
	if filePath == "" {
		return nil, errors.New("ファイルパスが空です")
	}

	normalizedPath := filepath.Clean(filePath)
	file, err := os.Open(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("RSSリンクファイルの読み込みに失敗: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := make(RssLinkMap)
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
			if _, exists := result[group]; !exists {
				result[group] = make(map[string]string)
			}
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

		result[currentGroup][name] = url
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("RSSリンクファイルの読み込み中にエラーが発生しました: %w", err)
	}

	return result, nil
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
