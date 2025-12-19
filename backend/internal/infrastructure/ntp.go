package infrastructure

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type NTPCollector struct {
	logPath string
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewNTPCollector(logPath string, logger *logger.Logger) *NTPCollector {
	return &NTPCollector{
		logPath: logPath,
		logger:  logger,
	}
}

func (n *NTPCollector) Start() error {
	n.ctx, n.cancel = context.WithCancel(context.Background())

	go n.collectStats()

	n.logger.Info("NTP Collector started", "path", n.logPath)
	return nil
}

func (n *NTPCollector) collectStats() {
	file, err := os.Open(n.logPath)
	if err != nil {
		n.logger.Error("Failed to open NTP log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "synchronized") || strings.Contains(line, "offset") {
					n.logger.Info("NTP Stats", "stat", line[:min(len(line), 100)])
					// TODO: Parse and send to pipeline
				}
			}
		}
	}
}

func (n *NTPCollector) Stop() {
	if n.cancel != nil {
		n.cancel()
		n.logger.Info("NTP Collector stopped")
	}
}

func (n *NTPCollector) Protocol() string {
	return "NTP"
}
