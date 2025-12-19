package custom

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
	"time"
)

// ProtocolHandler defines the interface for custom protocol handlers
type ProtocolHandler interface {
	Parse(data []byte) (interface{}, error)
	Name() string
}

// CustomProtocolListener provides a generic TCP listener for proprietary protocols
type CustomProtocolListener struct {
	port    int
	handler ProtocolHandler
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewCustomProtocolListener(port int, handler ProtocolHandler, logger *logger.Logger) *CustomProtocolListener {
	return &CustomProtocolListener{
		port:    port,
		handler: handler,
		logger:  logger,
	}
}

func (c *CustomProtocolListener) Start() error {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	addr := fmt.Sprintf(":%d", c.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	c.logger.Info("Custom Protocol Listener started",
		"protocol", c.handler.Name(),
		"port", c.port,
	)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-c.ctx.Done():
					return
				default:
					c.logger.Error("Custom protocol accept error", err)
					continue
				}
			}
			go c.handleConnection(conn)
		}
	}()

	return nil
}

func (c *CustomProtocolListener) handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	parsed, err := c.handler.Parse(buf[:n])
	if err != nil {
		c.logger.Error("Failed to parse custom protocol", err)
		return
	}

	c.logger.Info("Received custom protocol data",
		"protocol", c.handler.Name(),
		"parsed", parsed,
	)
	// TODO: Send to ingestion pipeline
}

func (c *CustomProtocolListener) Stop() {
	if c.cancel != nil {
		c.cancel()
		c.logger.Info("Custom Protocol Listener stopped", "protocol", c.handler.Name())
	}
}

func (c *CustomProtocolListener) Protocol() string {
	return "Custom-" + c.handler.Name()
}

// Example: Simple binary protocol handler
type SimpleBinaryHandler struct{}

func (s *SimpleBinaryHandler) Parse(data []byte) (interface{}, error) {
	// Example: just return hex representation
	return fmt.Sprintf("Binary[%d bytes]", len(data)), nil
}

func (s *SimpleBinaryHandler) Name() string {
	return "SimpleBinary"
}
