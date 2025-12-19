package infrastructure

import (
	"bufio"
	"context"
	"enterprise-core/backend/pkg/logger"
	"os"
	"strings"
	"time"
)

type DHCPMonitor struct {
	logPath string
	logger  *logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewDHCPMonitor(logPath string, logger *logger.Logger) *DHCPMonitor {
	return &DHCPMonitor{
		logPath: logPath,
		logger:  logger,
	}
}

func (d *DHCPMonitor) Start() error {
	d.ctx, d.cancel = context.WithCancel(context.Background())

	go d.monitorLeases()

	d.logger.Info("DHCP Monitor started", "path", d.logPath)
	return nil
}

func (d *DHCPMonitor) monitorLeases() {
	file, err := os.Open(d.logPath)
	if err != nil {
		d.logger.Error("Failed to open DHCP log", err)
		return
	}
	defer file.Close()

	file.Seek(0, os.SEEK_END)
	scanner := bufio.NewScanner(file)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "DHCPACK") || strings.Contains(line, "DHCPREQUEST") {
					d.logger.Info("DHCP Event", "event", line[:min(len(line), 80)])
					// TODO: Parse and send to pipeline
				}
			}
		}
	}
}

func (d *DHCPMonitor) Stop() {
	if d.cancel != nil {
		d.cancel()
		d.logger.Info("DHCP Monitor stopped")
	}
}

func (d *DHCPMonitor) Protocol() string {
	return "DHCP"
}
