package cluster

import (
	"context"
	"enterprise-core/backend/internal/buffer"
	"enterprise-core/backend/internal/proto"
	"enterprise-core/backend/pkg/logger"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

// MapReduceEngine coordinates distributed search execution
type MapReduceEngine struct {
	manager      *NodeManager
	shardMgr     *ShardManager
	loadBalancer *LoadBalancer
	logger       *logger.Logger
}

func NewMapReduceEngine(mgr *NodeManager, shardMgr *ShardManager, lb *LoadBalancer, l *logger.Logger) *MapReduceEngine {
	return &MapReduceEngine{
		manager:      mgr,
		shardMgr:     shardMgr,
		loadBalancer: lb,
		logger:       l,
	}
}

// ExecuteDistributed runs a query across all available indexers
func (e *MapReduceEngine) ExecuteDistributed(ctx context.Context, queryStr string, limit int) ([]buffer.Event, error) {
	nodes := e.manager.GetNodes()
	var indexers []Node
	for _, n := range nodes {
		if (n.Role == RoleIndexer || n.Role == RoleAll) && n.Status == NodeOnline {
			indexers = append(indexers, n)
		}
	}

	if len(indexers) == 0 {
		e.logger.Warn("No available indexers for distributed search")
		return []buffer.Event{}, nil
	}

	// Use load balancer to select best nodes
	selectedNodes := e.loadBalancer.SelectNodes(indexers, len(indexers))
	
	e.logger.Info("Dispatching distributed search", 
		"total_indexers", len(indexers), 
		"selected_nodes", len(selectedNodes), 
		"query", queryStr)

	var wg sync.WaitGroup
	resultsChan := make(chan SearchResult, len(selectedNodes))

	for _, indexer := range selectedNodes {
		wg.Add(1)
		go func(node Node) {
			defer wg.Done()
			events, err := e.searchOnNode(ctx, node, queryStr)
			res := SearchResult{NodeID: node.ID, Events: events}
			if err != nil {
				e.logger.Error("Search failed on node", err, "node_id", node.ID)
				res.Error = err.Error()
			}
			resultsChan <- res
		}(indexer)
	}

	wg.Wait()
	close(resultsChan)

	var partialResults []SearchResult
	for res := range resultsChan {
		partialResults = append(partialResults, res)
	}

	// Reduce phase: merge results
	merger := NewResultMerger(limit)
	return merger.Merge(partialResults), nil
}

func (e *MapReduceEngine) searchOnNode(ctx context.Context, node Node, queryStr string) ([]buffer.Event, error) {
	conn, err := grpc.Dial(node.Address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := proto.NewClusterServiceClient(conn)
	
	// Use a slightly shorter timeout for the network call than the parent context
	searchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := client.DistributedSearch(searchCtx, &proto.SearchRequest{
		Query:     queryStr,
		RequestID: "dist-" + time.Now().Format("150405"),
	})

	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}

	return UnmarshalEvents(resp.ResultsJSON)
}
