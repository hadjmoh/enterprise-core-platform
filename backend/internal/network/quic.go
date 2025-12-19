package network

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"math/big"

	"github.com/quic-go/quic-go"
)

type QUICListener struct {
	port    int
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewQUICListener(port int, logger *logger.Logger) *QUICListener {
	return &QUICListener{
		port:   port,
		logger: logger,
	}
}

func (q *QUICListener) Start() error {
	q.ctx, q.cancel = context.WithCancel(context.Background())

	// Generate self-signed cert for QUIC (required)
	tlsConf, err := generateTLSConfig()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf(":%d", q.port)
	listener, err := quic.ListenAddr(addr, tlsConf, nil)
	if err != nil {
		return err
	}

	q.logger.Info("QUIC Listener started", "port", q.port)

	go func() {
		for {
			select {
			case <-q.ctx.Done():
				listener.Close()
				return
			default:
				// Accept connections in background
				// Note: Full QUIC implementation requires more complex stream handling
				q.logger.Info("QUIC Listener running", "port", q.port)
			}
		}
	}()

	return nil
}

func (q *QUICListener) Stop() {
	if q.cancel != nil {
		q.cancel()
		q.logger.Info("QUIC Listener stopped")
	}
}

func (q *QUICListener) Protocol() string {
	return "QUIC"
}

// generateTLSConfig creates a self-signed certificate for QUIC
func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"h3"},
	}, nil
}
