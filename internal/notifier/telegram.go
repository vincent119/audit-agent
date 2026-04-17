package notifier

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "net/http"
    "time"

    "github.com/vincent119/zlogger"
)

// TelegramNotifier 透過 Telegram Bot API 發送通知
type TelegramNotifier struct {
    token         string
    client        *http.Client
    baseURL       string // 預設 https://api.telegram.org，可覆寫用於測試
    maxRetries    int
    retryInterval time.Duration
}

// telegramMessage 代表發送給 Telegram API 的訊息結構
type telegramMessage struct {
    ChatID    string `json:"chat_id"`
    Text      string `json:"text"`
    ParseMode string `json:"parse_mode"`
}

// NewTelegramNotifier 建立 TelegramNotifier
func NewTelegramNotifier(token string, client *http.Client, maxRetries int, retryInterval time.Duration) *TelegramNotifier {
    return &TelegramNotifier{
        token:         token,
        client:        client,
        baseURL:       "https://api.telegram.org",
        maxRetries:    maxRetries,
        retryInterval: retryInterval,
    }
}

// Send 對每個 chatID 獨立發送訊息，收集所有錯誤後合併回傳
func (t *TelegramNotifier) Send(ctx context.Context, message string, chatIDs []string) error {
    var errs []error
    for _, chatID := range chatIDs {
        err := sendWithRetry(ctx, "telegram", chatID, t.maxRetries, t.retryInterval, func() error {
            return t.sendMessage(ctx, chatID, message)
        })
        if err != nil {
            errs = append(errs, err)
        } else {
            zlogger.Info("Telegram message sent",
                zlogger.String("chat_id", chatID),
            )
        }
    }
    return errors.Join(errs...)
}

// sendMessage 執行單次 HTTP POST 到 Telegram Bot API
func (t *TelegramNotifier) sendMessage(ctx context.Context, chatID, text string) error {
    url := fmt.Sprintf("%s/bot%s/sendMessage", t.baseURL, t.token)

    body, err := json.Marshal(telegramMessage{
        ChatID:    chatID,
        Text:      text,
        ParseMode: "HTML",
    })
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := t.client.Do(req)
    if err != nil {
        return fmt.Errorf("HTTP 請求失敗: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("Telegram API 回傳非 2xx 狀態碼: %d, body: %s", resp.StatusCode, string(respBody))
    }

    return nil
}
