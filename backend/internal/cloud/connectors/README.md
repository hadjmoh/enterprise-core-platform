# Cloud Connectors

Multi-cloud telemetry connectors for AWS, Azure, and GCP.

## Overview

The cloud connectors enable real-time ingestion of cloud assets and events from multiple cloud providers:

- **AWS**: EC2 instances, CloudTrail events
- **Azure**: Virtual Machines, Activity Logs
- **GCP**: Compute Engine instances, Cloud Audit Logs

## Architecture

```
ConnectorManager
├── AWSConnector (AWS SDK v2)
├── AzureConnector (Azure SDK for Go)
└── GCPConnector (GCP Client Libraries)
```

## Prerequisites

### AWS
- AWS credentials configured (via environment variables, IAM role, or `~/.aws/credentials`)
- Required permissions:
  - `ec2:DescribeInstances`
  - `cloudtrail:LookupEvents`

### Azure
- Azure credentials configured (via Azure CLI, environment variables, or managed identity)
- Required permissions:
  - `Microsoft.Compute/virtualMachines/read`
  - `Microsoft.Insights/activityLogs/read`

### GCP
- GCP credentials configured (via `GOOGLE_APPLICATION_CREDENTIALS` environment variable)
- Required permissions:
  - `compute.instances.list`
  - `logging.logEntries.list`

## Configuration

### Environment Variables

```bash
# AWS
export AWS_REGION=us-east-1
export AWS_ACCOUNT_ID=123456789012

# Azure
export AZURE_SUBSCRIPTION_ID=your-subscription-id
export AZURE_LOCATION=eastus

# GCP
export GCP_PROJECT_ID=your-project-id
export GCP_ZONE=us-central1-a
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
```

## Usage

### Basic Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "enterprise-core-platform/backend/internal/cloud/connectors"
)

func main() {
    ctx := context.Background()

    // Configure connectors
    config := connectors.ConnectorConfig{
        AWSRegion:           "us-east-1",
        AWSAccountID:        "123456789012",
        AzureSubscriptionID: "your-subscription-id",
        AzureLocation:       "eastus",
        GCPProjectID:        "your-project-id",
        GCPZone:             "us-central1-a",
    }

    // Create manager
    manager, err := connectors.NewConnectorManager(ctx, config)
    if err != nil {
        panic(err)
    }
    defer manager.Close()

    // Fetch all assets
    assets, err := manager.FetchAllAssets(ctx)
    if err != nil {
        fmt.Printf("Error fetching assets: %v\n", err)
    }

    fmt.Printf("Fetched %d assets\n", len(assets))

    // Fetch events from last hour
    endTime := time.Now()
    startTime := endTime.Add(-1 * time.Hour)
    
    events, err := manager.FetchAllEvents(ctx, startTime, endTime)
    if err != nil {
        fmt.Printf("Error fetching events: %v\n", err)
    }

    fmt.Printf("Fetched %d events\n", len(events))
}
```

### Individual Connector Usage

#### AWS

```go
awsConn, err := connectors.NewAWSConnector(ctx, "us-east-1")
if err != nil {
    panic(err)
}

assets, err := awsConn.FetchAssets(ctx)
events, err := awsConn.FetchEvents(ctx, startTime, endTime)
```

#### Azure

```go
azureConn, err := connectors.NewAzureConnector("subscription-id", "eastus")
if err != nil {
    panic(err)
}

assets, err := azureConn.FetchAssets(ctx)
events, err := azureConn.FetchEvents(ctx, startTime, endTime)
```

#### GCP

```go
gcpConn, err := connectors.NewGCPConnector(ctx, "project-id", "us-central1-a")
if err != nil {
    panic(err)
}
defer gcpConn.Close()

