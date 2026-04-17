package notifier

import (
    "context"
    "fmt"
    "time"

    "github.com/vincent119/zlogger"
)

// Notifier 定義通知發送介面
type Notifier interface {
    Send(ctx context.Context, message string, chatIDs []string) error
}

// sendWithRetry 對 fn 執行重試邏輯，最多 maxRetries 次，間隔 interval。
// channel 和 chatID 用於 log 識別。
func sendWithRetry(ctx context.Context, channel, chatID string, maxRetries int, interval time.Duration, fn func() error) error {
    var lastErr error
    for i := 0; i < maxRetries; i++ {
        if err := ctx.Err(); err != nil {
            // context 已取消或超時
            zlogger.Warn("context cancelled, aborting retry",
                zlogger.String("channel", channel),
                zlogger.String("chat_id", chatID),
                zlogger.Int("attempt", i+1),
                zlogger.Err(err),
            )
            return fmt.Errorf("%s send to %s failed (context cancelled): %w", channel, chatID, err)
        }
        lastErr = fn()
        if lastErr == nil {
            return nil
        }
        zlogger.Warn("send failed, retrying",
            zlogger.String("channel", channel),
            zlogger.String("chat_id", chatID),
            zlogger.Int("attempt", i+1),
            zlogger.Int("max_retries", maxRetries),
            zlogger.Err(lastErr),
        )
        if i < maxRetries-1 {
            select {
            case <-time.After(interval):
            case <-ctx.Done():
                return fmt.Errorf("%s send to %s failed (context cancelled): %w", channel, chatID, ctx.Err())
            }
        }
    }
    return fmt.Errorf("%s send to %s failed (retried %d times): %w", channel, chatID, maxRetries, lastErr)
}
