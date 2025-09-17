package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
    "strings"
)

const maxCallerDepth = 16

var projectRoot = detectProjectRoot()

// LoadFile はファイルパス文字列を解析して対象ファイルを読み込む。
// 目的: "./" で始まる場合は呼び出し元ファイルを基準に、それ以外はプロジェクトルート起点での解決を提供する。
func LoadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("ファイルパスが空です")
	}

	if strings.HasPrefix(path, "./") {
		localPath, err := resolveFromCaller(path[2:])
		if err != nil {
			return nil, err
		}
		return readFile(localPath)
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

func resolveFromCaller(relative string) (string, error) {
	relative = filepath.Clean(relative)

	for depth := 1; depth <= maxCallerDepth; depth++ {
		_, callerFile, _, ok := runtime.Caller(depth)
		if !ok {
			break
		}

        // 自身（このパッケージ）の内部からの呼び出しはスキップする
        if strings.HasSuffix(callerFile, "internal/infra/file/file.go") {
            continue
        }

		baseDir := filepath.Dir(callerFile)
		candidate := filepath.Join(baseDir, relative)
		if fileExists(candidate) {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("呼び出し元からの相対パス解決に失敗しました: ./%s", relative)
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
