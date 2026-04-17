# Lambda IAM 權限說明

本文件說明 audit-notifier Lambda 所需的 IAM 權限。IAM Role 和 EventBridge Target 不在本專案範圍內，需在其他 IaC 專案中建立。

## Trust Policy

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

## Permissions Policy

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

注意：如果 SSM SecureString 使用 AWS managed key（`alias/aws/ssm`），則不需要額外的 `kms:Decrypt` 權限。若使用 customer managed key（CMK），需加入以下 statement：

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

## 說明

1. CloudWatch Logs：Lambda 執行日誌寫入所需權限
2. SSM GetParameter：讀取 ENCRYPTION_KEY（SecureString 類型）
3. KMS Decrypt：僅在使用 customer managed key（CMK）時需要。使用 AWS managed key（`alias/aws/ssm`）時不需要
4. EventBridge Lambda Permission：屬於 Lambda resource-based policy，允許特定 EventBridge Rule 觸發此 Lambda

請將 `123456789012` 替換為實際的 AWS Account ID，`<RULE-NAME>` 替換為 EventBridge Rule 名稱。

---

# Lambda IAM Policy Reference

This document describes the IAM permissions required by the audit-notifier Lambda function. The IAM Role and EventBridge Target are not managed by this project and should be created in a separate IaC project.

## Trust Policy

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

## Permissions Policy

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

Note: If the SSM SecureString uses the AWS managed key (`alias/aws/ssm`), no additional `kms:Decrypt` permission is needed. If using a customer managed key (CMK), add the following statement:

```json
{
  "Sid": "KMSDecrypt",
  "Effect": "Allow",
  "Action": "kms:Decrypt",
  "Resource": "arn:aws:kms:*:123456789012:key/<YOUR-CMK-KEY-ID>"
}
```

## EventBridge Lambda Permission

A resource-based policy must be added to the Lambda function to allow EventBridge to invoke it:

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

## Notes

1. CloudWatch Logs: Required for Lambda execution log output
2. SSM GetParameter: Reads the ENCRYPTION_KEY (SecureString type)
3. KMS Decrypt: Only required when using a customer managed key (CMK). Not needed when using the AWS managed key (`alias/aws/ssm`)
4. EventBridge Lambda Permission: This is a Lambda resource-based policy that allows a specific EventBridge Rule to invoke this Lambda

Replace `123456789012` with your actual AWS Account ID and `<RULE-NAME>` with the EventBridge Rule name.
