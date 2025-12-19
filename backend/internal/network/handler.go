package network

import (
	"enterprise-core/backend/internal/api"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"strings"
	"time"
)

type SyslogHandler struct {
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
}

func NewSyslogHandler(pipe *pipeline.IngestionPipeline, logger *logger.Logger) *SyslogHandler {
	return &SyslogHandler{
		pipeline: pipe,
		logger:   logger,
	}
}

func (h *SyslogHandler) Handle(data []byte, meta map[string]string) error {
	raw := string(data)
	
	// Basic RFC5424/RFC3164 heuristic parsing (MVP)
	// In production, use a proper parser library like go-syslog
	
	event := api.Event{
		Time:   time.Now().Unix(),
		Host:   meta["remote_addr"],
		Source: meta["proto"] + ":514",
		Index:  "default",
		Event:  raw,
		Fields: make(map[string]interface{}),
	}

	// Simple priority extraction <PRI>
	if strings.HasPrefix(raw, "<") {
		end := strings.Index(raw, ">")
		if end > 0 {
			event.Fields["priority"] = raw[1:end]
		}
	}

	h.logger.Info("Ingested Syslog", "source", event.Source, "length", len(raw))
	
	// Send to Pipeline
	return h.pipeline.Process(buffer.Event{
		Timestamp: time.Unix(event.Time, 0).Format(time.RFC3339),
		Source:    event.Source,
		Data:      event.Fields, // Basic fields
	})
}
