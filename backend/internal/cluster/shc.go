package cluster

import (
	"context"
	"enterprise-core/backend/internal/proto"
	"enterprise-core/backend/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

// StateStore represents a local cache of distributed search head state
type StateStore struct {
	data map[string]map[string][]byte
	mu   sync.RWMutex
}

func NewStateStore() *StateStore {
	return &StateStore{
		data: make(map[string]map[string][]byte),
	}
}

func (s *StateStore) Set(msgType, key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[msgType]; !ok {
		s.data[msgType] = make(map[string][]byte)
	}
	s.data[msgType][key] = value
}

func (s *StateStore) Get(msgType, key string) ([]byte, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if m, ok := s.data[msgType]; ok {
		val, ok := m[key]
		return val, ok
	}
	return nil, false
}

func (s *StateStore) GetAll(msgType string) map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make(map[string][]byte)
	if m, ok := s.data[msgType]; ok {
		for k, v := range m {
			res[k] = v
		}
	}
	return res
}

// SHCManager orchestrates state synchronization between search heads
type SHCManager struct {
	nodeID     string
	nodeMgr    *NodeManager
	store      *StateStore
	logger     *logger.Logger
}

func NewSHCManager(nodeID string, nodeMgr *NodeManager, l *logger.Logger) *SHCManager {
	return &SHCManager{
		nodeID:  nodeID,
		nodeMgr: nodeMgr,
		store:   NewStateStore(),
		logger:  l,
	}
}

func (m *SHCManager) GetStore() *StateStore {
	return m.store
}

// ReplicateState pushes a state update to all other search heads
func (m *SHCManager) ReplicateState(ctx context.Context, msgType, key string, value []byte) {
	m.store.Set(msgType, key, value)
	
	nodes := m.nodeMgr.GetNodes()
	var searchHeads []Node
	for _, n := range nodes {
		if (n.Role == RoleSearchHead || n.Role == RoleAll) && n.ID != m.nodeID && n.Status == NodeOnline {
			searchHeads = append(searchHeads, n)
		}
	}

	if len(searchHeads) == 0 {
		return
	}

	m.logger.Debug("Replicating state update", "type", msgType, "key", key, "nodes", len(searchHeads))

	update := &proto.StateUpdate{
		Type:         msgType,
		Key:          key,
		ValueJSON:    value,
		OriginNodeID: m.nodeID,
	}

	for _, sh := range searchHeads {
		go m.syncWithNode(sh, update)
	}
}

func (m *SHCManager) syncWithNode(node Node, update *proto.StateUpdate) {
	conn, err := grpc.Dial(node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		m.logger.Error("Failed to connect for state sync", err, "node_id", node.ID)
		return
	}
	defer conn.Close()

	client := proto.NewClusterServiceClient(conn)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.SyncState(ctx, update)
	if err != nil {
		m.logger.Warn("Failed to sync state with node", "node_id", node.ID, "error", err)
		return
	}

	if !resp.Success {
		m.logger.Warn("Node rejected state sync", "node_id", node.ID, "reason", resp.Error)
	}
}

// HandleSyncUpdate processes an incoming state update from another node
func (m *SHCManager) HandleSyncUpdate(update *proto.StateUpdate) error {
	m.logger.Debug("Received state sync update", "type", update.Type, "key", update.Key, "origin", update.OriginNodeID)
	m.store.Set(update.Type, update.Key, update.ValueJSON)
	return nil
}
