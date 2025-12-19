package ha

import (
	"context"
	"enterprise-core/backend/pkg/logger"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// LeaderElection handles cluster leadership coordination
type LeaderElection struct {
	client     *clientv3.Client
	session    *concurrency.Session
	election   *concurrency.Election
	logger     *logger.Logger
	mu         sync.RWMutex
	isLeader   bool
	nodeID     string
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewLeaderElection(etcdEndpoints []string, nodeID string, logger *logger.Logger) (*LeaderElection, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdEndpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &LeaderElection{
		client: cli,
		nodeID: nodeID,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (l *LeaderElection) Run() error {
	sess, err := concurrency.NewSession(l.client, concurrency.WithTTL(10))
	if err != nil {
		return err
	}
	l.session = sess

	election := concurrency.NewElection(sess, "/enterprise-core/leader")
	l.election = election

	go l.observeLeader()

	l.logger.Info("Attempting to acquire leadership", "nodeID", l.nodeID)
	if err := election.Campaign(l.ctx, l.nodeID); err != nil {
		return err
	}

	l.mu.Lock()
	l.isLeader = true
	l.mu.Unlock()

	l.logger.Info("Acquired cluster leadership", "nodeID", l.nodeID)
	return nil
}

func (l *LeaderElection) observeLeader() {
	observeChan := l.election.Observe(l.ctx)
	for {
		select {
		case <-l.ctx.Done():
			return
		case resp := <-observeChan:
			if len(resp.Kvs) > 0 {
				currentLeader := string(resp.Kvs[0].Value)
				l.logger.Info("Current cluster leader", "leader", currentLeader)
				
				l.mu.Lock()
				l.isLeader = (currentLeader == l.nodeID)
				l.mu.Unlock()
			}
		}
	}
}

func (l *LeaderElection) IsLeader() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.isLeader
}

func (l *LeaderElection) Resign() error {
	l.cancel()
	if l.election != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return l.election.Resign(ctx)
	}
	if l.session != nil {
		return l.session.Close()
	}
	return nil
}

func (l *LeaderElection) Close() error {
	l.cancel()
	if l.client != nil {
		return l.client.Close()
	}
	return nil
}
