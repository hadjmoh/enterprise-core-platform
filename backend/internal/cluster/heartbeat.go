package cluster

import (
	"context"
	"enterprise-core/backend/internal/proto"
	"enterprise-core/backend/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type HeartbeatWorker struct {
	nodeID    string
	address   string
	discovery *PeerDiscovery
	manager   *NodeManager
	logger    *logger.Logger
	stop      chan struct{}
}

func NewHeartbeatWorker(nodeID, addr string, disc *PeerDiscovery, mgr *NodeManager, l *logger.Logger) *HeartbeatWorker {
	return &HeartbeatWorker{
		nodeID:    nodeID,
		address:   addr,
		discovery: disc,
		manager:   mgr,
		logger:    l,
		stop:      make(chan struct{}),
	}
}

func (w *HeartbeatWorker) Start() {
	ticker := time.NewTicker(5 * time.Second)
	w.logger.Info("Heartbeat worker started", "node_id", w.nodeID)
	
	go func() {
		for {
			select {
			case <-ticker.C:
				w.sendHeartbeats()
				w.manager.MarkOffline(15 * time.Second) // Clean up nodes not seen for 15s
			case <-w.stop:
				ticker.Stop()
				return
			}
		}
	}()
}

func (w *HeartbeatWorker) Stop() {
	close(w.stop)
}

func (w *HeartbeatWorker) sendHeartbeats() {
	peers := w.discovery.GetPeerAddresses()
	for _, peerAddr := range peers {
		if peerAddr == w.address {
			continue // Don't heartbeat self
		}
		
		go w.sendToPeer(peerAddr)
	}
}

func (w *HeartbeatWorker) sendToPeer(addr string) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		w.logger.Debug("Failed to connect to cluster peer", "address", addr, "error", err)
		return
	}
	defer conn.Close()

	client := proto.NewClusterServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err = client.Heartbeat(ctx, &proto.HeartbeatRequest{
		NodeID:    w.nodeID,
		Address:   w.address,
		CPULoad:   0.5, // Mock values for now
		MemUsage:  0.4,
		Timestamp: time.Now().Unix(),
	})

	if err != nil {
		w.logger.Debug("Heartbeat to peer failed", "address", addr, "error", err)
	} else {
		w.logger.Debug("Sent heartbeat to peer", "address", addr)
	}
}

// Heartbeat client implementation removed from here as it is now in cluster.pb.go
