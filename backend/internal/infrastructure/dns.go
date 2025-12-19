package infrastructure

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type DNSLogParser struct {
	logPath string
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewDNSLogParser(logPath string, logger *logger.Logger) *DNSLogParser {
	return &DNSLogParser{
		logPath: logPath,
		logger:  logger,
	}
}

func (d *DNSLogParser) Start() error {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	go d.parseLogs()

	d.logger.Info("DNS Log Parser started", "path", d.logPath)
	return nil
}

func (d *DNSLogParser) parseLogs() {
	file, err := os.Open(d.logPath)
	if err != nil {
		d.logger.Error("Failed to open DNS log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "query") {
					d.logger.Info("DNS Query detected", "line", line[:min(len(line), 100)])
					// TODO: Parse query and send to pipeline
				}
			}
		}
	}
}

func (d *DNSLogParser) Stop() {
	if d.cancel != nil {
		d.cancel()
		d.logger.Info("DNS Log Parser stopped")
	}
}

func (d *DNSLogParser) Protocol() string {
	return "DNS"
}

