package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"enterprise-core/backend/internal/cloud"
	"enterprise-core/backend/internal/cloud/connectors"

	"go.uber.org/zap"
)

var (
	// Polling intervals
	assetPollInterval = flag.Duration("asset-interval", 5*time.Minute, "Asset polling interval")
	eventPollInterval = flag.Duration("event-interval", 1*time.Minute, "Event polling interval")

	// AWS config
	awsRegion    = flag.String("aws-region", "", "AWS region")
	awsAccountID = flag.String("aws-account", "", "AWS account ID")

	// Azure config
	azureSubscriptionID = flag.String("azure-subscription", "", "Azure subscription ID")
	azureLocation       = flag.String("azure-location", "", "Azure location")

	// GCP config
	gcpProjectID = flag.String("gcp-project", "", "GCP project ID")
	gcpZone      = flag.String("gcp-zone", "", "GCP zone")

	// Service config
	dryRun = flag.Bool("dry-run", false, "Dry run mode (no persistence)")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Info("Starting cloud ingestion service",
		zap.Duration("asset_interval", *assetPollInterval),
		zap.Duration("event_interval", *eventPollInterval),
		zap.Bool("dry_run", *dryRun),
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize connector manager
	config := connectors.ConnectorConfig{
		AWSRegion:           *awsRegion,
		AWSAccountID:        *awsAccountID,
		AzureSubscriptionID: *azureSubscriptionID,
		AzureLocation:       *azureLocation,
		GCPProjectID:        *gcpProjectID,
		GCPZone:             *gcpZone,
	}

	manager, err := connectors.NewConnectorManager(ctx, config)
	if err != nil {
		logger.Fatal("Failed to create connector manager", zap.Error(err))
	}
	defer manager.Close()

	// Log configured connectors
	stats := manager.GetStats()
	logger.Info("Connector manager initialized", zap.Any("stats", stats))

	// Initialize asset graph
	assetGraph := cloud.NewAssetGraph()

	// Create ingestion service
	service := &IngestionService{
		logger:     logger,
		manager:    manager,
		assetGraph: assetGraph,
		dryRun:     *dryRun,
	}

	// Start polling goroutines
	go service.pollAssets(ctx, *assetPollInterval)
	go service.pollEvents(ctx, *eventPollInterval)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logger.Info("Cloud ingestion service running. Press Ctrl+C to stop.")
	<-sigChan

	logger.Info("Shutting down cloud ingestion service...")
	cancel()

	// Give goroutines time to finish
	time.Sleep(2 * time.Second)
	logger.Info("Cloud ingestion service stopped")
}

// IngestionService handles continuous cloud data ingestion
type IngestionService struct {
	logger     *zap.Logger
	manager    *connectors.ConnectorManager
	assetGraph *cloud.AssetGraph
	dryRun     bool
}

// pollAssets continuously fetches cloud assets
func (s *IngestionService) pollAssets(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on startup
	s.fetchAssets(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Asset polling stopped")
			return
		case <-ticker.C:
			s.fetchAssets(ctx)
		}
	}
}

// fetchAssets fetches assets from all cloud providers
func (s *IngestionService) fetchAssets(ctx context.Context) {
	s.logger.Info("Fetching cloud assets...")

	assets, err := s.manager.FetchAllAssets(ctx)
	if err != nil {
		s.logger.Error("Failed to fetch assets", zap.Error(err))
		return
	}

	s.logger.Info("Assets fetched",
		zap.Int("count", len(assets)),
	)

	// Update asset graph
	for _, asset := range assets {
		s.assetGraph.AddAsset(asset)
	}

	// Log statistics
	stats := s.assetGraph.GetStats()
	s.logger.Info("Asset graph updated",
		zap.Any("total_assets", stats["total_assets"]),
		zap.Any("total_relationships", stats["total_relationships"]),
		zap.Any("by_provider", stats["by_provider"]),
	)

	// In production, would persist to database here
	if !s.dryRun {
		s.logger.Info("Persisting assets to database (not implemented in prototype)")
		// TODO: Persist to database
	}
}

// pollEvents continuously fetches cloud events
func (s *IngestionService) pollEvents(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Event polling stopped")
			return
		case <-ticker.C:
			s.fetchEvents(ctx)
		}
	}
}

// fetchEvents fetches events from all cloud providers
func (s *IngestionService) fetchEvents(ctx context.Context) {
	s.logger.Info("Fetching cloud events...")

	// Fetch events from last polling interval
	endTime := time.Now()
	startTime := endTime.Add(-*eventPollInterval)

	events, err := s.manager.FetchAllEvents(ctx, startTime, endTime)
	if err != nil {
		s.logger.Error("Failed to fetch events", zap.Error(err))
		return
	}

	s.logger.Info("Events fetched",
		zap.Int("count", len(events)),
		zap.Time("start_time", startTime),
		zap.Time("end_time", endTime),
	)

	// Analyze events
	s.analyzeEvents(events)

	// In production, would persist to database here
	if !s.dryRun {
		s.logger.Info("Persisting events to database (not implemented in prototype)")
		// TODO: Persist to database
	}
}

// analyzeEvents performs basic event analysis
func (s *IngestionService) analyzeEvents(events []*cloud.CloudEvent) {
	if len(events) == 0 {
		return
	}

	// Count by provider
	byProvider := make(map[cloud.Provider]int)
	byType := make(map[string]int)
	failedEvents := 0

	for _, event := range events {
		byProvider[event.Provider]++
		byType[event.EventType]++
		if !event.Success {
			failedEvents++
		}
	}

	s.logger.Info("Event analysis",
		zap.Any("by_provider", byProvider),
		zap.Any("by_type", byType),
		zap.Int("failed_events", failedEvents),
	)

	// Alert on high failure rate
	if failedEvents > len(events)/2 {
		s.logger.Warn("High event failure rate detected",
			zap.Int("failed", failedEvents),
			zap.Int("total", len(events)),
			zap.Float64("rate", float64(failedEvents)/float64(len(events))),
		)
	}
}
