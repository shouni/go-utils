package retry

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
)

const (
	DefaultMaxRetries      = 3
	InitialBackoffInterval = 5 * time.Second
	MaxBackoffInterval     = 30 * time.Second
)

// Operation はリトライ可能な処理を表す関数です。
type Operation func() error

// ShouldRetryFunc はエラーを受け取り、そのエラーがリトライ可能かどうかを判定します。
type ShouldRetryFunc func(error) bool

// Config はリトライ動作を設定するための構造体です。
type Config struct {
	MaxRetries      uint64
	InitialInterval time.Duration
	MaxInterval     time.Duration
}

// DefaultConfig は推奨されるデフォルト設定を返します。
func DefaultConfig() Config {
	return Config{
		MaxRetries:      DefaultMaxRetries,
		InitialInterval: InitialBackoffInterval,
		MaxInterval:     MaxBackoffInterval,
	}
}

// newBackOffPolicy は設定とコンテキストから backoff.BackOff を生成します。
func newBackOffPolicy(ctx context.Context, cfg Config) backoff.BackOff {
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = cfg.InitialInterval
	b.MaxInterval = cfg.MaxInterval
	// 指数増加の倍率(デフォルト1.5)や、ランダムな揺らぎ(Jitter: デフォルト0.5)を
	// 調整したい場合はここで設定可能です。

	// 最大リトライ回数を制限
	bo := backoff.WithMaxRetries(b, cfg.MaxRetries)
	// コンテキストによるキャンセル・タイムアウトを監視
	return backoff.WithContext(bo, ctx)
}

// Do は指数バックオフとカスタムエラー判定を使用して操作をリトライします。
func Do(ctx context.Context, cfg Config, operationName string, op Operation, shouldRetryFn ShouldRetryFunc) error {
	bo := newBackOffPolicy(ctx, cfg)

	var lastErr error
	var isPermanent bool

	retryableOp := func() error {
		err := op()
		if err == nil {
			return nil
		}
		lastErr = err // 最後のエラーを保持

		// リトライ不要判定
		if shouldRetryFn != nil && !shouldRetryFn(err) {
			isPermanent = true
			return backoff.Permanent(err)
		}
		return err
	}

	err := backoff.Retry(retryableOp, bo)
	if err != nil {
		// 1. 永続的エラーとしてマークされている場合、または errors.As で判定できる場合
		var pErr *backoff.PermanentError
		if isPermanent || errors.As(err, &pErr) {
			// pErr が取れない場合は lastErr を使う
			finalErr := lastErr
			if pErr != nil {
				finalErr = pErr.Err
			}
			return fmt.Errorf("%sに失敗しました: 致命的なエラーのため中止: %w", operationName, finalErr)
		}

		// 2. コンテキストのキャンセルまたはタイムアウト
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("%sに失敗しました: タイムアウトまたはキャンセルされました: %w", operationName, err)
		}

		// 3. 最大リトライ回数到達
		return fmt.Errorf("%sに失敗しました: 最大リトライ回数 (%d回) を超えました。最終エラー: %w", operationName, cfg.MaxRetries, err)
	}

	return nil
}
