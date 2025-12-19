package network

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
)

type SCTPListener struct {
	port    int
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewSCTPListener(port int, logger *logger.Logger) *SCTPListener {
	return &SCTPListener{
		port:   port,
		logger: logger,
	}
}

func (s *SCTPListener) Start() error {
	s.ctx, s.cancel = context.WithCancel(context.Background())

	addr := fmt.Sprintf(":%d", s.port)
	laddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		return err
	}

	s.logger.Info("SCTP Listener started", "port", s.port)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-s.ctx.Done():
					return
				default:
					s.logger.Error("SCTP accept error", err)
					continue
				}
			}
			go s.handleConnection(conn)
		}
	}()

	return nil
}

func (s *SCTPListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		s.logger.Info("Received SCTP data", "bytes", n, "remote", conn.RemoteAddr())
		// TODO: Send to ingestion pipeline
	}
}

func (s *SCTPListener) Stop() {
	if s.cancel != nil {
		s.cancel()
		s.logger.Info("SCTP Listener stopped")
	}
}

func (s *SCTPListener) Protocol() string {
	return "SCTP"
}
