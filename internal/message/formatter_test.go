package message

import (
    "strings"
    "testing"

    "github.com/vincent119/audit-notifier/internal/event"
)

// 完整事件測試資料
func fullEvent() *event.AuditEvent {
    return &event.AuditEvent{
        EventName:         "CreateUser",
        EventSource:       "iam.amazonaws.com",
        EventTime:         "2024-01-01T00:00:00Z",
        AWSRegion:         "us-east-1",
        SourceIPAddress:   "1.2.3.4",
        UserIdentity:      "arn:aws:iam::123456:user/admin",
        Resources:         []string{"arn:aws:iam::123456:user/newuser"},
        ErrorCode:         "",
        ErrorMessage:      "",
        RequestParameters: `{"userName":"new-user","path":"/"}`,
        ResponseElements:  `{"user":{"userName":"new-user"}}`,
    }
}

func TestFormatMessage_English_FullEvent(t *testing.T) {
    result := FormatMessage(fullEvent(), "en")

    expects := []string{
        "[AWS CloudTrail Audit Alert]",
        "Event Name: CreateUser",
        "Event Source: iam.amazonaws.com",
        "Event Time: 2024-01-01T00:00:00Z",
        "Region: us-east-1",
        "Source IP: 1.2.3.4",
        "User: arn:aws:iam::123456:user/admin",
        "Resources: arn:aws:iam::123456:user/newuser",
        `Request Parameters: 
  {
    "userName": "new-user",
    "path": "/"
  }`,
        `Response Elements: 
  {
    "user": {
      "userName": "new-user"
    }
  }`,
    }
    for _, exp := range expects {
        if !strings.Contains(result, exp) {
            t.Errorf("期望包含 %q，實際輸出:\n%s", exp, result)
        }
    }
    // 無錯誤欄位不應出現
    if strings.Contains(result, "Error Code") {
        t.Error("不應包含 Error Code 欄位")
    }
}

func TestFormatMessage_ZhTW_FullEvent(t *testing.T) {
    result := FormatMessage(fullEvent(), "zh-TW")

    expects := []string{
        "[AWS CloudTrail 審計警報]",
        "事件名稱: CreateUser",
        "事件來源: iam.amazonaws.com",
        "事件時間: 2024-01-01T00:00:00Z",
        "區域: us-east-1",
        "來源 IP: 1.2.3.4",
        "操作者: arn:aws:iam::123456:user/admin",
        "影響資源: arn:aws:iam::123456:user/newuser",
        "請求參數: \n  {\n",
        "回應內容: \n  {\n",
    }
    for _, exp := range expects {
        if !strings.Contains(result, exp) {
            t.Errorf("期望包含 %q，實際輸出:\n%s", exp, result)
        }
    }
}

func TestFormatMessage_ZhCN_FullEvent(t *testing.T) {
    result := FormatMessage(fullEvent(), "zh-CN")

    expects := []string{
        "[AWS CloudTrail 审计警报]",
        "事件名称: CreateUser",
        "事件来源: iam.amazonaws.com",
        "事件时间: 2024-01-01T00:00:00Z",
        "区域: us-east-1",
        "来源 IP: 1.2.3.4",
        "操作者: arn:aws:iam::123456:user/admin",
        "影响资源: arn:aws:iam::123456:user/newuser",
        "请求参数: \n  {\n",
        "响应内容: \n  {\n",
    }
    for _, exp := range expects {
        if !strings.Contains(result, exp) {
            t.Errorf("期望包含 %q，實際輸出:\n%s", exp, result)
        }
    }
}

func TestFormatMessage_MissingFields(t *testing.T) {
    e := &event.AuditEvent{
        EventName: "StopInstances",
        EventTime: "2024-06-15T12:00:00Z",
    }
    result := FormatMessage(e, "en")

    // 應包含有值欄位
    if !strings.Contains(result, "Event Name: StopInstances") {
        t.Errorf("期望包含 Event Name，實際輸出:\n%s", result)
    }
    if !strings.Contains(result, "Event Time: 2024-06-15T12:00:00Z") {
        t.Errorf("期望包含 Event Time，實際輸出:\n%s", result)
    }
    // 不應包含空值欄位
    for _, absent := range []string{"Event Source:", "Region:", "Source IP:", "User:", "Resources:"} {
        if strings.Contains(result, absent) {
            t.Errorf("不應包含 %q，實際輸出:\n%s", absent, result)
        }
    }
}

