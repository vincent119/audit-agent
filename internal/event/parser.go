package event

import "encoding/json"

// AuditEvent 代表從 CloudTrail 提取的審計事件
type AuditEvent struct {
    EventName         string
    EventSource       string
    EventTime         string
    AWSRegion         string
    SourceIPAddress   string
    UserIdentity      string
    Resources         []string
    ErrorCode         string
    ErrorMessage      string
    RequestParameters string
    ResponseElements  string
    MFAUsed           string
    LoginResult       string
}

// eventBridgeEvent 代表 EventBridge 的信封結構
type eventBridgeEvent struct {
    Source     string          `json:"source"`
    DetailType string         `json:"detail-type"`
    Detail     json.RawMessage `json:"detail"`
}

// cloudTrailDetail 代表 CloudTrail 事件的詳細內容
type cloudTrailDetail struct {
    EventName           string          `json:"eventName"`
    EventSource         string          `json:"eventSource"`
    EventTime           string          `json:"eventTime"`
    AWSRegion           string          `json:"awsRegion"`
    SourceIPAddress     string          `json:"sourceIPAddress"`
    UserIdentity        userIdentity    `json:"userIdentity"`
    Resources           []resource      `json:"resources"`
    ErrorCode           string          `json:"errorCode"`
    ErrorMessage        string          `json:"errorMessage"`
    RequestParameters   json.RawMessage `json:"requestParameters"`
    ResponseElements    json.RawMessage `json:"responseElements"`
    AdditionalEventData json.RawMessage `json:"additionalEventData"`
}

type userIdentity struct {
    Type     string `json:"type"`
    ARN      string `json:"arn"`
    UserName string `json:"userName"`
}

type resource struct {
    ARN string `json:"ARN"`
}

// rawToString 將 json.RawMessage 轉為字串，null 或空值回傳空字串
func rawToString(raw json.RawMessage) string {
    if len(raw) == 0 || string(raw) == "null" {
        return ""
    }
    return string(raw)
}

// ParseEvent 解析 EventBridge 包裝的 CloudTrail 事件，回傳 AuditEvent。
// 缺失欄位以空字串處理，不會 panic。
func ParseEvent(payload []byte) (*AuditEvent, error) {
    var envelope eventBridgeEvent
    if err := json.Unmarshal(payload, &envelope); err != nil {
        return nil, err
    }

    var detail cloudTrailDetail
    if err := json.Unmarshal(envelope.Detail, &detail); err != nil {
        return nil, err
    }

    // 提取 userIdentity：優先 ARN，fallback userName，再 fallback type
    identity := detail.UserIdentity.ARN
    if identity == "" {
        identity = detail.UserIdentity.UserName
    }
    if identity == "" {
        identity = detail.UserIdentity.Type
    }

    // 提取 resources ARN 列表
    resources := make([]string, 0, len(detail.Resources))
    for _, r := range detail.Resources {
        resources = append(resources, r.ARN)
    }

    ev := &AuditEvent{
        EventName:         detail.EventName,
        EventSource:       detail.EventSource,
        EventTime:         detail.EventTime,
        AWSRegion:         detail.AWSRegion,
        SourceIPAddress:   detail.SourceIPAddress,
        UserIdentity:      identity,
        Resources:         resources,
        ErrorCode:         detail.ErrorCode,
        ErrorMessage:      detail.ErrorMessage,
        RequestParameters: rawToString(detail.RequestParameters),
        ResponseElements:  rawToString(detail.ResponseElements),
    }

    // ConsoleLogin 特殊處理
    if detail.EventName == "ConsoleLogin" {
        // 操作者優先用 userName
        if detail.UserIdentity.UserName != "" {
            ev.UserIdentity = detail.UserIdentity.UserName
        }
        // 從 additionalEventData 解析 MFAUsed
        if len(detail.AdditionalEventData) > 0 {
            var additional struct {
                MFAUsed string `json:"MFAUsed"`
            }
            if json.Unmarshal(detail.AdditionalEventData, &additional) == nil {
                ev.MFAUsed = additional.MFAUsed
            }
        }
        // 從 responseElements 解析登入結果
        if len(detail.ResponseElements) > 0 && string(detail.ResponseElements) != "null" {
            var respElems struct {
                ConsoleLogin string `json:"ConsoleLogin"`
            }
            if json.Unmarshal(detail.ResponseElements, &respElems) == nil {
                ev.LoginResult = respElems.ConsoleLogin
            }
        }
    }

    return ev, nil
}
