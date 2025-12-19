package network

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"fmt"
	"net"
)

type UDPListener struct {
	port    int
	handler Handler
	logger  *logger.Logger
	conn    *net.UDPConn
}

func NewUDPListener(port int, handler Handler, logger *logger.Logger) *UDPListener {
	return &UDPListener{
		port:    port,
		handler: handler,
		logger:  logger,
	}
}

func (u *UDPListener) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", u.port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	u.conn = conn
	u.logger.Info("UDP Listener started", "port", u.port)

	go func() {
		buf := make([]byte, 65535) // Max UDP packet size
		for {
			n, remoteAddr, err := u.conn.ReadFromUDP(buf)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					u.logger.Error("UDP Read error", err)
					continue
				}
			}

			// Copy buffer to avoid race conditions if Handler is async
			payload := make([]byte, n)
			copy(payload, buf[:n])

			meta := map[string]string{
				"remote_addr": remoteAddr.String(),
				"proto":       "udp",
			}
			
			if err := u.handler.Handle(payload, meta); err != nil {
				u.logger.Error("Error handling UDP data", err)
			}
		}
	}()

	return nil
}

func (u *UDPListener) Stop() error {
	if u.conn != nil {
		return u.conn.Close()
	}
	return nil
}

func (u *UDPListener) Protocol() string {
	return "UDP"
}

func (u *UDPListener) Port() int {
	return u.port
}
