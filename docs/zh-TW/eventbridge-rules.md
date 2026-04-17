# EventBridge Rules

**語言: [English](../en/eventbridge-rules.md) | 繁體中文**

## Rule 涵蓋範圍

```mermaid
graph LR
    subgraph AP_RULES [ap-east-1 Rules - Regional Events]
        subgraph CN [audit-compute-network-events]
            CN1[EC2 Instance<br/>Run/Stop/Start/Terminate/Reboot]
            CN2[Security Group<br/>Create/Delete/Authorize/Revoke/Modify]
            CN3[VPC Network<br/>Vpc/Subnet/Route/Peering/NAT]
            CN4[EBS<br/>CreateSnapshot/ModifySnapshotAttribute]
            CN5[ELB<br/>Create/Delete/Modify]
        end

        subgraph DS [audit-datastore-events]
            DS1[S3<br/>Bucket/ACL/Policy/Encryption/Versioning]
            DS2[SQS<br/>Create/Delete/SetAttributes/Purge]
            DS3[KMS<br/>DisableKey/ScheduleKeyDeletion<br/>Excluded by CloudTrail]
            DS4[RDS<br/>Delete/Modify/Snapshot/Restore]
            DS5[DynamoDB<br/>CreateTable/DeleteTable]
        end

        subgraph PS [audit-platform-security-events]
            PS1[ECR<br/>PutImage/Repository/Lifecycle]
            PS2[ECS<br/>TaskDefinition/Service]
            PS3[EKS<br/>Cluster/Config]
            PS4[Lambda<br/>Function/Code/Config/Permission]
            PS5[WAF<br/>WebACL/Rule/IPSet]
            PS6[Security Monitoring<br/>CloudTrail/Config/GuardDuty/CloudWatch]
            PS7[EventBridge<br/>PutRule/DeleteRule/PutTargets<br/>RemoveTargets/DisableRule]
            PS8[Organizations<br/>Leave/CreateAccount/Remove]
        end

        subgraph ID [audit-identity-events]
            ID1[IAM User<br/>Create/Delete/LoginProfile/AccessKey]
            ID2[IAM Role and Policy<br/>Create/Delete/Attach/Detach/Put]
            ID3[IAM MFA<br/>Deactivate/Delete]
            ID4[IAM Federation<br/>SAML/OIDC]
            ID5[STS<br/>GetFederationToken]
        end

        subgraph RECV [audit-console-login-receiver]
            RECV1[ConsoleLogin events<br/>forwarded from 4 regions]
        end
    end

    subgraph GLOBAL_RULES [us-east-1 Rules - Global Events]
        subgraph GI [audit-global-identity-events]
            GI1[ConsoleLogin + IAM + STS<br/>Same as ap-east-1 identity-events]
        end

        subgraph GPS [audit-global-platform-security-events]
            GPS1[Route53 + Organizations<br/>Same as ap-east-1 platform-security]
        end
    end

    subgraph LOGIN_RULES [ConsoleLogin Rules x4 Regions]
        subgraph CL [audit-console-login]
            CL1[us-east-1<br/>Root / IAM no cookie]
            CL2[us-east-2<br/>IAM Americas]
            CL3[eu-north-1<br/>IAM Europe]
            CL4[ap-southeast-2<br/>IAM Asia-Pacific]
        end
    end

    style DS3 fill:#5c3a1a,stroke:#ff9e4a,color:#fff
    style LOGIN_RULES fill:#1a3a5c,stroke:#4a9eff,color:#fff
    style RECV fill:#1a4a3c,stroke:#4aff9e,color:#fff
```

## ap-east-1 Regional Rules

| Rule | Source | 監控事件 |
|---|---|---|
| `audit-compute-network-events` | aws.ec2, aws.elasticloadbalancing | EC2 lifecycle, Security Group, VPC, EBS, ELB |
| `audit-datastore-events` | aws.s3, aws.sqs, aws.kms, aws.rds, aws.dynamodb | Bucket/Queue/Key/DB 操作 |
| `audit-platform-security-events` | aws.ecr/ecs/eks/lambda/route53/wafv2/waf/cloudtrail/config/guardduty/monitoring/organizations/events | 容器、Serverless、DNS、WAF、安全監控、EventBridge |
| `audit-identity-events` | aws.iam, aws.sts | IAM User/Role/Policy/MFA, STS |
| `audit-console-login-receiver` | aws.signin | ConsoleLogin（從其他 Region 轉發） |

