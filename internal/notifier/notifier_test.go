package notifier

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/http/httptest"
    "os"
    "sync"
    "sync/atomic"
    "testing"
    "time"

    "github.com/vincent119/zlogger"
)

func TestMain(m *testing.M) {
    zlogger.Init(nil)
    os.Exit(m.Run())
}

// TestTelegramNotifier_SendSuccess 驗證 mock server 回傳 200 時，request body 正確
func TestTelegramNotifier_SendSuccess(t *testing.T) {
    var received telegramMessage
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
            t.Errorf("解析 request body 失敗: %v", err)
        }
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, `{"ok":true}`)
    }))
    defer server.Close()

    n := NewTelegramNotifier("test-token", server.Client(), 3, 10*time.Millisecond)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "hello", []string{"123"})
    if err != nil {
        t.Fatalf("預期成功，但收到錯誤: %v", err)
    }

    if received.ChatID != "123" {
        t.Errorf("chat_id 預期 123，實際 %s", received.ChatID)
    }
    if received.Text != "hello" {
        t.Errorf("text 預期 hello，實際 %s", received.Text)
    }
    if received.ParseMode != "HTML" {
        t.Errorf("parse_mode 預期 HTML，實際 %s", received.ParseMode)
    }
}

// TestTelegramNotifier_SendMultipleChatIDs 驗證多個 chatID 都收到訊息
func TestTelegramNotifier_SendMultipleChatIDs(t *testing.T) {
    var mu sync.Mutex
    receivedIDs := make(map[string]bool)

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var msg telegramMessage
        if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
            t.Errorf("解析 request body 失敗: %v", err)
        }
        mu.Lock()
        receivedIDs[msg.ChatID] = true
        mu.Unlock()
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, `{"ok":true}`)
    }))
    defer server.Close()

    n := NewTelegramNotifier("test-token", server.Client(), 3, 10*time.Millisecond)
    n.baseURL = server.URL

    chatIDs := []string{"111", "222", "333"}
    err := n.Send(context.Background(), "broadcast", chatIDs)
    if err != nil {
        t.Fatalf("預期成功，但收到錯誤: %v", err)
    }

    for _, id := range chatIDs {
        if !receivedIDs[id] {
            t.Errorf("chatID %s 未收到訊息", id)
        }
    }
}

// TestTelegramNotifier_RetrySuccess 驗證前 2 次失敗、第 3 次成功的重試邏輯
func TestTelegramNotifier_RetrySuccess(t *testing.T) {
    var count atomic.Int32

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        c := count.Add(1)
        if c < 3 {
            w.WriteHeader(http.StatusInternalServerError)
            fmt.Fprint(w, `{"ok":false}`)
            return
        }
        w.WriteHeader(http.StatusOK)
        fmt.Fprint(w, `{"ok":true}`)
    }))
    defer server.Close()

    n := NewTelegramNotifier("test-token", server.Client(), 3, 10*time.Millisecond)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "retry-test", []string{"456"})
    if err != nil {
        t.Fatalf("預期第 3 次成功，但收到錯誤: %v", err)
    }

    if count.Load() != 3 {
        t.Errorf("預期呼叫 3 次，實際 %d 次", count.Load())
    }
}

// TestTelegramNotifier_RetryExhausted 驗證重試耗盡後回傳 error
func TestTelegramNotifier_RetryExhausted(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
        fmt.Fprint(w, `{"ok":false}`)
    }))
    defer server.Close()

    n := NewTelegramNotifier("test-token", server.Client(), 3, 10*time.Millisecond)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "fail-test", []string{"789"})
    if err == nil {
        t.Fatal("預期收到錯誤，但回傳 nil")
    }
}

