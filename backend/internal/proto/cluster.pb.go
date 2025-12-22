package proto

import (
	"context"
	"google.golang.org/grpc"
)

// HeartbeatRequest mimics the proto message
type HeartbeatRequest struct {
	NodeID    string  `json:"node_id"`
	Address   string  `json:"address"`
	CPULoad   float32 `json:"cpu_load"`
	MemUsage  float32 `json:"mem_usage"`
	Timestamp int64   `json:"timestamp"`
}

// HeartbeatResponse mimics the proto response
type HeartbeatResponse struct {
	Success    bool  `json:"success"`
	ServerTime int64 `json:"server_time"`
}

// RegisterRequest mimics the proto message
type RegisterRequest struct {
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
	Role    string `json:"role"`
}

// RegisterResponse mimics the proto response
type RegisterResponse struct {
	Accepted  bool   `json:"accepted"`
	ClusterID string `json:"cluster_id"`
}

// SearchRequest mimics the proto message
type SearchRequest struct {
	Query     string            `json:"query"`
	RequestID string            `json:"request_id"`
	Metadata  map[string]string `json:"metadata"`
}

// SearchResponse mimics the proto response
type SearchResponse struct {
	ResultsJSON []byte `json:"results_json"`
	Error       string `json:"error"`
}

// StateUpdate mimics the proto message
type StateUpdate struct {
	Type         string `json:"type"`
	Key          string `json:"key"`
	ValueJSON    []byte `json:"value_json"`
	OriginNodeID string `json:"origin_node_id"`
}

// SyncResponse mimics the proto response
type SyncResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ConfigRequest mimics the proto message
type ConfigRequest struct {
	Version  string            `json:"version"`
	Checksum string            `json:"checksum"`
	Content  map[string]string `json:"content"`
}

// ConfigResponse mimics the proto response
type ConfigResponse struct {
	Success bool   `json:"success"`
	Version string `json:"version"`
	Error   string `json:"error"`
}

// ClusterServiceClient is the client API
type ClusterServiceClient interface {
	Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error)
	RegisterNode(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error)
	DistributedSearch(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error)
	SyncState(ctx context.Context, in *StateUpdate, opts ...grpc.CallOption) (*SyncResponse, error)
	DeployConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error)
	GetConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error)
}

// ClusterServiceServer is the server API
type ClusterServiceServer interface {
	Heartbeat(context.Context, *HeartbeatRequest) (*HeartbeatResponse, error)
	RegisterNode(context.Context, *RegisterRequest) (*RegisterResponse, error)
	DistributedSearch(context.Context, *SearchRequest) (*SearchResponse, error)
	SyncState(context.Context, *StateUpdate) (*SyncResponse, error)
	DeployConfig(context.Context, *ConfigRequest) (*ConfigResponse, error)
	GetConfig(context.Context, *ConfigRequest) (*ConfigResponse, error)
}

// RegisterClusterServiceServer registers the server
func RegisterClusterServiceServer(s *grpc.Server, srv ClusterServiceServer) {
	s.RegisterService(&_ClusterService_serviceDesc, srv)
}

var _ClusterService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.ClusterService",
	HandlerType: (*ClusterServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Heartbeat",
			Handler:    nil, // Would be implemented in a real generation
		},
		{
			MethodName: "RegisterNode",
			Handler:    nil,
		},
		{
			MethodName: "DistributedSearch",
			Handler:    nil,
		},
		{
			MethodName: "SyncState",
			Handler:    nil,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cluster.proto",
}

// NewClusterServiceClient is a helper since we are mocking the pb file
func NewClusterServiceClient(cc *grpc.ClientConn) ClusterServiceClient {
	return &clusterServiceClient{cc}
}

type clusterServiceClient struct {
	cc *grpc.ClientConn
}

func (c *clusterServiceClient) Heartbeat(ctx context.Context, in *HeartbeatRequest, opts ...grpc.CallOption) (*HeartbeatResponse, error) {
	out := new(HeartbeatResponse)
	err := c.cc.Invoke(ctx, "/proto.ClusterService/Heartbeat", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) RegisterNode(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*RegisterResponse, error) {
	out := new(RegisterResponse)
	err := c.cc.Invoke(ctx, "/proto.ClusterService/RegisterNode", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) DistributedSearch(ctx context.Context, in *SearchRequest, opts ...grpc.CallOption) (*SearchResponse, error) {
	out := new(SearchResponse)
	err := c.cc.Invoke(ctx, "/proto.ClusterService/DistributedSearch", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) SyncState(ctx context.Context, in *StateUpdate, opts ...grpc.CallOption) (*SyncResponse, error) {
	out := new(SyncResponse)
	err := c.cc.Invoke(ctx, "/proto.ClusterService/SyncState", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) DeployConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error) {
	out := new(ConfigResponse)
	err := c.cc.Invoke(ctx, "/proto.ClusterService/DeployConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *clusterServiceClient) GetConfig(ctx context.Context, in *ConfigRequest, opts ...grpc.CallOption) (*ConfigResponse, error) {
	out := new(ConfigResponse)
	err := c.cc.Invoke(ctx, "/proto.ClusterService/GetConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}
