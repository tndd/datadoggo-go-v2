package infra

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LoadFile は指定パスのファイルを読み込み内容をバイト列で返す。
// 目的: YAMLやJSONなど汎用的な設定ファイルをinfra層経由で取得できるようにする。
func LoadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("ファイルパスが空です")
	}

	normalizedPath := filepath.Clean(path)
	file, err := os.Open(normalizedPath)
	if err != nil {
		return nil, fmt.Errorf("ファイルのオープンに失敗しました: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ファイルの読み込みに失敗しました: %w", err)
	}

	return data, nil
}
