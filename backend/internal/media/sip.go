package media

import (
	"bufio"
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type SIPCollector struct {
	logPath  string
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewSIPCollector(logPath string, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *SIPCollector {
	return &SIPCollector{
		logPath:  logPath,
		pipeline: pipe,
		logger:   logger,
	}
}

func (s *SIPCollector) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	go s.collectLogs()

	s.logger.Info("SIP Collector started", "path", s.logPath)
	return nil
}

func (s *SIPCollector) collectLogs() {
	file, err := os.Open(s.logPath)
	if err != nil {
		s.logger.Error("Failed to open SIP log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "INVITE") || strings.Contains(line, "BYE") || strings.Contains(line, "REGISTER") {
					s.logger.Info("SIP Call Event", "event", line[:min(len(line), 100)])
					
					// Send to pipeline
					s.pipeline.Process(buffer.Event{
						Timestamp: time.Now().Format(time.RFC3339),
						Source:    "media:sip",
						Data: map[string]interface{}{
							"log_path": s.logPath,
							"event":    line,
						},
					})
				}
			}
		}
	}
}

func (s *SIPCollector) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.logger.Info("SIP Collector stopped")
	}
}

func (s *SIPCollector) Protocol() string {
	return "SIP"
}

