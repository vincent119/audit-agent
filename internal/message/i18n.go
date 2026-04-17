package message

// 翻譯 key 常數
const (
    KeyTitle             = "title"
    KeyEventName         = "event_name"
    KeyEventSource       = "event_source"
    KeyEventTime         = "event_time"
    KeyRegion            = "region"
    KeySourceIP          = "source_ip"
    KeyUserIdentity      = "user_identity"
    KeyResources         = "resources"
    KeyErrorCode         = "error_code"
    KeyErrorMessage      = "error_message"
    KeyRequestParameters = "request_parameters"
    KeyResponseElements  = "response_elements"
    KeyLoginResult       = "login_result"
    KeyMFAUsed           = "mfa_used"
)

// 多語言翻譯表
var translations = map[string]map[string]string{
    "en": {
        KeyTitle:             "AWS CloudTrail Audit Alert",
        KeyEventName:         "Event Name",
        KeyEventSource:       "Event Source",
        KeyEventTime:         "Event Time",
        KeyRegion:            "Region",
        KeySourceIP:          "Source IP",
        KeyUserIdentity:      "User",
        KeyResources:         "Resources",
        KeyErrorCode:         "Error Code",
        KeyErrorMessage:      "Error Message",
        KeyRequestParameters: "Request Parameters",
        KeyResponseElements:  "Response Elements",
        KeyLoginResult:       "Login Result",
        KeyMFAUsed:           "MFA",
    },
    "zh-TW": {
        KeyTitle:             "AWS CloudTrail 審計警報",
        KeyEventName:         "事件名稱",
        KeyEventSource:       "事件來源",
        KeyEventTime:         "事件時間",
        KeyRegion:            "區域",
        KeySourceIP:          "來源 IP",
        KeyUserIdentity:      "操作者",
        KeyResources:         "影響資源",
        KeyErrorCode:         "錯誤代碼",
        KeyErrorMessage:      "錯誤訊息",
        KeyRequestParameters: "請求參數",
        KeyResponseElements:  "回應內容",
        KeyLoginResult:       "登入結果",
        KeyMFAUsed:           "MFA",
    },
    "zh-CN": {
        KeyTitle:             "AWS CloudTrail 审计警报",
        KeyEventName:         "事件名称",
        KeyEventSource:       "事件来源",
        KeyEventTime:         "事件时间",
        KeyRegion:            "区域",
        KeySourceIP:          "来源 IP",
        KeyUserIdentity:      "操作者",
        KeyResources:         "影响资源",
        KeyErrorCode:         "错误代码",
        KeyErrorMessage:      "错误消息",
        KeyRequestParameters: "请求参数",
        KeyResponseElements:  "响应内容",
        KeyLoginResult:       "登入结果",
        KeyMFAUsed:           "MFA",
    },
}

// GetTranslation 取得指定語言的翻譯文字，找不到語言時 fallback 到 en
func GetTranslation(lang, key string) string {
    if t, ok := translations[lang]; ok {
        if v, ok := t[key]; ok {
            return v
        }
    }
    // fallback 到英文
    if t, ok := translations["en"]; ok {
        if v, ok := t[key]; ok {
            return v
        }
    }
    return key
}
