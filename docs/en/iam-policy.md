# IAM Policy Reference

**Language: English | [繁體中文](../zh-TW/iam-policy.md)**

This document describes the IAM permissions required by the audit-notifier Lambda function. The IAM Role and EventBridge Target are not managed by this project and should be created in a separate IaC project.

## Lambda Execution Role

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

> **Note:** If the SSM SecureString uses the AWS managed key (`alias/aws/ssm`), no additional `kms:Decrypt` permission is needed. If using a customer managed key (CMK), add the following statement:

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

## Permission Details

| # | Permission | Purpose |
|---|---|---|
| 1 | CloudWatch Logs | Required for Lambda execution log output |
| 2 | SSM GetParameter | Reads the ENCRYPTION_KEY (SecureString type) |
| 3 | KMS Decrypt | Only required when using a customer managed key (CMK). Not needed with AWS managed key (`alias/aws/ssm`) |
| 4 | EventBridge Lambda Permission | Lambda resource-based policy that allows a specific EventBridge Rule to invoke this Lambda |

> Replace `123456789012` with your actual AWS Account ID and `<RULE-NAME>` with the EventBridge Rule name.

## EventBridge Cross-Region IAM Role

IAM Role required for cross-region event forwarding (PutEvents), used by EventBridge rules.

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

## IAM Resource Summary

| Resource | Type | Trust | Permissions |
|---|---|---|---|
| `Audit-Notice-Lambda-Role` | IAM Role | lambda.amazonaws.com | AuditNoticeLambdaPolicy |
| `AuditNoticeLambdaPolicy` | IAM Policy | - | logs:CreateLogGroup/Stream/PutLogEvents, ssm:GetParameter |
| `Audit-EventBridge-CrossRegion-Role` | IAM Role | events.amazonaws.com | events:PutEvents to ap-east-1 default bus |
