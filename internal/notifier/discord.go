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

// DiscordNotifier 透過 Discord Bot API 發送通知
type DiscordNotifier struct {
    token         string
    client        *http.Client
    baseURL       string // 預設 https://discord.com/api/v10，可覆寫用於測試
    maxRetries    int
    retryInterval time.Duration
    maxLength     int // 訊息最大字元數，預設 2000
}

// discordMessage 代表發送給 Discord API 的訊息結構
type discordMessage struct {
    Content string `json:"content"`
}

// NewDiscordNotifier 建立 DiscordNotifier
func NewDiscordNotifier(token string, client *http.Client, maxRetries int, retryInterval time.Duration, maxLength int) *DiscordNotifier {
    if maxLength <= 0 {
        maxLength = 2000
    }
    return &DiscordNotifier{
        token:         token,
        client:        client,
        baseURL:       "https://discord.com/api/v10",
        maxRetries:    maxRetries,
        retryInterval: retryInterval,
        maxLength:     maxLength,
    }
}

// Send 對每個 channelID 獨立發送訊息，收集所有錯誤後合併回傳
func (d *DiscordNotifier) Send(ctx context.Context, message string, channelIDs []string) error {
    // 截斷邏輯：超過 maxLength 時截斷並加上後綴
    msg := message
    if len(msg) > d.maxLength {
        originalLen := len(msg)
        msg = msg[:d.maxLength-15] + "\n...(truncated)"
        zlogger.Warn("Discord message exceeded length limit, truncated",
            zlogger.Int("original_length", originalLen),
            zlogger.Int("truncated_length", len(msg)),
        )
    }

    var errs []error
    for _, channelID := range channelIDs {
        err := sendWithRetry(ctx, "discord", channelID, d.maxRetries, d.retryInterval, func() error {
            return d.sendMessage(ctx, channelID, msg)
        })
        if err != nil {
            errs = append(errs, err)
        } else {
            zlogger.Info("Discord message sent",
                zlogger.String("channel_id", channelID),
            )
        }
    }
    return errors.Join(errs...)
}

// sendMessage 執行單次 HTTP POST 到 Discord Bot API
func (d *DiscordNotifier) sendMessage(ctx context.Context, channelID, content string) error {
    url := fmt.Sprintf("%s/channels/%s/messages", d.baseURL, channelID)

    body, err := json.Marshal(discordMessage{
        Content: content,
    })
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bot %s", d.token))
    req.Header.Set("Content-Type", "application/json")

    resp, err := d.client.Do(req)
    if err != nil {
        return fmt.Errorf("HTTP 請求失敗: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("Discord API 回傳非 2xx 狀態碼: %d, body: %s", resp.StatusCode, string(respBody))
    }

    return nil
}