func TestFormatMessage_WithError(t *testing.T) {
    e := fullEvent()
    e.ErrorCode = "AccessDenied"
    e.ErrorMessage = "User is not authorized"
    result := FormatMessage(e, "en")

    if !strings.Contains(result, "Error Code: AccessDenied") {
        t.Errorf("期望包含 Error Code，實際輸出:\n%s", result)
    }
    if !strings.Contains(result, "Error Message: User is not authorized") {
        t.Errorf("期望包含 Error Message，實際輸出:\n%s", result)
    }
}

func TestFormatMessage_UnknownLang(t *testing.T) {
    result := FormatMessage(fullEvent(), "ja")

    // 未知語言應 fallback 到英文
    if !strings.Contains(result, "[AWS CloudTrail Audit Alert]") {
        t.Errorf("未知語言應 fallback 到英文，實際輸出:\n%s", result)
    }
    if !strings.Contains(result, "Event Name: CreateUser") {
        t.Errorf("未知語言應 fallback 到英文標籤，實際輸出:\n%s", result)
    }
}

func TestGetTranslation_Fallback(t *testing.T) {
    // 已知語言已知 key
    if v := GetTranslation("zh-TW", KeyTitle); v != "AWS CloudTrail 審計警報" {
        t.Errorf("zh-TW title 期望 'AWS CloudTrail 審計警報'，得到 %q", v)
    }
    // 未知語言 fallback 到 en
    if v := GetTranslation("fr", KeyTitle); v != "AWS CloudTrail Audit Alert" {
        t.Errorf("未知語言 fallback 期望英文標題，得到 %q", v)
    }
    // 未知 key 回傳 key 本身
    if v := GetTranslation("en", "unknown_key"); v != "unknown_key" {
        t.Errorf("未知 key 期望回傳 key 本身，得到 %q", v)
    }
}

func TestFormatMessage_ConsoleLogin_ZhCN(t *testing.T) {
    e := &event.AuditEvent{
        EventName:        "ConsoleLogin",
        EventSource:      "signin.amazonaws.com",
        EventTime:        "2026-04-17T08:10:54Z",
        AWSRegion:        "us-east-1",
        SourceIPAddress:  "61.222.155.230",
        UserIdentity:     "vincent",
        LoginResult:      "Success",
        MFAUsed:          "Yes",
        ResponseElements: `{"ConsoleLogin":"Success"}`,
    }
    result := FormatMessage(e, "zh-CN")

    expects := []string{
        "登入结果: Success",
        "MFA: Yes",
        "操作者: vincent",
    }
    for _, exp := range expects {
        if !strings.Contains(result, exp) {
            t.Errorf("期望包含 %q，實際輸出:\n%s", exp, result)
        }
    }
    // ConsoleLogin 不應顯示 responseElements
    if strings.Contains(result, "响应内容") {
        t.Errorf("ConsoleLogin 不應顯示 responseElements，實際輸出:\n%s", result)
    }
}

func TestFormatMessage_ConsoleLogin_En(t *testing.T) {
    e := &event.AuditEvent{
        EventName:        "ConsoleLogin",
        EventSource:      "signin.amazonaws.com",
        EventTime:        "2026-04-17T08:10:54Z",
        AWSRegion:        "us-east-1",
        SourceIPAddress:  "61.222.155.230",
        UserIdentity:     "test-user",
        LoginResult:      "Failure",
        MFAUsed:          "No",
        ResponseElements: `{"ConsoleLogin":"Failure"}`,
    }
    result := FormatMessage(e, "en")

    expects := []string{
        "Login Result: Failure",
        "MFA: No",
        "User: test-user",
    }
    for _, exp := range expects {
        if !strings.Contains(result, exp) {
            t.Errorf("期望包含 %q，實際輸出:\n%s", exp, result)
        }
    }
    if strings.Contains(result, "Response Elements") {
        t.Errorf("ConsoleLogin 不應顯示 Response Elements，實際輸出:\n%s", result)
    }
}

func TestFormatMessage_NonConsoleLogin_ShowsResponseElements(t *testing.T) {
    e := &event.AuditEvent{
        EventName:        "CreateUser",
        EventSource:      "iam.amazonaws.com",
        EventTime:        "2026-04-17T08:10:54Z",
        AWSRegion:        "us-east-1",
        UserIdentity:     "arn:aws:iam::123456:user/admin",
        ResponseElements: `{"user":{"userName":"new-user"}}`,
    }
    result := FormatMessage(e, "en")

    if !strings.Contains(result, "Response Elements") {
        t.Errorf("非 ConsoleLogin 應顯示 Response Elements，實際輸出:\n%s", result)
    }
    if strings.Contains(result, "Login Result") {
        t.Errorf("非 ConsoleLogin 不應顯示 Login Result，實際輸出:\n%s", result)
    }
}