package infrastructure

import (
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
	"time"

	"github.com/gosnmp/gosnmp"
)

type SNMPTrapReceiver struct {
	port     uint16
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	server   *gosnmp.TrapListener
}

func NewSNMPTrapReceiver(port uint16, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *SNMPTrapReceiver {
	return &SNMPTrapReceiver{
		port:     port,
		pipeline: pipe,
		logger:   logger,
	}
}

func (s *SNMPTrapReceiver) Start() error {
	s.server = gosnmp.NewTrapListener()
	s.server.OnNewTrap = s.handleTrap
	s.server.Params = gosnmp.Default

	addr := fmt.Sprintf("0.0.0.0:%d", s.port)
	
	s.logger.Info("SNMP Trap Receiver started", "port", s.port)
	
	go func() {
		if err := s.server.Listen(addr); err != nil {
			s.logger.Error("SNMP Trap Receiver error", err)
		}
	}()

	return nil
}

func (s *SNMPTrapReceiver) handleTrap(packet *gosnmp.SnmpPacket, addr *net.UDPAddr) {
	s.logger.Info("Received SNMP Trap",
		"from", addr.String(),
		"version", packet.Version,
		"community", packet.Community,
		"variables", len(packet.Variables),
	)
	
	// Send to pipeline
	s.pipeline.Process(buffer.Event{
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    fmt.Sprintf("snmp:%s", addr.String()),
		Data: map[string]interface{}{
			"version":   packet.Version.String(),
			"community": packet.Community,
			"variables": len(packet.Variables),
		},
	})
}

func (s *SNMPTrapReceiver) Stop() {
	if s.server != nil {
		s.server.Close()
		s.logger.Info("SNMP Trap Receiver stopped")
	}
}

func (s *SNMPTrapReceiver) Protocol() string {
	return "SNMP"
}
