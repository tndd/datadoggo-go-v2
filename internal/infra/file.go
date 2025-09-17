package infra

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

const maxCallerDepth = 16

var projectRoot = detectProjectRoot()

// LoadFile はプロジェクトルートを起点にファイルを解決して読み込む。
// 目的: 既存のloaderの互換性を保ちつつ、明示的なプロジェクト相対パスでアクセスできるようにする。
func LoadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("ファイルパスが空です")
	}

	normalizedPath := filepath.Clean(path)

	if !filepath.IsAbs(normalizedPath) {
		base := projectRoot
		if base == "" {
			return nil, fmt.Errorf("プロジェクトルートが特定できません: %s", path)
		}
		normalizedPath = filepath.Join(base, normalizedPath)
	}

	return readFile(normalizedPath)
}

// LoadLocalFile は呼び出し元ファイルからの相対パスでファイルを読み込む。
// 目的: 各モジュールが自分のディレクトリを基準に設定/モックファイルへアクセスできるようにする。
func LoadLocalFile(relativePath string) ([]byte, error) {
	if relativePath == "" {
		return nil, fmt.Errorf("ファイルパスが空です")
	}

	normalizedPath := filepath.Clean(relativePath)

	for depth := 1; depth <= maxCallerDepth; depth++ {
		_, callerFile, _, ok := runtime.Caller(depth)
		if !ok {
			break
		}

		baseDir := filepath.Dir(callerFile)
		candidate := filepath.Join(baseDir, normalizedPath)
		if fileExists(candidate) {
			return readFile(candidate)
		}
	}

	return nil, fmt.Errorf("呼び出し元からの相対パス解決に失敗しました: %s", relativePath)
}

func readFile(absolutePath string) ([]byte, error) {
	file, err := os.Open(absolutePath)
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

// detectProjectRoot はgo.modが存在するディレクトリを探索する。
func detectProjectRoot() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	dir := filepath.Dir(thisFile)

	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