assets, err := gcpConn.FetchAssets(ctx)
events, err := gcpConn.FetchEvents(ctx, startTime, endTime)
```

## Data Models

### CloudAsset

```go
type CloudAsset struct {
    ID             string            // Unified ID: aws:ec2:i-123456
    Provider       Provider          // aws | azure | gcp
    Type           string            // ec2 | vm | compute-instance
    Name           string
    Region         string
    Account        string
    Tags           map[string]string
    State          string
    PrivateIP      string
    PublicIP       string
    VPC            string
    SecurityGroups []string
    IAMRole        string
    Permissions    []string
    RiskScore      int               // 0-100
    RiskFactors    []string
    CreatedAt      time.Time
    LastSeen       time.Time
}
```

### CloudEvent

```go
type CloudEvent struct {
    ID           string
    Provider     Provider
    EventType    string    // Normalized: compute:instance:create
    EventName    string    // Original: RunInstances
    Timestamp    time.Time
    ActorID      string
    ActorType    string
    ActorName    string
    ResourceID   string
    ResourceType string
    SourceIP     string
    Region       string
    Account      string
    Success      bool
    ErrorCode    string
    ErrorMessage string
}
```

## Risk Scoring

Each connector calculates a risk score (0-100) based on:

### AWS
- Public IP: +30
- Running state: +10
- IAM role attached: +20
- Multiple security groups: +10

### Azure
- Public IP: +30
- Succeeded state: +10
- Production tag: +15

### GCP
- External IP: +30
- Running state: +10
- Service account: +20
- Broad IAM scopes: +15

## Event Normalization

Events are normalized to a common taxonomy:

| Provider | Original Event | Normalized Type |
|----------|---------------|-----------------|
| AWS | RunInstances | compute:instance:create |
| Azure | Microsoft.Compute/virtualMachines/write | compute:instance:create |
| GCP | compute.instances.insert | compute:instance:create |

## Error Handling

The connector manager uses concurrent fetching with error aggregation:

```go
assets, err := manager.FetchAllAssets(ctx)
// Returns partial results even if some connectors fail
// err contains aggregated errors from failed connectors
```

## Performance

- **Concurrent Execution**: All connectors run in parallel
- **Pagination**: Handles large result sets automatically
- **Rate Limiting**: Respects cloud provider API limits
- **Timeouts**: Configurable context timeouts

## Security Considerations

> [!WARNING]
> **PROTOTYPE MODE - Session 7.10 Required**
> 
> Current implementation uses environment variables for credentials.
> 
> **Production Requirements** (Session 7.10):
> - Migrate to encrypted credential vault
> - Add OPA policy enforcement
> - Implement kill-switches
> - Enable centralized audit logging
> - Add config drift detection

### Best Practices

1. **Least Privilege**: Use minimal IAM permissions
2. **Credential Rotation**: Rotate credentials regularly
3. **Audit Logging**: Log all API calls
4. **Network Security**: Use VPC endpoints where possible
5. **Encryption**: Use TLS for all API calls (enforced by SDKs)

## Troubleshooting

### AWS

```bash
# Test credentials
aws sts get-caller-identity

# Test EC2 access
aws ec2 describe-instances --region us-east-1
```

### Azure

```bash
# Test credentials
az account show

# Test VM access
az vm list --subscription your-subscription-id
```

### GCP

```bash
# Test credentials
gcloud auth application-default print-access-token

# Test Compute access
gcloud compute instances list --project your-project-id
```

## Dependencies

```go
require (
    github.com/aws/aws-sdk-go-v2 v1.x.x
    github.com/aws/aws-sdk-go-v2/config v1.x.x
    github.com/aws/aws-sdk-go-v2/service/cloudtrail v1.x.x
    github.com/aws/aws-sdk-go-v2/service/ec2 v1.x.x
    
    github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.x.x
    github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute v1.x.x
    github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v1.x.x
    
    cloud.google.com/go/compute v1.x.x
    cloud.google.com/go/logging v1.x.x
    google.golang.org/api v0.x.x
)
```

## Future Enhancements

- [ ] Additional AWS services (RDS, S3, Lambda)
- [ ] Additional Azure services (Storage, Functions)
- [ ] Additional GCP services (Cloud Storage, Cloud Functions)
- [ ] Caching layer for frequently accessed data
- [ ] Incremental sync (delta updates)
- [ ] Webhook support for real-time events
- [ ] Multi-region support
- [ ] Custom risk scoring rules

## License

Internal use only - Enterprise Core Platform
