package retry

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	require.Equal(t, uint64(DefaultMaxRetries), cfg.MaxRetries, "MaxRetries should match DefaultMaxRetries constant.")
	require.Equal(t, InitialBackoffInterval, cfg.InitialInterval, "InitialInterval should match constant.")
	require.Equal(t, MaxBackoffInterval, cfg.MaxInterval, "MaxInterval should match constant.")
}

func TestNewBackOffPolicy(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxRetries:      5,
		InitialInterval: 10 * time.Millisecond,
		MaxInterval:     500 * time.Millisecond,
	}

	bo := newBackOffPolicy(ctx, cfg)
	require.NotNil(t, bo)
}

func TestDo(t *testing.T) {
	// テスト用の高速な設定
	testCfg := Config{MaxRetries: 3, InitialInterval: 1 * time.Millisecond, MaxInterval: 5 * time.Millisecond}
	opName := "test_operation"

	tests := []struct {
		name          string
		ctx           func() context.Context
		cfg           Config
		operationName string
		operation     Operation
		shouldRetry   ShouldRetryFunc
		expectedMsg   string
	}{
		{
			name:          "successful operation",
			ctx:           func() context.Context { return context.Background() },
			cfg:           testCfg,
			operationName: opName,
			operation:     func() error { return nil },
			shouldRetry:   func(err error) bool { return false },
			expectedMsg:   "",
		},
		{
			name:          "retryable error and success within max retries",
			ctx:           func() context.Context { return context.Background() },
			cfg:           testCfg,
			operationName: opName,
			operation: func() Operation {
				attempt := 0
				return func() error {
					attempt++
					if attempt < 3 {
						return errors.New("retryable error")
					}
					return nil
				}
			}(),
			shouldRetry: func(err error) bool { return err.Error() == "retryable error" },
			expectedMsg: "",
		},
		{
			name:          "permanent error via shouldRetryFn",
			ctx:           func() context.Context { return context.Background() },
			cfg:           testCfg,
			operationName: opName,
			operation:     func() error { return errors.New("fatal error") },
			shouldRetry:   func(err error) bool { return false }, // ここで確実に止める
			expectedMsg:   "致命的なエラーのため中止: fatal error",
		},
		{
			name: "context canceled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			cfg:           testCfg,
			operationName: opName,
			operation:     func() error { return errors.New("some error") },
			shouldRetry:   func(err error) bool { return true },
			expectedMsg:   "タイムアウトまたはキャンセルされました: context canceled",
		},
		{
			name: "context timeout",
			ctx: func() context.Context {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				time.Sleep(2 * time.Millisecond)
				defer cancel()
				return ctx
			},
			cfg:           testCfg,
			operationName: opName,
			operation:     func() error { return errors.New("some error") },
			shouldRetry:   func(err error) bool { return true },
			expectedMsg:   "タイムアウトまたはキャンセルされました: context deadline exceeded",
		},
		{
			name:          "max retries exceeded",
			ctx:           func() context.Context { return context.Background() },
			cfg:           testCfg,
			operationName: opName,
			operation: func() error {
				return errors.New("retryable error")
			},
			shouldRetry: func(err error) bool { return true },
			expectedMsg: fmt.Sprintf("%sに失敗しました: 最大リトライ回数 (%d回) を超えました。最終エラー: retryable error", opName, testCfg.MaxRetries),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Do(tt.ctx(), tt.cfg, tt.operationName, tt.operation, tt.shouldRetry)

			if tt.expectedMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.expectedMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
