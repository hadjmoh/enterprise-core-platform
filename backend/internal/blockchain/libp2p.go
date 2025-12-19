package blockchain

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type LibP2PCollector struct {
	logPath string
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewLibP2PCollector(logPath string, logger *logger.Logger) *LibP2PCollector {
	return &LibP2PCollector{
		logPath: logPath,
		logger:  logger,
	}
}

func (l *LibP2PCollector) Start() error {
	l.ctx, l.cancel = context.WithCancel(context.Background())

	go l.collectEvents()

	l.logger.Info("LibP2P Collector started", "path", l.logPath)
	return nil
}

func (l *LibP2PCollector) collectEvents() {
	file, err := os.Open(l.logPath)
	if err != nil {
		l.logger.Error("Failed to open LibP2P log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-l.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "peer") || strings.Contains(line, "dht") || strings.Contains(line, "pubsub") {
					l.logger.Info("LibP2P Event", "event", line[:min(len(line), 100)])
					// TODO: Parse and send to pipeline
				}
			}
		}
	}
}

func (l *LibP2PCollector) Stop() {
	if l.cancel != nil {
		l.cancel()
		l.logger.Info("LibP2P Collector stopped")
	}
}

func (l *LibP2PCollector) Protocol() string {
	return "LibP2P"
}

