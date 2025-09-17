package rss

import (
    "errors"
    "fmt"
    "sort"

    infrafile "datadoggo-go-v2/internal/infra/file"
)

const rssLinkFilePath = "./link.yml"

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

    entries, err := infrafile.LoadGroupedStringEntries(filePath)
	if err != nil {
		return nil, fmt.Errorf("RSSリンクファイルの読み込みに失敗: %w", err)
	}

	links := make([]RssLink, 0, len(entries))
	for _, entry := range entries {
		links = append(links, RssLink{Group: entry.Group, Name: entry.Name, URL: entry.Value})
	}

	return links, nil
}
