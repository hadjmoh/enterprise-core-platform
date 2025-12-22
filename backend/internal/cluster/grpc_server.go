package cluster

import (
	"context"
	"enterprise-core/backend/internal/proto"
	"enterprise-core/backend/internal/query"
	"enterprise-core/backend/pkg/logger"
	"encoding/json"
	"time"
)

type GRPCServer struct {
	manager    *NodeManager
	dispatcher *query.Dispatcher
	shc        *SHCManager
	logger     *logger.Logger
}

func NewGRPCServer(mgr *NodeManager, disp *query.Dispatcher, shc *SHCManager, l *logger.Logger) *GRPCServer {
	return &GRPCServer{
		manager:    mgr,
		dispatcher: disp,
		shc:        shc,
		logger:     l,
	}
}

func (s *GRPCServer) Heartbeat(ctx context.Context, req *proto.HeartbeatRequest) (*proto.HeartbeatResponse, error) {
	s.manager.UpdateNode(&Node{
		ID:       req.NodeID,
		Address:  req.Address,
		Status:   NodeOnline,
		CPUUsage: req.CPULoad,
		MemUsage: req.MemUsage,
	})
	
	return &proto.HeartbeatResponse{
		Success:    true,
		ServerTime: time.Now().Unix(),
	}, nil
}

func (s *GRPCServer) RegisterNode(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	s.manager.UpdateNode(&Node{
		ID:      req.NodeID,
		Address: req.Address,
		Role:    NodeRole(req.Role),
		Status:  NodeOnline,
	})
	
	s.logger.Info("New node registered in cluster", "node_id", req.NodeID, "address", req.Address)
	
	return &proto.RegisterResponse{
		Accepted:  true,
		ClusterID: "global-cluster-01",
	}, nil
}

func (s *GRPCServer) DistributedSearch(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	s.logger.Debug("Received distributed search request", "request_id", req.RequestID, "query", req.Query)

	// Execute locally as internal search
	results, err := s.dispatcher.Execute(ctx, req.Query, "internal", nil)
	if err != nil {
		return &proto.SearchResponse{Error: err.Error()}, nil
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return &proto.SearchResponse{Error: "Failed to marshal results: " + err.Error()}, nil
	}

	return &proto.SearchResponse{
		ResultsJSON: resultsJSON,
	}, nil
}

func (s *GRPCServer) SyncState(ctx context.Context, req *proto.StateUpdate) (*proto.SyncResponse, error) {
	if s.shc == nil {
		return &proto.SyncResponse{Success: false, Error: "SHC not initialized"}, nil
	}

	err := s.shc.HandleSyncUpdate(req)
	if err != nil {
		return &proto.SyncResponse{Success: false, Error: err.Error()}, nil
	}

	return &proto.SyncResponse{Success: true}, nil
}
