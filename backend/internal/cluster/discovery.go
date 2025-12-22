package cluster

import (
	"enterprise-core/backend/pkg/logger"
	"sync"
)

// PeerDiscovery handles node registration and peer lookup
type PeerDiscovery struct {
	manager *NodeManager
	peers   []string // Initial static peer list (addresses)
	logger  *logger.Logger
	mu      sync.RWMutex
}

func NewPeerDiscovery(mgr *NodeManager, staticPeers []string, l *logger.Logger) *PeerDiscovery {
	return &PeerDiscovery{
		manager: mgr,
		peers:   staticPeers,
		logger:  l,
	}
}

// Register self and discover neighbors
func (d *PeerDiscovery) Discover() {
	d.logger.Info("Starting cluster peer discovery", "static_peers", len(d.peers))
	
	for _, addr := range d.peers {
		// In a real implementation, we would probe these addresses
		// For Session 8.1, we'll just log the intent.
		d.logger.Debug("Probing static cluster peer", "address", addr)
	}
}

// GetPeerAddresses returns a list of candidate addresses for cluster join
func (d *PeerDiscovery) GetPeerAddresses() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.peers
}

// AddPeer dynamically adds a peer address to the discovery list
func (d *PeerDiscovery) AddPeer(addr string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	for _, existing := range d.peers {
		if existing == addr {
			return
		}
	}
	d.peers = append(d.peers, addr)
}