// TestSlackNotifier_SendSuccess 驗證 mock server 回傳 200 時，header 和 body 正確
func TestSlackNotifier_SendSuccess(t *testing.T) {
    var received slackMessage
    var authHeader string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader = r.Header.Get("Authorization")
        if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
            t.Errorf("解析 request body 失敗: %v", err)
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    n := NewSlackNotifier("test-token", server.Client(), 3, 10*time.Millisecond)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "hello slack", []string{"C12345"})
    if err != nil {
        t.Fatalf("預期成功，但收到錯誤: %v", err)
    }

    if authHeader != "Bearer test-token" {
        t.Errorf("Authorization 預期 'Bearer test-token'，實際 '%s'", authHeader)
    }
    if received.Channel != "C12345" {
        t.Errorf("channel 預期 C12345，實際 %s", received.Channel)
    }
    if received.Text != "hello slack" {
        t.Errorf("text 預期 'hello slack'，實際 '%s'", received.Text)
    }
}

// TestSlackNotifier_RetryExhausted 驗證重試耗盡後回傳 error
func TestSlackNotifier_RetryExhausted(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer server.Close()

    n := NewSlackNotifier("test-token", server.Client(), 3, 10*time.Millisecond)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "fail", []string{"C99999"})
    if err == nil {
        t.Fatal("預期收到錯誤，但回傳 nil")
    }
}

// TestDiscordNotifier_SendSuccess 驗證 mock server 回傳 200 時，header、URL 和 body 正確
func TestDiscordNotifier_SendSuccess(t *testing.T) {
    var received discordMessage
    var authHeader string
    var requestPath string
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader = r.Header.Get("Authorization")
        requestPath = r.URL.Path
        if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
            t.Errorf("解析 request body 失敗: %v", err)
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    n := NewDiscordNotifier("test-token", server.Client(), 3, 10*time.Millisecond, 2000)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "hello discord", []string{"ch-001"})
    if err != nil {
        t.Fatalf("預期成功，但收到錯誤: %v", err)
    }

    if authHeader != "Bot test-token" {
        t.Errorf("Authorization 預期 'Bot test-token'，實際 '%s'", authHeader)
    }
    if requestPath != "/channels/ch-001/messages" {
        t.Errorf("URL path 預期 '/channels/ch-001/messages'，實際 '%s'", requestPath)
    }
    if received.Content != "hello discord" {
        t.Errorf("content 預期 'hello discord'，實際 '%s'", received.Content)
    }
}

// TestDiscordNotifier_RetryExhausted 驗證重試耗盡後回傳 error
func TestDiscordNotifier_RetryExhausted(t *testing.T) {
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusInternalServerError)
    }))
    defer server.Close()

    n := NewDiscordNotifier("test-token", server.Client(), 3, 10*time.Millisecond, 2000)
    n.baseURL = server.URL

    err := n.Send(context.Background(), "fail", []string{"ch-999"})
    if err == nil {
        t.Fatal("預期收到錯誤，但回傳 nil")
    }
}

// TestDiscordNotifier_MessageTruncation 驗證超過 maxLength 的訊息會被截斷
func TestDiscordNotifier_MessageTruncation(t *testing.T) {
    var received discordMessage
    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
            t.Errorf("解析 request body 失敗: %v", err)
        }
        w.WriteHeader(http.StatusOK)
    }))
    defer server.Close()

    maxLen := 100
    n := NewDiscordNotifier("test-token", server.Client(), 3, 10*time.Millisecond, maxLen)
    n.baseURL = server.URL

    // 建立超過 maxLength 的訊息
    longMsg := ""
    for i := 0; i < 150; i++ {
        longMsg += "A"
    }

    err := n.Send(context.Background(), longMsg, []string{"ch-trunc"})
    if err != nil {
        t.Fatalf("預期成功，但收到錯誤: %v", err)
    }

    if len(received.Content) > maxLen {
        t.Errorf("截斷後長度 %d 超過 maxLength %d", len(received.Content), maxLen)
    }
    if received.Content[len(received.Content)-len("...(truncated)"):] != "...(truncated)" {
        t.Errorf("截斷後訊息應以 '...(truncated)' 結尾，實際: '%s'", received.Content)
    }
}
