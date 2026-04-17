# Lambda IAM 權限說明

**語言: [English](../en/iam-policy.md) | 繁體中文**

本文件說明 audit-notifier Lambda 所需的 IAM 權限。IAM Role 和 EventBridge Target 不在本專案範圍內，需在其他 IaC 專案中建立。

## Lambda 執行角色

### Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

### Permissions Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CloudWatchLogs",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:aws:logs:*:123456789012:log-group:/aws/lambda/audit-notifier:*"
    },
    {
      "Sid": "SSMGetParameter",
      "Effect": "Allow",
      "Action": "ssm:GetParameter",
      "Resource": "arn:aws:ssm:*:123456789012:parameter/audit-notifier/encryption-key"
    }
  ]
}
```

> **注意：** 如果 SSM SecureString 使用 AWS managed key（`alias/aws/ssm`），則不需要額外的 `kms:Decrypt` 權限。若使用 customer managed key（CMK），需加入以下 statement：

```json
{
  "Sid": "KMSDecrypt",
  "Effect": "Allow",
  "Action": "kms:Decrypt",
  "Resource": "arn:aws:kms:*:123456789012:key/<YOUR-CMK-KEY-ID>"
}
```

## EventBridge Lambda Permission

需要在 Lambda 上新增 resource-based policy，允許 EventBridge invoke：

```json
{
  "Sid": "EventBridgeInvoke",
  "Effect": "Allow",
  "Principal": {
    "Service": "events.amazonaws.com"
  },
  "Action": "lambda:InvokeFunction",
  "Resource": "arn:aws:lambda:*:123456789012:function:audit-notifier",
  "Condition": {
    "ArnLike": {
      "AWS:SourceArn": "arn:aws:events:*:123456789012:rule/<RULE-NAME>"
    }
  }
}
```

## 權限說明

| # | 權限 | 用途 |
|---|---|---|
| 1 | CloudWatch Logs | Lambda 執行日誌寫入所需權限 |
| 2 | SSM GetParameter | 讀取 ENCRYPTION_KEY（SecureString 類型） |
| 3 | KMS Decrypt | 僅在使用 customer managed key（CMK）時需要。使用 AWS managed key（`alias/aws/ssm`）時不需要 |
| 4 | EventBridge Lambda Permission | 屬於 Lambda resource-based policy，允許特定 EventBridge Rule 觸發此 Lambda |

> 請將 `123456789012` 替換為實際的 AWS Account ID，`<RULE-NAME>` 替換為 EventBridge Rule 名稱。

## EventBridge Cross-Region IAM Role

跨 Region 轉發事件（PutEvents）所需的 IAM Role，供 EventBridge rules 使用。

### Trust Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
```

### Permissions Policy

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "EventBridgePutEventsToApEast1",
      "Effect": "Allow",
      "Action": "events:PutEvents",
      "Resource": "arn:aws:events:ap-east-1:123456789012:event-bus/default"
    }
  ]
}
```

## IAM 資源總覽

| 資源 | 類型 | Trust | 權限 |
|---|---|---|---|
| `Audit-Notice-Lambda-Role` | IAM Role | lambda.amazonaws.com | AuditNoticeLambdaPolicy |
| `AuditNoticeLambdaPolicy` | IAM Policy | - | logs:CreateLogGroup/Stream/PutLogEvents, ssm:GetParameter |
| `Audit-EventBridge-CrossRegion-Role` | IAM Role | events.amazonaws.com | events:PutEvents to ap-east-1 default bus |
