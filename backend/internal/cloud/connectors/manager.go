package connectors

import (
	"context"
	"fmt"
	"sync"
	"time"

	"enterprise-core/backend/internal/cloud"
)

// ConnectorManager orchestrates multiple cloud connectors
type ConnectorManager struct {
	awsConnector   *AWSConnector
	azureConnector *AzureConnector
	gcpConnector   *GCPConnector
	mu             sync.RWMutex
}

// ConnectorConfig holds configuration for all connectors
type ConnectorConfig struct {
	// AWS
	AWSRegion    string
	AWSAccountID string

	// Azure
	AzureSubscriptionID string
	AzureLocation       string

	// GCP
	GCPProjectID string
	GCPZone      string
}

// NewConnectorManager creates a new connector manager
func NewConnectorManager(ctx context.Context, config ConnectorConfig) (*ConnectorManager, error) {
	manager := &ConnectorManager{}

	// Initialize AWS connector if configured
	if config.AWSRegion != "" {
		awsConn, err := NewAWSConnector(ctx, config.AWSRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to create AWS connector: %w", err)
		}
		awsConn.SetAccountID(config.AWSAccountID)
		manager.awsConnector = awsConn
	}

	// Initialize Azure connector if configured
	if config.AzureSubscriptionID != "" {
		azureConn, err := NewAzureConnector(config.AzureSubscriptionID, config.AzureLocation)
		if err != nil {
			return nil, fmt.Errorf("failed to create Azure connector: %w", err)
		}
		manager.azureConnector = azureConn
	}

	// Initialize GCP connector if configured
	if config.GCPProjectID != "" {
		gcpConn, err := NewGCPConnector(ctx, config.GCPProjectID, config.GCPZone)
		if err != nil {
			return nil, fmt.Errorf("failed to create GCP connector: %w", err)
		}
		manager.gcpConnector = gcpConn
	}

	return manager, nil
}

// FetchAllAssets fetches assets from all configured cloud providers
func (m *ConnectorManager) FetchAllAssets(ctx context.Context) ([]*cloud.CloudAsset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allAssets := make([]*cloud.CloudAsset, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Fetch from AWS
	if m.awsConnector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assets, err := m.awsConnector.FetchAssets(ctx)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("AWS: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			allAssets = append(allAssets, assets...)
			mu.Unlock()
		}()
	}

	// Fetch from Azure
	if m.azureConnector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assets, err := m.azureConnector.FetchAssets(ctx)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("Azure: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			allAssets = append(allAssets, assets...)
			mu.Unlock()
		}()
	}

	// Fetch from GCP
	if m.gcpConnector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assets, err := m.gcpConnector.FetchAssets(ctx)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("GCP: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			allAssets = append(allAssets, assets...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(errors) > 0 {
		return allAssets, fmt.Errorf("errors fetching assets: %v", errors)
	}

	return allAssets, nil
}

// FetchAllEvents fetches events from all configured cloud providers
func (m *ConnectorManager) FetchAllEvents(ctx context.Context, startTime, endTime time.Time) ([]*cloud.CloudEvent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	allEvents := make([]*cloud.CloudEvent, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := make([]error, 0)

	// Fetch from AWS
	if m.awsConnector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			events, err := m.awsConnector.FetchEvents(ctx, startTime, endTime)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("AWS: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			allEvents = append(allEvents, events...)
			mu.Unlock()
		}()
	}

	// Fetch from Azure
	if m.azureConnector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			events, err := m.azureConnector.FetchEvents(ctx, startTime, endTime)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("Azure: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			allEvents = append(allEvents, events...)
			mu.Unlock()
		}()
	}

	// Fetch from GCP
	if m.gcpConnector != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			events, err := m.gcpConnector.FetchEvents(ctx, startTime, endTime)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("GCP: %w", err))
				mu.Unlock()
				return
			}
			mu.Lock()
			allEvents = append(allEvents, events...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	if len(errors) > 0 {
		return allEvents, fmt.Errorf("errors fetching events: %v", errors)
	}

	return allEvents, nil
}

// GetStats returns statistics about configured connectors
func (m *ConnectorManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	
	connectors := make([]string, 0)
	if m.awsConnector != nil {
		connectors = append(connectors, "aws")
	}
	if m.azureConnector != nil {
		connectors = append(connectors, "azure")
	}
	if m.gcpConnector != nil {
		connectors = append(connectors, "gcp")
	}

	stats["configured_connectors"] = connectors
	stats["connector_count"] = len(connectors)

	return stats
}

// Close closes all connectors
func (m *ConnectorManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.gcpConnector != nil {
		if err := m.gcpConnector.Close(); err != nil {
			return err
		}
	}

	return nil
}
