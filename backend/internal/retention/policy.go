package retention

import (
	"enterprise-core/backend/pkg/logger"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Policy defines data retention rules
type Policy struct {
	DataDir      string
	Retention    time.Duration
	AuditDir     string
	AuditReten   time.Duration
	Interval     time.Duration
	logger       *logger.Logger
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

func NewPolicy(dataDir string, retention time.Duration, auditDir string, auditReten time.Duration, logger *logger.Logger) *Policy {
	return &Policy{
		DataDir:    dataDir,
		Retention:  retention,
		AuditDir:   auditDir,
		AuditReten: auditReten,
		Interval:   24 * time.Hour,
		logger:     logger,
		stopChan:   make(chan struct{}),
	}
}

func (p *Policy) Start() {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.cleanAll() // Run immediately on start
		
		ticker := time.NewTicker(p.Interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				p.cleanAll()
			case <-p.stopChan:
				return
			}
		}
	}()
}

func (p *Policy) cleanAll() {
	p.logger.Info("Starting data retention cleanup")
	p.cleanDir(p.DataDir, p.Retention)
	p.cleanDir(p.AuditDir, p.AuditReten)
}

func (p *Policy) cleanDir(dir string, retention time.Duration) {
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && time.Since(info.ModTime()) > retention {
			p.logger.Info("Retiring old file", "path", path, "age", time.Since(info.ModTime()))
			return os.Remove(path)
		}
		return nil
	})
	if err != nil {
		p.logger.Error("Error during retention cleanup", err, "dir", dir)
	}
}

func (p *Policy) Stop() {
	close(p.stopChan)
	p.wg.Wait()
}
