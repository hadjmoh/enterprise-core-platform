package connectors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"enterprise-core/backend/internal/cloud"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/logging/logadmin"
	"google.golang.org/api/iterator"
)

// GCPConnector handles GCP cloud telemetry ingestion
type GCPConnector struct {
	projectID      string
	zone           string
	computeClient  *compute.InstancesClient
	loggingClient  *logadmin.Client
}

// NewGCPConnector creates a new GCP connector
func NewGCPConnector(ctx context.Context, projectID, zone string) (*GCPConnector, error) {
	computeClient, err := compute.NewInstancesRESTClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	loggingClient, err := logadmin.NewClient(ctx, projectID)
	if err != nil {
		computeClient.Close()
		return nil, fmt.Errorf("failed to create logging client: %w", err)
	}

	return &GCPConnector{
		projectID:     projectID,
		zone:          zone,
		computeClient: computeClient,
		loggingClient: loggingClient,
	}, nil
}

// Close closes the GCP clients
func (c *GCPConnector) Close() error {
	if err := c.computeClient.Close(); err != nil {
		return err
	}
	return c.loggingClient.Close()
}

// FetchAssets retrieves Compute Engine instances from GCP
func (c *GCPConnector) FetchAssets(ctx context.Context) ([]*cloud.CloudAsset, error) {
	assets := make([]*cloud.CloudAsset, 0)

	req := &computepb.ListInstancesRequest{
		Project: c.projectID,
		Zone:    c.zone,
	}

	it := c.computeClient.List(ctx, req)
	
	for {
		instance, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list instances: %w", err)
		}

		asset := c.convertGCPInstance(instance)
		assets = append(assets, asset)
	}

	return assets, nil
}

// convertGCPInstance converts GCP Compute Instance to CloudAsset
func (c *GCPConnector) convertGCPInstance(instance *computepb.Instance) *cloud.CloudAsset {
	asset := &cloud.CloudAsset{
		ID:          fmt.Sprintf("gcp:compute:%s", *instance.Id),
		Provider:    cloud.ProviderGCP,
		Type:        "compute-instance",
		Name:        *instance.Name,
		Region:      c.zone,
		Account:     c.projectID,
		Tags:        make(map[string]string),
		State:       strings.ToLower(*instance.Status),
		LastSeen:    time.Now(),
		RawMetadata: make(map[string]interface{}),
	}

	// Extract labels (GCP's version of tags)
	if instance.Labels != nil {
		for key, value := range instance.Labels {
			asset.Tags[key] = value
		}
	}

	// Network information
	if len(instance.NetworkInterfaces) > 0 {
		ni := instance.NetworkInterfaces[0]
		if ni.NetworkIP != nil {
			asset.PrivateIP = *ni.NetworkIP
		}
		
		// Check for external IP
		if len(ni.AccessConfigs) > 0 && ni.AccessConfigs[0].NatIP != nil {
			asset.PublicIP = *ni.AccessConfigs[0].NatIP
		}
	}

	// Service accounts (similar to IAM roles)
	if len(instance.ServiceAccounts) > 0 {
		asset.IAMRole = *instance.ServiceAccounts[0].Email
		asset.Permissions = instance.ServiceAccounts[0].Scopes
	}

	// Creation timestamp
	if instance.CreationTimestamp != nil {
		if t, err := time.Parse(time.RFC3339, *instance.CreationTimestamp); err == nil {
			asset.CreatedAt = t
		}
	}

	// Calculate risk score
	asset.RiskScore = c.calculateRiskScore(asset)
	asset.RiskFactors = c.identifyRiskFactors(asset)

	return asset
}

// calculateRiskScore calculates risk score for GCP asset
func (c *GCPConnector) calculateRiskScore(asset *cloud.CloudAsset) int {
	score := 0

	// Public IP exposure
	if asset.PublicIP != "" {
		score += 30
	}

	// Running state
	if asset.State == "running" {
		score += 10
	}

	// Service account attached
	if asset.IAMRole != "" {
		score += 20
	}

	// Broad permissions
	if len(asset.Permissions) > 5 {
		score += 15
	}

	return min(score, 100)
}

