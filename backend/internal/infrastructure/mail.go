package infrastructure

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type MailCollector struct {
	logPaths []string // SMTP, IMAP, POP3 logs
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewMailCollector(logPaths []string, logger *logger.Logger) *MailCollector {
	return &MailCollector{
		logPaths: logPaths,
		logger:   logger,
	}
}

func (m *MailCollector) Start() error {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	for _, path := range m.logPaths {
		go m.collectMailLog(path)
	}

	m.logger.Info("Mail Collector started", "logs", len(m.logPaths))
	return nil
}

func (m *MailCollector) collectMailLog(logPath string) {
	file, err := os.Open(logPath)
	if err != nil {
		m.logger.Error("Failed to open mail log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "from=") || strings.Contains(line, "to=") || strings.Contains(line, "LOGIN") {
					m.logger.Info("Mail Event", "log", logPath, "event", line[:min(len(line), 100)])
					// TODO: Parse and send to pipeline
				}
			}
		}
	}
}

func (m *MailCollector) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.logger.Info("Mail Collector stopped")
	}
}

func (m *MailCollector) Protocol() string {
	return "Mail"
}
