package message

import (
    "bytes"
    "encoding/json"
    "strings"

    "github.com/vincent119/audit-notifier/internal/event"
)

// prettyJSON 將壓縮的 JSON 字串格式化為縮排形式，失敗時回傳原始字串
func prettyJSON(raw string) string {
    var buf bytes.Buffer
    if err := json.Indent(&buf, []byte(raw), "  ", "  "); err != nil {
        return raw
    }
    return "\n  " + buf.String()
}

// FormatMessage 將 AuditEvent 格式化為多語言通知訊息
func FormatMessage(e *event.AuditEvent, lang string) string {
    var b strings.Builder

    // 標題行
    b.WriteString("[")
    b.WriteString(GetTranslation(lang, KeyTitle))
    b.WriteString("]\n")

    isConsoleLogin := e.EventName == "ConsoleLogin"

    // 依序輸出有值的欄位
    fields := []struct {
        key   string
        value string
    }{
        {KeyEventName, e.EventName},
        {KeyEventSource, e.EventSource},
        {KeyEventTime, e.EventTime},
        {KeyRegion, e.AWSRegion},
        {KeySourceIP, e.SourceIPAddress},
        {KeyUserIdentity, e.UserIdentity},
    }

    // ConsoleLogin：插入登入結果和 MFA
    if isConsoleLogin {
        fields = append(fields,
            struct{ key, value string }{KeyLoginResult, e.LoginResult},
            struct{ key, value string }{KeyMFAUsed, e.MFAUsed},
        )
    }

    // 通用欄位
    fields = append(fields,
        struct{ key, value string }{KeyResources, strings.Join(e.Resources, ", ")},
        struct{ key, value string }{KeyErrorCode, e.ErrorCode},
        struct{ key, value string }{KeyErrorMessage, e.ErrorMessage},
        struct{ key, value string }{KeyRequestParameters, e.RequestParameters},
    )

    // ConsoleLogin 不顯示 responseElements（內容已提取到 LoginResult）
    if !isConsoleLogin {
        fields = append(fields, struct{ key, value string }{KeyResponseElements, e.ResponseElements})
    }

    jsonFields := map[string]bool{
        KeyRequestParameters: true,
        KeyResponseElements:  true,
    }

    for _, f := range fields {
        if f.value == "" {
            continue
        }
        b.WriteString(GetTranslation(lang, f.key))
        b.WriteString(": ")
        if jsonFields[f.key] {
            b.WriteString(prettyJSON(f.value))
        } else {
            b.WriteString(f.value)
        }
        b.WriteString("\n")
    }

    return b.String()
}
