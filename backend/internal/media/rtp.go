package media

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type RTPCollector struct {
	logPath string
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewRTPCollector(logPath string, logger *logger.Logger) *RTPCollector {
	return &RTPCollector{
		logPath: logPath,
		logger:  logger,
	}
}

func (r *RTPCollector) Start() error {
	r.ctx, r.cancel = context.WithCancel(context.Background())

	go r.collectLogs()

	r.logger.Info("RTP/RTSP Collector started", "path", r.logPath)
	return nil
}

func (r *RTPCollector) collectLogs() {
	file, err := os.Open(r.logPath)
	if err != nil {
		r.logger.Error("Failed to open RTP log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "RTP") || strings.Contains(line, "RTSP") || strings.Contains(line, "stream") {
					r.logger.Info("Media Stream Event", "event", line[:min(len(line), 100)])
					// TODO: Parse media event and send to pipeline
				}
			}
		}
	}
}

func (r *RTPCollector) Stop() {
	if r.cancel != nil {
		r.cancel()
		r.logger.Info("RTP Collector stopped")
	}
}

func (r *RTPCollector) Protocol() string {
	return "RTP/RTSP"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
