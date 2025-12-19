package network

import (
	"context"
	"crypto/tls"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
)

// TLSListener wraps any TCP listener with TLS encryption
type TLSListener struct {
	port      int
	certFile  string
	keyFile   string
	logger    *logger.Logger
	ctx       context.Context
	cancel    context.CancelFunc
	tlsConfig *tls.Config
}

func NewTLSListener(port int, certFile, keyFile string, logger *logger.Logger) (*TLSListener, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS cert: %w", err)
	}

	return &TLSListener{
		port:     port,
		certFile: certFile,
		keyFile:  keyFile,
		logger:   logger,
		tlsConfig: &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		},
	}, nil
}

func (t *TLSListener) Start() error {
	t.ctx, t.cancel = context.WithCancel(context.Background())

	addr := fmt.Sprintf(":%d", t.port)
	listener, err := tls.Listen("tcp", addr, t.tlsConfig)
	if err != nil {
		return err
	}

	t.logger.Info("TLS Listener started", "port", t.port)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-t.ctx.Done():
					return
				default:
					t.logger.Error("TLS accept error", err)
					continue
				}
			}
			go t.handleConnection(conn)
		}
	}()

	return nil
}

func (t *TLSListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 65536)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}

		t.logger.Info("Received TLS data", "bytes", n, "remote", conn.RemoteAddr())
		// TODO: Send to ingestion pipeline
	}
}

func (t *TLSListener) Stop() {
	if t.cancel != nil {
		t.cancel()
		t.logger.Info("TLS Listener stopped")
	}
}

func (t *TLSListener) Protocol() string {
	return "TLS"
}
