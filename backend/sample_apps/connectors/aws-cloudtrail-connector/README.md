# AWS CloudTrail Connector

Ingest and analyze AWS CloudTrail logs for security monitoring and compliance.

## Features

- Real-time CloudTrail event ingestion
- IAM activity monitoring
- API call tracking
- Security event alerting
- Compliance reporting

## Configuration

After installation, configure the connector with:

- **AWS Access Key ID**: Your AWS access key
- **AWS Secret Access Key**: Your AWS secret key
- **AWS Region**: Target region (e.g., us-east-1)
- **S3 Bucket**: CloudTrail log bucket name
- **Alert Threshold**: Number of failed API calls to trigger alert

## Permissions Required

- `read:cloud` - Access cloud asset information
- `write:data` - Write CloudTrail events to indexes
- `write:alerts` - Create security alerts

## Dashboard

The connector includes a comprehensive dashboard showing:

- CloudTrail events timeline
- Total events (24h)
- Top IAM users
- API call distribution

## Installation

1. Navigate to App Store
2. Search for "AWS CloudTrail Connector"
3. Click Install
4. Configure AWS credentials
5. View dashboard at `/apps/aws-cloudtrail-connector/dashboards/main`

## Support

For issues or questions, contact the Enterprise Security Team.
