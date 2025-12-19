package infrastructure

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type LDAPCollector struct {
	logPath string
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewLDAPCollector(logPath string, logger *logger.Logger) *LDAPCollector {
	return &LDAPCollector{
		logPath: logPath,
		logger:  logger,
	}
}

func (l *LDAPCollector) Start() error {
	l.ctx, l.cancel = context.WithCancel(context.Background())

	go l.collectLogs()

	l.logger.Info("LDAP Collector started", "path", l.logPath)
	return nil
}

func (l *LDAPCollector) collectLogs() {
	file, err := os.Open(l.logPath)
	if err != nil {
		l.logger.Error("Failed to open LDAP log", err)
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
				if strings.Contains(line, "BIND") || strings.Contains(line, "SEARCH") {
					l.logger.Info("LDAP Operation", "op", line[:min(len(line), 100)])
					// TODO: Parse and send to pipeline
				}
			}
		}
	}
}

func (l *LDAPCollector) Stop() {
	if l.cancel != nil {
		l.cancel()
		l.logger.Info("LDAP Collector stopped")
	}
}

func (l *LDAPCollector) Protocol() string {
	return "LDAP"
}
