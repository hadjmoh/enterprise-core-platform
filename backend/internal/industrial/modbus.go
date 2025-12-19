package industrial

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/pipeline"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
	"time"
)

type ModbusListener struct {
	port     int
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewModbusListener(port int, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *ModbusListener {
	return &ModbusListener{
		port:     port,
		pipeline: pipe,
		logger:   logger,
	}
}

func (m *ModbusListener) Start() error {
	m.ctx, m.cancel = context.WithCancel(context.Background())

	addr := fmt.Sprintf(":%d", m.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	m.logger.Info("Modbus TCP Listener started", "port", m.port)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-m.ctx.Done():
					return
				default:
					m.logger.Error("Modbus accept error", err)
					continue
				}
			}
			go m.handleConnection(conn)
		}
	}()

	return nil
}

func (m *ModbusListener) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	buf := make([]byte, 260) // Modbus TCP max frame size
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	m.logger.Info("Received Modbus frame",
		"remote", conn.RemoteAddr(),
		"size", n,
	)
	
	// Send to pipeline
	m.pipeline.Process(buffer.Event{
		Timestamp: time.Now().Format(time.RFC3339),
		Source:    fmt.Sprintf("modbus:%d", m.port),
		Data: map[string]interface{}{
			"raw_hex": fmt.Sprintf("%x", buf[:n]),
			"length":  n,
			"remote":  conn.RemoteAddr().String(),
		},
	})
}

func (m *ModbusListener) Stop() {
	if m.cancel != nil {
		m.cancel()
		m.logger.Info("Modbus Listener stopped")
	}
}

func (m *ModbusListener) Protocol() string {
	return "Modbus"
}
