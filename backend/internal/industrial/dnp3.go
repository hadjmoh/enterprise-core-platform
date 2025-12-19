package industrial

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

type DNP3Parser struct {
	logPath  string
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewDNP3Parser(logPath string, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *DNP3Parser {
	return &DNP3Parser{
		logPath:  logPath,
		pipeline: pipe,
		logger:   logger,
	}
}

func (d *DNP3Parser) Start() error {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	go d.parseLogs()

	d.logger.Info("DNP3 Parser started", "path", d.logPath)
	return nil
}

func (d *DNP3Parser) parseLogs() {
	file, err := os.Open(d.logPath)
	if err != nil {
		d.logger.Error("Failed to open DNP3 log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "DNP3") || strings.Contains(line, "outstation") {
					d.logger.Info("DNP3 Event", "event", line[:min(len(line), 100)])
					
					// Send to pipeline
					d.pipeline.Process(buffer.Event{
						Timestamp: time.Now().Format(time.RFC3339),
						Source:    "industrial:dnp3",
						Data: map[string]interface{}{
							"log_path": d.logPath,
							"event":    line,
						},
					})
				}
			}
		}
	}
}

func (d *DNP3Parser) Stop() {
	if d.cancel != nil {
		d.cancel()
		d.logger.Info("DNP3 Parser stopped")
	}
}

func (d *DNP3Parser) Protocol() string {
	return "DNP3"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