// identifyRiskFactors identifies specific risk factors for GCP
func (c *GCPConnector) identifyRiskFactors(asset *cloud.CloudAsset) []string {
	factors := make([]string, 0)

	if asset.PublicIP != "" {
		factors = append(factors, "External IP Address")
	}

	if asset.IAMRole != "" {
		factors = append(factors, "Service Account Attached")
	}

	if len(asset.Permissions) > 5 {
		factors = append(factors, "Broad IAM Scopes")
	}

	if asset.State == "running" && asset.PublicIP != "" {
		factors = append(factors, "Internet-Accessible Instance")
	}

	return factors
}

// FetchEvents retrieves Cloud Audit Logs from GCP
func (c *GCPConnector) FetchEvents(ctx context.Context, startTime, endTime time.Time) ([]*cloud.CloudEvent, error) {
	events := make([]*cloud.CloudEvent, 0)

	// Build filter for audit logs
	filter := fmt.Sprintf(
		`logName:"cloudaudit.googleapis.com" AND timestamp>="%s" AND timestamp<="%s"`,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339),
	)

	it := c.loggingClient.Entries(ctx,
		logadmin.Filter(filter),
		logadmin.NewestFirst(),
	)

	// Limit to 50 entries for demo
	count := 0
	maxEntries := 50

	for count < maxEntries {
		entry, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to fetch log entries: %w", err)
		}

		cloudEvent := c.convertGCPLogEntry(entry)
		if cloudEvent != nil {
			events = append(events, cloudEvent)
			count++
		}
	}

	return events, nil
}

// convertGCPLogEntry converts GCP log entry to CloudEvent
func (c *GCPConnector) convertGCPLogEntry(entry *logadmin.Entry) *cloud.CloudEvent {
	cloudEvent := &cloud.CloudEvent{
		ID:        entry.InsertID,
		Provider:  cloud.ProviderGCP,
		Timestamp: entry.Timestamp,
		Region:    c.zone,
		Account:   c.projectID,
		RawEvent:  make(map[string]interface{}),
	}

	// Extract method name from protoPayload
	if payload, ok := entry.Payload.(map[string]interface{}); ok {
		if methodName, ok := payload["methodName"].(string); ok {
			cloudEvent.EventName = methodName
			cloudEvent.EventType = c.normalizeEventType(methodName)
		}

		// Extract authentication info
		if authInfo, ok := payload["authenticationInfo"].(map[string]interface{}); ok {
			if principalEmail, ok := authInfo["principalEmail"].(string); ok {
				cloudEvent.ActorName = principalEmail
				cloudEvent.ActorID = principalEmail
				cloudEvent.ActorType = "user"
			}
		}

		// Extract resource name
		if resourceName, ok := payload["resourceName"].(string); ok {
			cloudEvent.ResourceID = resourceName
		}

		// Extract status
		if status, ok := payload["status"].(map[string]interface{}); ok {
			if code, ok := status["code"].(float64); ok {
				cloudEvent.Success = code == 0
				if code != 0 {
					cloudEvent.ErrorCode = fmt.Sprintf("%d", int(code))
				}
			}
		}
	}

	return cloudEvent
}

// normalizeEventType converts GCP method name to generic event type
func (c *GCPConnector) normalizeEventType(methodName string) string {
	lower := strings.ToLower(methodName)

	eventMap := map[string]string{
		"compute.instances.insert": "compute:instance:create",
		"compute.instances.delete": "compute:instance:delete",
		"compute.instances.stop":   "compute:instance:stop",
		"compute.instances.start":  "compute:instance:start",
		"storage.buckets.create":   "storage:bucket:create",
		"storage.buckets.delete":   "storage:bucket:delete",
	}

	for key, value := range eventMap {
		if strings.Contains(lower, key) {
			return value
		}
	}

	return methodName
}

// GetProjectID returns the project ID
func (c *GCPConnector) GetProjectID() string {
	return c.projectID
}
