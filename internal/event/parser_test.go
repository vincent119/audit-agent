package event

import (
    "testing"
)

// buildPayload 組合 EventBridge envelope JSON
func buildPayload(detail string) []byte {
    return []byte(`{
    "source": "aws.cloudtrail",
    "detail-type": "AWS API Call via CloudTrail",
    "detail": ` + detail + `
}`)
}

func TestParseEvent_FullEvent(t *testing.T) {
    detail := `{
        "eventName": "CreateUser",
        "eventSource": "iam.amazonaws.com",
        "eventTime": "2026-04-16T07:00:00Z",
        "awsRegion": "us-east-1",
        "sourceIPAddress": "1.2.3.4",
        "userIdentity": {
            "type": "IAMUser",
            "arn": "arn:aws:iam::123456789012:user/admin",
            "userName": "admin"
        },
        "resources": [
            {"ARN": "arn:aws:iam::123456789012:user/new-user"}
        ],
        "errorCode": "",
        "errorMessage": "",
        "requestParameters": {"userName": "new-user", "path": "/"},
        "responseElements": {"user": {"userName": "new-user", "userId": "AIDA12345"}}
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }

    if ev.EventName != "CreateUser" {
        t.Errorf("EventName = %q, 預期 %q", ev.EventName, "CreateUser")
    }
    if ev.EventSource != "iam.amazonaws.com" {
        t.Errorf("EventSource = %q, 預期 %q", ev.EventSource, "iam.amazonaws.com")
    }
    if ev.AWSRegion != "us-east-1" {
        t.Errorf("AWSRegion = %q, 預期 %q", ev.AWSRegion, "us-east-1")
    }
    if ev.SourceIPAddress != "1.2.3.4" {
        t.Errorf("SourceIPAddress = %q, 預期 %q", ev.SourceIPAddress, "1.2.3.4")
    }
    if ev.UserIdentity != "arn:aws:iam::123456789012:user/admin" {
        t.Errorf("UserIdentity = %q, 預期 ARN 值", ev.UserIdentity)
    }
    if len(ev.Resources) != 1 || ev.Resources[0] != "arn:aws:iam::123456789012:user/new-user" {
        t.Errorf("Resources = %v, 預期包含一個 ARN", ev.Resources)
    }
    if ev.ErrorCode != "" {
        t.Errorf("ErrorCode = %q, 預期空字串", ev.ErrorCode)
    }
    if ev.RequestParameters == "" {
        t.Error("RequestParameters 不應為空")
    }
    if ev.ResponseElements == "" {
        t.Error("ResponseElements 不應為空")
    }
}

func TestParseEvent_MultiRegion(t *testing.T) {
    detail := `{
        "eventName": "DescribeInstances",
        "eventSource": "ec2.amazonaws.com",
        "awsRegion": "ap-east-1",
        "userIdentity": {"type": "IAMUser", "arn": "arn:aws:iam::123456789012:user/dev"}
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.AWSRegion != "ap-east-1" {
        t.Errorf("AWSRegion = %q, 預期 %q", ev.AWSRegion, "ap-east-1")
    }
}

func TestParseEvent_WithError(t *testing.T) {
    detail := `{
        "eventName": "DeleteBucket",
        "eventSource": "s3.amazonaws.com",
        "awsRegion": "us-west-2",
        "userIdentity": {"type": "IAMUser", "arn": "arn:aws:iam::123456789012:user/dev"},
        "errorCode": "AccessDenied",
        "errorMessage": "User is not authorized to perform this action"
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.ErrorCode != "AccessDenied" {
        t.Errorf("ErrorCode = %q, 預期 %q", ev.ErrorCode, "AccessDenied")
    }
    if ev.ErrorMessage != "User is not authorized to perform this action" {
        t.Errorf("ErrorMessage = %q, 預期包含拒絕訊息", ev.ErrorMessage)
    }
}

func TestParseEvent_MissingFields(t *testing.T) {
    // 最小化事件：只有 eventName
    detail := `{"eventName": "GetObject"}`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.EventName != "GetObject" {
        t.Errorf("EventName = %q, 預期 %q", ev.EventName, "GetObject")
    }
    if ev.EventSource != "" {
        t.Errorf("EventSource = %q, 預期空字串", ev.EventSource)
    }
    if ev.UserIdentity != "" {
        t.Errorf("UserIdentity = %q, 預期空字串", ev.UserIdentity)
    }
    if len(ev.Resources) != 0 {
        t.Errorf("Resources = %v, 預期空切片", ev.Resources)
    }
    if ev.RequestParameters != "" {
        t.Errorf("RequestParameters = %q, 預期空字串", ev.RequestParameters)
    }
    if ev.ResponseElements != "" {
        t.Errorf("ResponseElements = %q, 預期空字串", ev.ResponseElements)
    }
}

func TestParseEvent_NullRequestParameters(t *testing.T) {
    detail := `{
        "eventName": "DescribeInstances",
        "requestParameters": null,
        "responseElements": null
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.RequestParameters != "" {
        t.Errorf("RequestParameters = %q, null 應轉為空字串", ev.RequestParameters)
    }
    if ev.ResponseElements != "" {
        t.Errorf("ResponseElements = %q, null 應轉為空字串", ev.ResponseElements)
    }
}

func TestParseEvent_InvalidJSON(t *testing.T) {
    _, err := ParseEvent([]byte(`not json`))
    if err == nil {
        t.Fatal("預期回傳錯誤，但得到 nil")
    }
}

func TestParseEvent_UserIdentityFallback(t *testing.T) {
    // 沒有 arn，應 fallback 到 userName
    detail := `{
        "eventName": "PutObject",
        "userIdentity": {"type": "IAMUser", "userName": "deploy-bot"}
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.UserIdentity != "deploy-bot" {
        t.Errorf("UserIdentity = %q, 預期 fallback 到 userName %q", ev.UserIdentity, "deploy-bot")
    }
}

func TestParseEvent_ConsoleLogin(t *testing.T) {
    detail := `{
        "eventName": "ConsoleLogin",
        "eventSource": "signin.amazonaws.com",
        "eventTime": "2026-04-17T08:10:54Z",
        "awsRegion": "us-east-1",
        "sourceIPAddress": "61.222.155.230",
        "userIdentity": {
            "type": "IAMUser",
            "arn": "arn:aws:iam::123456789012:user/vincent",
            "userName": "vincent"
        },
        "requestParameters": null,
        "responseElements": {"ConsoleLogin": "Success"},
        "additionalEventData": {"MFAUsed": "Yes", "MobileVersion": "No"}
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.UserIdentity != "vincent" {
        t.Errorf("UserIdentity = %q, ConsoleLogin 應優先用 userName", ev.UserIdentity)
    }
    if ev.LoginResult != "Success" {
        t.Errorf("LoginResult = %q, 預期 %q", ev.LoginResult, "Success")
    }
    if ev.MFAUsed != "Yes" {
        t.Errorf("MFAUsed = %q, 預期 %q", ev.MFAUsed, "Yes")
    }
}

func TestParseEvent_ConsoleLogin_NoMFA(t *testing.T) {
    detail := `{
        "eventName": "ConsoleLogin",
        "eventSource": "signin.amazonaws.com",
        "awsRegion": "us-east-1",
        "userIdentity": {
            "type": "IAMUser",
            "arn": "arn:aws:iam::123456789012:user/test",
            "userName": "test"
        },
        "responseElements": {"ConsoleLogin": "Failure"},
        "additionalEventData": {"MFAUsed": "No"}
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    if ev.LoginResult != "Failure" {
        t.Errorf("LoginResult = %q, 預期 %q", ev.LoginResult, "Failure")
    }
    if ev.MFAUsed != "No" {
        t.Errorf("MFAUsed = %q, 預期 %q", ev.MFAUsed, "No")
    }
}

func TestParseEvent_NonConsoleLogin_NoSpecialFields(t *testing.T) {
    detail := `{
        "eventName": "CreateUser",
        "userIdentity": {
            "type": "IAMUser",
            "arn": "arn:aws:iam::123456789012:user/admin",
            "userName": "admin"
        },
        "additionalEventData": {"MFAUsed": "Yes"}
    }`
    payload := buildPayload(detail)

    ev, err := ParseEvent(payload)
    if err != nil {
        t.Fatalf("預期無錯誤，但得到: %v", err)
    }
    // 非 ConsoleLogin 不應解析 MFAUsed/LoginResult
    if ev.MFAUsed != "" {
        t.Errorf("MFAUsed = %q, 非 ConsoleLogin 應為空", ev.MFAUsed)
    }
    if ev.LoginResult != "" {
        t.Errorf("LoginResult = %q, 非 ConsoleLogin 應為空", ev.LoginResult)
    }
    // 非 ConsoleLogin 應用 ARN
    if ev.UserIdentity != "arn:aws:iam::123456789012:user/admin" {
        t.Errorf("UserIdentity = %q, 非 ConsoleLogin 應用 ARN", ev.UserIdentity)
    }
}