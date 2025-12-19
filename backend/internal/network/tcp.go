package network

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
	"time"
)

type TCPListener struct {
	port    int
	handler Handler
	logger  *logger.Logger
	listener net.Listener
}

func NewTCPListener(port int, handler Handler, logger *logger.Logger) *TCPListener {
	return &TCPListener{
		port:    port,
		handler: handler,
		logger:  logger,
	}
}

func (t *TCPListener) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", t.port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	t.listener = l
	t.logger.Info("TCP Listener started", "port", t.port)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					t.logger.Error("TCP Accept error", err)
					continue
				}
			}
			go t.handleConnection(conn)
		}
	}()

	return nil
}

func (t *TCPListener) handleConnection(conn net.Conn) {
	defer conn.Close()
	// Set read deadline to avoid leaking connections
	conn.SetReadDeadline(time.Now().Add(5 * time.Minute))

	scanner := bufio.NewScanner(conn)
	// Default split is ScanLines, suitable for Syslog-TCP
	for scanner.Scan() {
		data := scanner.Bytes()
		meta := map[string]string{
			"remote_addr": conn.RemoteAddr().String(),
			"proto":       "tcp",
		}
		if err := t.handler.Handle(data, meta); err != nil {
			t.logger.Error("Error handling TCP data", err)
		}
	}
}

func (t *TCPListener) Stop() error {
	if t.listener != nil {
		return t.listener.Close()
	}
	return nil
}

func (t *TCPListener) Protocol() string {
	return "TCP"
}

func (t *TCPListener) Port() int {
	return t.port
}
