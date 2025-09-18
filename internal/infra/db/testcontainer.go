package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testUser     = "datadoggo"
	testPassword = "datadoggo"
	testDatabase = "datadoggo_test"
)

// ErrDockerUnavailable はDockerデーモンに接続できない場合に返される。
var ErrDockerUnavailable = errors.New("dockerが利用できません")

// TestContainer はテスト実行時に利用するPostgreSQLコンテナと接続を管理する。
type TestContainer struct {
	Container testcontainers.Container
	DB        *sql.DB
	dsn       string
}

// SetupTestContainer はPostgreSQLのテストコンテナを起動し、必要であればマイグレーションを適用する。
func SetupTestContainer(ctx context.Context, migrationsDir string) (*TestContainer, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       testDatabase,
			"POSTGRES_PASSWORD": testPassword,
			"POSTGRES_USER":     testUser,
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, wrapDockerError(fmt.Errorf("PostgreSQLコンテナの起動に失敗しました: %w", err))
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, wrapDockerError(fmt.Errorf("コンテナのホスト取得に失敗しました: %w", err))
	}

	mappedPort, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, wrapDockerError(fmt.Errorf("コンテナのポート解決に失敗しました: %w", err))
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", testUser, testPassword, host, mappedPort.Port(), testDatabase)
	db, err := Connect(ctx, dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, err
	}

	if migrationsDir != "" {
		if err := RunMigrations(ctx, db, migrationsDir); err != nil {
			if closeErr := db.Close(); closeErr != nil {
				err = fmt.Errorf("%w: DBクローズに失敗しました: %v", err, closeErr)
			}
			_ = container.Terminate(ctx)
			return nil, err
		}
	}

	return &TestContainer{Container: container, DB: db, dsn: dsn}, nil
}

// Close は接続を閉じた上でコンテナを停止する。
func (tc *TestContainer) Close(ctx context.Context) error {
	if tc == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	var resultErr error
	if tc.DB != nil {
		if err := tc.DB.Close(); err != nil {
			resultErr = fmt.Errorf("DBクローズに失敗しました: %w", err)
		}
	}
	if tc.Container != nil {
		if err := tc.Container.Terminate(ctx); err != nil {
			if resultErr != nil {
				return fmt.Errorf("コンテナの停止に失敗しました: %w (既存エラー: %v)", err, resultErr)
			}
			return fmt.Errorf("コンテナの停止に失敗しました: %w", err)
		}
	}
	return resultErr
}

// DSN はテスト用DBへの接続文字列を返す。
func (tc *TestContainer) DSN() string {
	if tc == nil {
		return ""
	}
	return tc.dsn
}

func wrapDockerError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "connect: operation not permitted") ||
		strings.Contains(msg, "Cannot connect to the Docker daemon") ||
		strings.Contains(strings.ToLower(msg), "rootless docker not found") {
		return fmt.Errorf("%w: %s", ErrDockerUnavailable, msg)
	}
	return err
}
