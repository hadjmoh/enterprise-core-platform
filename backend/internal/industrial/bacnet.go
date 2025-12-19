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

type BACnetCollector struct {
	port     int
	pipeline *pipeline.IngestionPipeline
	logger   *logger.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	conn     *net.UDPConn
}

func NewBACnetCollector(port int, pipe *pipeline.IngestionPipeline, logger *logger.Logger) *BACnetCollector {
	return &BACnetCollector{
		port:     port,
		pipeline: pipe,
		logger:   logger,
	}
}

func (b *BACnetCollector) Start() error {
	b.ctx, b.cancel = context.WithCancel(context.Background())

	addr := fmt.Sprintf(":%d", b.port)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	b.conn = conn

	b.logger.Info("BACnet Collector started", "port", b.port)

	go b.collect()
	return nil
}

func (b *BACnetCollector) collect() {
	buf := make([]byte, 1500) // BACnet max packet size
	for {
		n, remoteAddr, err := b.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-b.ctx.Done():
				return
			default:
				b.logger.Error("BACnet read error", err)
				continue
			}
		}

		b.logger.Info("Received BACnet packet",
			"from", remoteAddr,
			"size", n,
		)
		
		// Send to pipeline
		b.pipeline.Process(buffer.Event{
			Timestamp: time.Now().Format(time.RFC3339),
			Source:    fmt.Sprintf("bacnet:%d", b.port),
			Data: map[string]interface{}{
				"remote":  remoteAddr.String(),
				"size":    n,
				"raw_hex": fmt.Sprintf("%x", buf[:n]),
			},
		})
	}
}

func (b *BACnetCollector) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
	if b.conn != nil {
		b.conn.Close()
		b.logger.Info("BACnet Collector stopped")
	}
}

func (b *BACnetCollector) Protocol() string {
	return "BACnet"
}
