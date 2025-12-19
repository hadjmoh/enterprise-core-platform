package blockchain

import (
	"bufio"
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"os"
	"strings"
	"time"
)

type CryptoNodeCollector struct {
	logPath  string
	nodeType string // "bitcoin" or "ethereum"
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewCryptoNodeCollector(logPath, nodeType string, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *CryptoNodeCollector {
	return &CryptoNodeCollector{
		logPath:  logPath,
		nodeType: nodeType,
		pipeline: pipe,
		logger:   logger,
	}
}

func (c *CryptoNodeCollector) Start() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	go c.collectLogs()

	c.logger.Info("Crypto Node Collector started", "type", c.nodeType, "path", c.logPath)
	return nil
}

func (c *CryptoNodeCollector) collectLogs() {
	file, err := os.Open(c.logPath)
	if err != nil {
		c.logger.Error("Failed to open crypto node log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if c.isRelevantEvent(line) {
					c.logger.Info("Crypto Node Event",
						"type", c.nodeType,
						"event", line[:min(len(line), 120)],
					)
					
					// Send to pipeline
					c.pipeline.Process(buffer.Event{
						Timestamp: time.Now().Format(time.RFC3339),
						Source:    fmt.Sprintf("blockchain:%s", c.nodeType),
						Data: map[string]interface{}{
							"node_type": c.nodeType,
							"event":     line,
						},
					})
				}
			}
		}
	}
}

func (c *CryptoNodeCollector) isRelevantEvent(line string) bool {
	keywords := []string{"block", "transaction", "peer", "mining", "sync", "rpc"}
	for _, kw := range keywords {
		if strings.Contains(strings.ToLower(line), kw) {
			return true
		}
	}
	return false
}

func (c *CryptoNodeCollector) Stop() {
	if c.cancel != nil {
		c.cancel()
		c.logger.Info("Crypto Node Collector stopped", "type", c.nodeType)
	}
}

func (c *CryptoNodeCollector) Protocol() string {
	return "Crypto-" + c.nodeType
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
