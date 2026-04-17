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

// SlackNotifier 透過 Slack API 發送通知
type SlackNotifier struct {
    token         string
    client        *http.Client
    baseURL       string // 預設 https://slack.com/api，可覆寫用於測試
    maxRetries    int
    retryInterval time.Duration
}

// slackMessage 代表發送給 Slack API 的訊息結構
type slackMessage struct {
    Channel string `json:"channel"`
    Text    string `json:"text"`
}

// NewSlackNotifier 建立 SlackNotifier
func NewSlackNotifier(token string, client *http.Client, maxRetries int, retryInterval time.Duration) *SlackNotifier {
    return &SlackNotifier{
        token:         token,
        client:        client,
        baseURL:       "https://slack.com/api",
        maxRetries:    maxRetries,
        retryInterval: retryInterval,
    }
}

// Send 對每個 chatID 獨立發送訊息，收集所有錯誤後合併回傳
func (s *SlackNotifier) Send(ctx context.Context, message string, chatIDs []string) error {
    var errs []error
    for _, chatID := range chatIDs {
        err := sendWithRetry(ctx, "slack", chatID, s.maxRetries, s.retryInterval, func() error {
            return s.sendMessage(ctx, chatID, message)
        })
        if err != nil {
            errs = append(errs, err)
        } else {
            zlogger.Info("Slack message sent",
                zlogger.String("channel", chatID),
            )
        }
    }
    return errors.Join(errs...)
}

// sendMessage 執行單次 HTTP POST 到 Slack API
func (s *SlackNotifier) sendMessage(ctx context.Context, channel, text string) error {
    url := fmt.Sprintf("%s/chat.postMessage", s.baseURL)

    body, err := json.Marshal(slackMessage{
        Channel: channel,
        Text:    text,
    })
    if err != nil {
        return fmt.Errorf("failed to marshal message: %w", err)
    }

    req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.token))
    req.Header.Set("Content-Type", "application/json")

    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("HTTP 請求失敗: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        respBody, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("Slack API 回傳非 2xx 狀態碼: %d, body: %s", resp.StatusCode, string(respBody))
    }

    return nil
}
