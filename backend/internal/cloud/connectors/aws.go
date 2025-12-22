package connectors

import (
	"context"
	"fmt"
	"time"

	"enterprise-core/backend/internal/cloud"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// AWSConnector handles AWS cloud telemetry ingestion
type AWSConnector struct {
	cfg        aws.Config
	cloudtrail *cloudtrail.Client
	ec2        *ec2.Client
	region     string
	accountID  string
}

// NewAWSConnector creates a new AWS connector
func NewAWSConnector(ctx context.Context, region string) (*AWSConnector, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &AWSConnector{
		cfg:        cfg,
		cloudtrail: cloudtrail.NewFromConfig(cfg),
		ec2:        ec2.NewFromConfig(cfg),
		region:     region,
	}, nil
}

// FetchAssets retrieves EC2 instances from AWS
func (c *AWSConnector) FetchAssets(ctx context.Context) ([]*cloud.CloudAsset, error) {
	input := &ec2.DescribeInstancesInput{}
	
	result, err := c.ec2.DescribeInstances(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to describe EC2 instances: %w", err)
	}

	assets := make([]*cloud.CloudAsset, 0)
	
	for _, reservation := range result.Reservations {
		for _, instance := range reservation.Instances {
			asset := c.convertEC2Instance(instance)
			assets = append(assets, asset)
		}
	}

	return assets, nil
}

// convertEC2Instance converts AWS EC2 instance to CloudAsset
func (c *AWSConnector) convertEC2Instance(instance types.Instance) *cloud.CloudAsset {
	asset := &cloud.CloudAsset{
		ID:          fmt.Sprintf("aws:ec2:%s", aws.ToString(instance.InstanceId)),
		Provider:    cloud.ProviderAWS,
		Type:        "ec2",
		Name:        aws.ToString(instance.InstanceId),
		Region:      c.region,
		Tags:        make(map[string]string),
		State:       string(instance.State.Name),
		LastSeen:    time.Now(),
		RawMetadata: make(map[string]interface{}),
	}

	// Extract name from tags
	for _, tag := range instance.Tags {
		key := aws.ToString(tag.Key)
		value := aws.ToString(tag.Value)
		asset.Tags[key] = value
		if key == "Name" {
			asset.Name = value
		}
	}

	// Network information
	if instance.PrivateIpAddress != nil {
		asset.PrivateIP = aws.ToString(instance.PrivateIpAddress)
	}
	if instance.PublicIpAddress != nil {
		asset.PublicIP = aws.ToString(instance.PublicIpAddress)
	}
	if instance.VpcId != nil {
		asset.VPC = aws.ToString(instance.VpcId)
	}

	// Security groups
	asset.SecurityGroups = make([]string, 0, len(instance.SecurityGroups))
	for _, sg := range instance.SecurityGroups {
		asset.SecurityGroups = append(asset.SecurityGroups, aws.ToString(sg.GroupId))
	}

	// IAM role
	if instance.IamInstanceProfile != nil {
		asset.IAMRole = aws.ToString(instance.IamInstanceProfile.Arn)
	}

	// Launch time
	if instance.LaunchTime != nil {
		asset.CreatedAt = *instance.LaunchTime
	}

	// Calculate risk score based on exposure
	asset.RiskScore = c.calculateRiskScore(asset)
	asset.RiskFactors = c.identifyRiskFactors(asset)

	return asset
}

// calculateRiskScore calculates risk score for an asset
func (c *AWSConnector) calculateRiskScore(asset *cloud.CloudAsset) int {
	score := 0

	// Public IP exposure
	if asset.PublicIP != "" {
		score += 30
	}

	// State-based risk
	if asset.State == "running" {
		score += 10 // Running instances have attack surface
	}

	// IAM role attached
	if asset.IAMRole != "" {
		score += 20 // Potential privilege escalation
	}

	// Multiple security groups
	if len(asset.SecurityGroups) > 2 {
		score += 10
	}

	return min(score, 100)
}

// identifyRiskFactors identifies specific risk factors
func (c *AWSConnector) identifyRiskFactors(asset *cloud.CloudAsset) []string {
	factors := make([]string, 0)

	if asset.PublicIP != "" {
		factors = append(factors, "Public IP Address")
	}

	if asset.IAMRole != "" {
		factors = append(factors, "IAM Role Attached")
	}

	if len(asset.SecurityGroups) > 3 {
		factors = append(factors, "Multiple Security Groups")
	}

	if asset.State == "running" && asset.PublicIP != "" {
		factors = append(factors, "Internet-Facing Instance")
	}

	return factors
}

// FetchEvents retrieves CloudTrail events
func (c *AWSConnector) FetchEvents(ctx context.Context, startTime, endTime time.Time) ([]*cloud.CloudEvent, error) {
	input := &cloudtrail.LookupEventsInput{
		StartTime: aws.Time(startTime),
		EndTime:   aws.Time(endTime),
		MaxResults: aws.Int32(50), // Limit for demo
	}

	result, err := c.cloudtrail.LookupEvents(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup CloudTrail events: %w", err)
	}

	events := make([]*cloud.CloudEvent, 0, len(result.Events))
	
	for _, event := range result.Events {
		cloudEvent := &cloud.CloudEvent{
			ID:        aws.ToString(event.EventId),
			Provider:  cloud.ProviderAWS,
			EventName: aws.ToString(event.EventName),
			EventType: c.normalizeEventType(aws.ToString(event.EventName)),
			Timestamp: aws.ToTime(event.EventTime),
			Region:    c.region,
			RawEvent:  make(map[string]interface{}),
		}

		// Extract actor information
		if event.Username != nil {
			cloudEvent.ActorName = aws.ToString(event.Username)
			cloudEvent.ActorID = aws.ToString(event.Username)
			cloudEvent.ActorType = "user"
		}

		// Extract resource information
		if len(event.Resources) > 0 {
			cloudEvent.ResourceID = aws.ToString(event.Resources[0].ResourceName)
			cloudEvent.ResourceType = aws.ToString(event.Resources[0].ResourceType)
		}

		// Success determination (CloudTrail doesn't have explicit success field)
		cloudEvent.Success = true // Assume success unless error code present

		events = append(events, cloudEvent)
	}

	return events, nil
}

// normalizeEventType converts AWS event name to generic event type
func (c *AWSConnector) normalizeEventType(eventName string) string {
	// Map common AWS events to generic types
	eventMap := map[string]string{
		"RunInstances":       "compute:instance:create",
		"TerminateInstances": "compute:instance:delete",
		"StopInstances":      "compute:instance:stop",
		"StartInstances":     "compute:instance:start",
		"CreateUser":         "iam:user:create",
		"DeleteUser":         "iam:user:delete",
		"AttachUserPolicy":   "iam:policy:attach",
		"CreateBucket":       "storage:bucket:create",
		"DeleteBucket":       "storage:bucket:delete",
	}

	if normalized, ok := eventMap[eventName]; ok {
		return normalized
	}

	return eventName
}

// GetAccountID retrieves the AWS account ID
func (c *AWSConnector) GetAccountID() string {
	return c.accountID
}

// SetAccountID sets the AWS account ID
func (c *AWSConnector) SetAccountID(accountID string) {
	c.accountID = accountID
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
