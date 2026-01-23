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

// TestConfigWithDefaults は、Config.withDefaults() メソッドが
// 0値のフィールドをデフォルト値で適切に補完することを確認します。
func TestConfigWithDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    Config
		expected Config
	}{
		{
			name:  "AllZeroValues_ShouldUseDefaults",
			input: Config{},
			expected: Config{
				MaxRetries:      DefaultMaxRetries,
				InitialInterval: InitialBackoffInterval,
				MaxInterval:     MaxBackoffInterval,
			},
		},
		{
			name: "OnlyMaxRetries_ShouldFillOthers",
			input: Config{
				MaxRetries: 5,
			},
			expected: Config{
				MaxRetries:      5,
				InitialInterval: InitialBackoffInterval,
				MaxInterval:     MaxBackoffInterval,
			},
		},
		{
			name: "OnlyInitialInterval_ShouldFillOthers",
			input: Config{
				InitialInterval: 10 * time.Second,
			},
			expected: Config{
				MaxRetries:      DefaultMaxRetries,
				InitialInterval: 10 * time.Second,
				MaxInterval:     MaxBackoffInterval,
			},
		},
		{
			name: "OnlyMaxInterval_ShouldFillOthers",
			input: Config{
				MaxInterval: 60 * time.Second,
			},
			expected: Config{
				MaxRetries:      DefaultMaxRetries,
				InitialInterval: InitialBackoffInterval,
				MaxInterval:     60 * time.Second,
			},
		},
		{
			name: "PartiallySet_MaxRetriesAndInitialInterval",
			input: Config{
				MaxRetries:      10,
				InitialInterval: 2 * time.Second,
			},
			expected: Config{
				MaxRetries:      10,
				InitialInterval: 2 * time.Second,
				MaxInterval:     MaxBackoffInterval,
			},
		},
		{
			name: "AllFieldsSet_ShouldNotChange",
			input: Config{
				MaxRetries:      7,
				InitialInterval: 3 * time.Second,
				MaxInterval:     45 * time.Second,
			},
			expected: Config{
				MaxRetries:      7,
				InitialInterval: 3 * time.Second,
				MaxInterval:     45 * time.Second,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.withDefaults()

			require.Equal(t, tt.expected.MaxRetries, result.MaxRetries,
				"MaxRetries mismatch")
			require.Equal(t, tt.expected.InitialInterval, result.InitialInterval,
				"InitialInterval mismatch")
			require.Equal(t, tt.expected.MaxInterval, result.MaxInterval,
				"MaxInterval mismatch")
		})
	}
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

// TestNewBackOffPolicy_WithDefaults は、newBackOffPolicy が
// Config に依存せず、渡された値をそのまま使用することを確認します。
func TestNewBackOffPolicy_WithDefaults(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{
			name: "FullySpecifiedConfig",
			cfg: Config{
				MaxRetries:      5,
				InitialInterval: 10 * time.Millisecond,
				MaxInterval:     100 * time.Millisecond,
			},
		},
		{
			name: "ConfigWithDefaults_AlreadyApplied",
			cfg: Config{
				MaxRetries:      DefaultMaxRetries,
				InitialInterval: InitialBackoffInterval,
				MaxInterval:     MaxBackoffInterval,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			bo := newBackOffPolicy(ctx, tt.cfg)

			require.NotNil(t, bo, "BackOff should not be nil")
		})
	}
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

// TestDo_WithZeroValueConfig は、Config{} (0値) で呼び出した場合に
// デフォルト値が適用されて正常に動作することを確認します。
func TestDo_WithZeroValueConfig(t *testing.T) {
	t.Run("ZeroValueConfig_UsesDefaults", func(t *testing.T) {
		ctx := context.Background()
		zeroConfig := Config{} // すべてゼロ値

		attempt := 0
		operation := func() error {
			attempt++
			if attempt < 2 {
				return errors.New("transient error")
			}
			return nil
		}

		err := Do(ctx, zeroConfig, "test_op", operation, func(err error) bool {
			return true // 常にリトライ
		})

		require.NoError(t, err, "Should succeed with default config")
		require.GreaterOrEqual(t, attempt, 2, "Should have retried at least once")
	})

	t.Run("ZeroValueConfig_RespectsMaxRetries", func(t *testing.T) {
		ctx := context.Background()
		zeroConfig := Config{} // デフォルトで MaxRetries = 3

		attempt := 0
		operation := func() error {
			attempt++
			return errors.New("always fails")
		}

		err := Do(ctx, zeroConfig, "test_op", operation, func(err error) bool {
			return true // 常にリトライ
		})

		require.Error(t, err, "Should fail after max retries")
		require.Contains(t, err.Error(), "最大リトライ回数")
		// DefaultMaxRetries (3回) + 初回実行 = 4回の試行
		require.Equal(t, int(DefaultMaxRetries)+1, attempt,
			"Should have attempted initial try + max retries")
	})
}

// TestDo_ShouldRetryFuncNil は、shouldRetryFn が nil の場合に
// すべてのエラーがリトライ可能として扱われることを確認します。
func TestDo_ShouldRetryFuncNil(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxRetries:      2,
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     5 * time.Millisecond,
	}

	attempt := 0
	operation := func() error {
		attempt++
		if attempt < 3 {
			return errors.New("retryable error")
		}
		return nil
	}

	err := Do(ctx, cfg, "test_op", operation, nil) // shouldRetryFn = nil

	require.NoError(t, err, "Should eventually succeed")
	require.Equal(t, 3, attempt, "Should have retried until success")
}

// TestDo_OperationRetriesAndSucceeds は、
// 複数回リトライして最終的に成功するシナリオをテストします。
func TestDo_OperationRetriesAndSucceeds(t *testing.T) {
	ctx := context.Background()
	cfg := Config{
		MaxRetries:      5,
		InitialInterval: 1 * time.Millisecond,
		MaxInterval:     10 * time.Millisecond,
	}

	attempt := 0
	operation := func() error {
		attempt++
		if attempt <= 3 {
			return errors.New("temporary failure")
		}
		return nil
	}

	err := Do(ctx, cfg, "flaky_operation", operation, func(err error) bool {
		return err.Error() == "temporary failure"
	})

	require.NoError(t, err)
	require.Equal(t, 4, attempt, "Should have succeeded on 4th attempt")
}
