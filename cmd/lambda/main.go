package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"

    "github.com/aws/aws-lambda-go/lambda"
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/ssm"
    "github.com/vincent119/audit-notifier/internal/crypto"
    "github.com/vincent119/audit-notifier/internal/event"
    "github.com/vincent119/audit-notifier/internal/message"
    "github.com/vincent119/audit-notifier/internal/notifier"
    "github.com/vincent119/zlogger"
    "golang.org/x/sync/errgroup"
)

// 全域變數（init phase 初始化，warm invocation 複用）
var (
    notifiers []channelNotifier
    msgLang   string
)

// channelNotifier 封裝單一通知頻道的 Notifier 與目標 chat IDs
type channelNotifier struct {
    name    string
    n       notifier.Notifier
    chatIDs []string
}

// handler 處理 Lambda 事件：解析、格式化、並行發送通知
func handler(ctx context.Context, rawEvent json.RawMessage) error {
    // 解析事件
    auditEvent, err := event.ParseEvent(rawEvent)
    if err != nil {
        zlogger.Error("failed to parse event", zlogger.Err(err))
        return fmt.Errorf("failed to parse event: %w", err)
    }

    // 格式化訊息
    msg := message.FormatMessage(auditEvent, msgLang)
    zlogger.Info("event parsed",
        zlogger.String("event_name", auditEvent.EventName),
        zlogger.String("event_source", auditEvent.EventSource),
        zlogger.String("region", auditEvent.AWSRegion),
    )

    // 並行發送到各頻道
    g, gctx := errgroup.WithContext(ctx)
    var failedChannels []string
    var mu sync.Mutex

    for _, cn := range notifiers {
        cn := cn // capture
        g.Go(func() error {
            if err := cn.n.Send(gctx, msg, cn.chatIDs); err != nil {
                mu.Lock()
                failedChannels = append(failedChannels, cn.name)
                mu.Unlock()
                zlogger.Error("channel send failed",
                    zlogger.String("channel", cn.name),
                    zlogger.Err(err),
                )
                return err
            }
            return nil
        })
    }

    err = g.Wait()

    // 部分成功時不回傳 error
    if err != nil {
        if len(failedChannels) < len(notifiers) {
            // 部分成功，log 失敗的頻道但不回傳 error
            zlogger.Error("partial channels failed",
                zlogger.String("failed_channels", strings.Join(failedChannels, ",")),
            )
            return nil
        }
        // 全部失敗，回傳 error 觸發 EventBridge 重試
        zlogger.Error("all channels failed",
            zlogger.String("failed_channels", strings.Join(failedChannels, ",")),
        )
        return fmt.Errorf("all channels failed: %s", strings.Join(failedChannels, ", "))
    }

    zlogger.Info("all channels sent successfully")
    return nil
}

func main() {
    // 初始化 zlogger
    logLevel := getEnv("LOG_LEVEL", "info")
    zlogger.Init(&zlogger.Config{
        Level:   logLevel,
        Format:  "json",
        Outputs: []string{"console"},
    })
    defer zlogger.Sync()

    ctx := context.Background()

    // 從 SSM 讀取 ENCRYPTION_KEY
    ssmKeyPath := os.Getenv("SSM_KEY_PATH")
    if ssmKeyPath == "" {
        zlogger.Fatal("missing required env SSM_KEY_PATH")
    }

    awsCfg, err := config.LoadDefaultConfig(ctx)
    if err != nil {
        zlogger.Fatal("failed to load AWS config", zlogger.Err(err))
    }

    ssmClient := ssm.NewFromConfig(awsCfg)
    paramOut, err := ssmClient.GetParameter(ctx, &ssm.GetParameterInput{
        Name:           aws.String(ssmKeyPath),
        WithDecryption: aws.Bool(true),
    })
    if err != nil {
        zlogger.Fatal("failed to get ENCRYPTION_KEY from SSM", zlogger.Err(err))
    }
    encryptionKey := aws.ToString(paramOut.Parameter.Value)

    // 讀取通知相關環境變數
    channels := getEnv("NOTIFY_CHANNELS", "")
    if channels == "" {
        zlogger.Fatal("missing required env NOTIFY_CHANNELS")
    }

    httpTimeout := time.Duration(getEnvInt("HTTP_TIMEOUT", 10)) * time.Second
    httpClient := &http.Client{Timeout: httpTimeout}
    msgLang = getEnv("MSG_LANG", "en")
    maxLength := getEnvInt("MESSAGE_MAX_LENGTH", 2000)

    // 初始化各頻道 notifier
    for _, ch := range strings.Split(channels, ",") {
        ch = strings.TrimSpace(ch)
        switch ch {
        case "telegram":
            token, ids := initChannel(encryptionKey, "TG_TOKEN", "TG_CHAT_IDS")
            n := notifier.NewTelegramNotifier(token, httpClient, 3, 10*time.Second)
            notifiers = append(notifiers, channelNotifier{name: "telegram", n: n, chatIDs: ids})
        case "slack":
            token, ids := initChannel(encryptionKey, "SLACK_TOKEN", "SLACK_CHAT_IDS")
            n := notifier.NewSlackNotifier(token, httpClient, 3, 10*time.Second)
            notifiers = append(notifiers, channelNotifier{name: "slack", n: n, chatIDs: ids})
        case "discord":
            token, ids := initChannel(encryptionKey, "DISCORD_TOKEN", "DISCORD_CHAT_IDS")
            n := notifier.NewDiscordNotifier(token, httpClient, 3, 10*time.Second, maxLength)
            notifiers = append(notifiers, channelNotifier{name: "discord", n: n, chatIDs: ids})
        default:
            zlogger.Warn("unsupported channel, skipped", zlogger.String("channel", ch))
        }
    }

    if len(notifiers) == 0 {
        zlogger.Fatal("no notifier initialized")
    }

    zlogger.Info("Lambda 初始化完成",
        zlogger.Int("notifier_count", len(notifiers)),
        zlogger.String("channels", channels),
    )

    lambda.Start(handler)
}

// initChannel 解密 token 並解析 chat IDs，失敗時直接 Fatal
func initChannel(encryptionKey, tokenEnv, chatIDsEnv string) (string, []string) {
    encryptedToken := os.Getenv(tokenEnv)
    if encryptedToken == "" {
        zlogger.Fatal("missing required env", zlogger.String("env", tokenEnv))
    }

    token, err := crypto.Decrypt(encryptionKey, encryptedToken)
    if err != nil {
        zlogger.Fatal("failed to decrypt token",
            zlogger.String("env", tokenEnv),
            zlogger.Err(err),
        )
    }

    chatIDsRaw := os.Getenv(chatIDsEnv)
    if chatIDsRaw == "" {
        zlogger.Fatal("missing required env", zlogger.String("env", chatIDsEnv))
    }

    var chatIDs []string
    for _, id := range strings.Split(chatIDsRaw, ",") {
        id = strings.TrimSpace(id)
        if id != "" {
            chatIDs = append(chatIDs, id)
        }
    }

    if len(chatIDs) == 0 {
        zlogger.Fatal("chat IDs 為空", zlogger.String("env", chatIDsEnv))
    }

    return token, chatIDs
}

// getEnv 讀取環境變數，若為空則回傳預設值
func getEnv(key, defaultVal string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return defaultVal
}

// getEnvInt 讀取環境變數並轉為 int，失敗或為空時回傳預設值
func getEnvInt(key string, defaultVal int) int {
    v := os.Getenv(key)
    if v == "" {
        return defaultVal
    }
    i, err := strconv.Atoi(v)
    if err != nil {
        return defaultVal
    }
    return i
}
