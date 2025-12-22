package connectors

import (
	"context"
	"fmt"
	"strings"
	"time"

	"enterprise-core/backend/internal/cloud"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
)

// AzureConnector handles Azure cloud telemetry ingestion
type AzureConnector struct {
	cred           *azidentity.DefaultAzureCredential
	subscriptionID string
	location       string
	computeClient  *armcompute.VirtualMachinesClient
	monitorClient  *armmonitor.ActivityLogsClient
}

// NewAzureConnector creates a new Azure connector
func NewAzureConnector(subscriptionID, location string) (*AzureConnector, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	computeClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create compute client: %w", err)
	}

	monitorClient, err := armmonitor.NewActivityLogsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create monitor client: %w", err)
	}

	return &AzureConnector{
		cred:           cred,
		subscriptionID: subscriptionID,
		location:       location,
		computeClient:  computeClient,
		monitorClient:  monitorClient,
	}, nil
}

// FetchAssets retrieves Virtual Machines from Azure
func (c *AzureConnector) FetchAssets(ctx context.Context) ([]*cloud.CloudAsset, error) {
	assets := make([]*cloud.CloudAsset, 0)

	// List all VMs in subscription
	pager := c.computeClient.NewListAllPager(nil)
	
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list VMs: %w", err)
		}

		for _, vm := range page.Value {
			asset := c.convertAzureVM(vm)
			assets = append(assets, asset)
		}
	}

	return assets, nil
}

// convertAzureVM converts Azure VM to CloudAsset
func (c *AzureConnector) convertAzureVM(vm *armcompute.VirtualMachine) *cloud.CloudAsset {
	asset := &cloud.CloudAsset{
		ID:          fmt.Sprintf("azure:vm:%s", *vm.ID),
		Provider:    cloud.ProviderAzure,
		Type:        "vm",
		Name:        *vm.Name,
		Region:      *vm.Location,
		Account:     c.subscriptionID,
		Tags:        make(map[string]string),
		State:       c.getVMState(vm),
		LastSeen:    time.Now(),
		RawMetadata: make(map[string]interface{}),
	}

	// Extract tags
	if vm.Tags != nil {
		for key, value := range vm.Tags {
			if value != nil {
				asset.Tags[key] = *value
			}
		}
	}

	// Network information (simplified - would need network interface lookup)
	if vm.Properties != nil && vm.Properties.NetworkProfile != nil {
		// In production, would fetch network interface details
		asset.PrivateIP = "10.0.0.0" // Placeholder
	}

	// Calculate risk score
	asset.RiskScore = c.calculateRiskScore(asset)
	asset.RiskFactors = c.identifyRiskFactors(asset)

	return asset
}

// getVMState extracts VM state from properties
func (c *AzureConnector) getVMState(vm *armcompute.VirtualMachine) string {
	if vm.Properties != nil && vm.Properties.ProvisioningState != nil {
		return *vm.Properties.ProvisioningState
	}
	return "unknown"
}

// calculateRiskScore calculates risk score for Azure asset
func (c *AzureConnector) calculateRiskScore(asset *cloud.CloudAsset) int {
	score := 0

	// Public IP exposure
	if asset.PublicIP != "" {
		score += 30
	}

	// Running state
	if strings.ToLower(asset.State) == "succeeded" {
		score += 10
	}

	// Production tag
	if env, ok := asset.Tags["environment"]; ok && strings.ToLower(env) == "production" {
		score += 15
	}

	return min(score, 100)
}

// identifyRiskFactors identifies specific risk factors for Azure
func (c *AzureConnector) identifyRiskFactors(asset *cloud.CloudAsset) []string {
	factors := make([]string, 0)

	if asset.PublicIP != "" {
		factors = append(factors, "Public Endpoint")
	}

	if env, ok := asset.Tags["environment"]; ok && strings.ToLower(env) == "production" {
		factors = append(factors, "Production Environment")
	}

	return factors
}

// FetchEvents retrieves Azure Activity Log events
func (c *AzureConnector) FetchEvents(ctx context.Context, startTime, endTime time.Time) ([]*cloud.CloudEvent, error) {
	events := make([]*cloud.CloudEvent, 0)

	// Azure Activity Logs filter
	filter := fmt.Sprintf("eventTimestamp ge '%s' and eventTimestamp le '%s'",
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339))

	pager := c.monitorClient.NewListPager(filter, nil)

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list activity logs: %w", err)
		}

		for _, event := range page.Value {
			cloudEvent := c.convertAzureEvent(event)
			events = append(events, cloudEvent)
		}
	}

	return events, nil
}

// convertAzureEvent converts Azure Activity Log to CloudEvent
func (c *AzureConnector) convertAzureEvent(event *armmonitor.EventData) *cloud.CloudEvent {
	cloudEvent := &cloud.CloudEvent{
		ID:       *event.ID,
		Provider: cloud.ProviderAzure,
		Region:   c.location,
		Account:  c.subscriptionID,
		RawEvent: make(map[string]interface{}),
	}

	if event.EventName != nil && event.EventName.Value != nil {
		cloudEvent.EventName = *event.EventName.Value
		cloudEvent.EventType = c.normalizeEventType(*event.EventName.Value)
	}

	if event.EventTimestamp != nil {
		cloudEvent.Timestamp = *event.EventTimestamp
	}

	if event.Caller != nil {
		cloudEvent.ActorName = *event.Caller
		cloudEvent.ActorID = *event.Caller
		cloudEvent.ActorType = "user"
	}

	if event.ResourceID != nil {
		cloudEvent.ResourceID = *event.ResourceID
	}

	if event.ResourceType != nil && event.ResourceType.Value != nil {
		cloudEvent.ResourceType = *event.ResourceType.Value
	}

	// Determine success from status
	if event.Status != nil && event.Status.Value != nil {
		cloudEvent.Success = strings.ToLower(*event.Status.Value) == "succeeded"
	}

	return cloudEvent
}

// normalizeEventType converts Azure operation name to generic event type
func (c *AzureConnector) normalizeEventType(operationName string) string {
	lower := strings.ToLower(operationName)

	eventMap := map[string]string{
		"microsoft.compute/virtualmachines/write":  "compute:instance:create",
		"microsoft.compute/virtualmachines/delete": "compute:instance:delete",
		"microsoft.storage/storageaccounts/write":  "storage:account:create",
		"microsoft.storage/storageaccounts/delete": "storage:account:delete",
	}

	for key, value := range eventMap {
		if strings.Contains(lower, key) {
			return value
		}
	}

	return operationName
}

// GetSubscriptionID returns the subscription ID
func (c *AzureConnector) GetSubscriptionID() string {
	return c.subscriptionID
}