## Cross-Region Rules

| Rule | Region | 監控事件 | Target |
|---|---|---|---|
| `audit-global-identity-events` | us-east-1 | ConsoleLogin, IAM, STS | ap-east-1 EventBridge (PutEvents) |
| `audit-global-platform-security-events` | us-east-1 | Route53, Organizations | ap-east-1 EventBridge (PutEvents) |
| `audit-console-login` | us-east-1 | ConsoleLogin | ap-east-1 EventBridge (PutEvents) |
| `audit-console-login` | us-east-2 | ConsoleLogin | ap-east-1 EventBridge (PutEvents) |
| `audit-console-login` | eu-north-1 | ConsoleLogin | ap-east-1 EventBridge (PutEvents) |
| `audit-console-login` | ap-southeast-2 | ConsoleLogin | ap-east-1 EventBridge (PutEvents) |

Cross-Region rules 使用 `Audit-EventBridge-CrossRegion-Role` IAM Role 執行 `events:PutEvents`，將事件轉發到 `ap-east-1` 的 default event bus。

## Lambda Resource-Based Policy (Permissions)

EventBridge 觸發 Lambda 需要在 Lambda 上設定 resource-based policy：

| StatementId | Source Rule | Region |
|---|---|---|
| `AllowEventBridgeComputeNetwork` | audit-compute-network-events | ap-east-1 |
| `AllowEventBridgeIdentity` | audit-identity-events | ap-east-1 |
| `AllowEventBridgeDataStore` | audit-datastore-events | ap-east-1 |
| `AllowEventBridgePlatformSecurity` | audit-platform-security-events | ap-east-1 |
| `AllowEventBridgeConsoleLoginReceiver` | audit-console-login-receiver | ap-east-1 |
| `AllowEventBridgeGlobalIdentity` | audit-global-identity-events | us-east-1 |
| `AllowEventBridgeGlobalPlatformSecurity` | audit-global-platform-security-events | us-east-1 |

## IaC 檔案結構

```
stacks/
├── global/aws/iam/
│   ├── roles/
│   │   ├── AuditNoticeLambdaRole.yaml
│   │   └── AuditEventBridgeCrossRegionRole.yaml
│   └── policies/
│       └── AuditNoticeLambdaPolicy.yaml
└── prod/aws/
    ├── EventBridge/
    │   ├── AuditTailLog/                      # ap-east-1 regional rules
    │   │   ├── AuditTailComputeNetworkRule.yaml
    │   │   ├── AuditTailDataStoreRule.yaml
    │   │   ├── AuditTailPlatformSecurityRule.yaml
    │   │   ├── AuditTailIdentityRule.yaml
    │   │   ├── AuditConsoleLoginReceiverRule.yaml
    │   │   └── AuditTailCloudTrailEventTarget.yaml
    │   ├── AuditGlobalUse1/                   # us-east-1 global + ConsoleLogin
    │   │   ├── AuditGlobalIdentityRule.yaml
    │   │   ├── AuditGlobalPlatformSecurityRule.yaml
    │   │   └── AuditGlobalEventTarget.yaml
    │   ├── AuditSigninUse2/                   # us-east-2 ConsoleLogin
    │   │   ├── AuditSigninRule.yaml
    │   │   └── AuditSigninTarget.yaml
    │   ├── AuditSigninEun1/                   # eu-north-1 ConsoleLogin
    │   │   ├── AuditSigninRule.yaml
    │   │   └── AuditSigninTarget.yaml
    │   └── AuditSigninApse2/                  # ap-southeast-2 ConsoleLogin
    │       ├── AuditSigninRule.yaml
    │       └── AuditSigninTarget.yaml
    └── Lambda/AuditLambda/
        ├── Function.yaml
        ├── LogGroup.yaml
        └── Permission.yaml
```

## 已知問題

| 問題 | 影響 | 狀態 |
|---|---|---|
| CloudTrail 排除 `kms.amazonaws.com` | KMS DisableKey, ScheduleKeyDeletion 不會觸發 | 評估中 |
| CloudTrail 事件延遲 | EventBridge 收到事件有 5-15 分鐘延遲 | AWS 限制 |
| 僅涵蓋 `ap-east-1` regional events | 其他 Region 的 regional events 不會被捕獲 | 評估擴展 |
| IAM 使用者 regional endpoint 登入 | 若使用者透過 regional endpoint 登入，ConsoleLogin 事件記錄在該 Region | 低機率，未涵蓋 |
