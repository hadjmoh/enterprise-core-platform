package storage

import (
	"hash/fnv"
)

// ShardDistributor determines which shard an event belongs to.
type ShardDistributor struct {
	ShardCount int
}

func NewShardDistributor(count int) *ShardDistributor {
	if count <= 0 {
		count = 8 // Default
	}
	return &ShardDistributor{ShardCount: count}
}

// GetShard returns a shard index (0 to ShardCount-1) for a given key.
func (d *ShardDistributor) GetShard(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() % uint32(d.ShardCount))
}
