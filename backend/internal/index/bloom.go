package index

import (
	"hash/fnv"
)

// BloomFilter provides a space-efficient way to check for the possible presence of a token.
type BloomFilter struct {
	Metadata BloomMetadata `json:"metadata"`
	Bitset   []uint64      `json:"bitset"`
	M        uint          `json:"m"`
	K        uint          `json:"k"`
}

type BloomMetadata struct {
	InsertedTokens     int     `json:"inserted_tokens"`
	FalsePositiveRate  float64 `json:"false_positive_rate"`
	MinTimestamp       string  `json:"min_timestamp"`
	MaxTimestamp       string  `json:"max_timestamp"`
	IndexVersion       int     `json:"index_version"`
}

func NewBloomFilter(n uint, p float64) *BloomFilter {
	// Simple formula for m and k
	// m = - (n * ln(p)) / (ln(2)^2)
	// k = (m / n) * ln(2)
	// For 10,000 events, 1% false positive: m=95850 bits (~12KB), k=7 hashes
	m := uint(95850)
	k := uint(7)

	return &BloomFilter{
		Metadata: BloomMetadata{
			FalsePositiveRate: p,
			IndexVersion:      1,
		},
		Bitset: make([]uint64, (m+63)/64),
		M:      m,
		K:      k,
	}
}

func (bf *BloomFilter) Add(token string) {
	bf.Metadata.InsertedTokens++
	h1, h2 := bf.hash(token)
	for i := uint(0); i < bf.K; i++ {
		bit := (h1 + i*h2) % bf.M
		bf.Bitset[bit/64] |= (1 << (bit % 64))
	}
}

func (bf *BloomFilter) UpdateRange(timestamp string) {
	if bf.Metadata.MinTimestamp == "" || timestamp < bf.Metadata.MinTimestamp {
		bf.Metadata.MinTimestamp = timestamp
	}
	if bf.Metadata.MaxTimestamp == "" || timestamp > bf.Metadata.MaxTimestamp {
		bf.Metadata.MaxTimestamp = timestamp
	}
}

func (bf *BloomFilter) Test(token string) bool {
	h1, h2 := bf.hash(token)
	for i := uint(0); i < bf.K; i++ {
		bit := (h1 + i*h2) % bf.M
		if (bf.Bitset[bit/64] & (1 << (bit % 64))) == 0 {
			return false
		}
	}
	return true
}

func (bf *BloomFilter) hash(token string) (uint, uint) {
	h := fnv.New64a()
	h.Write([]byte(token))
	v := h.Sum64()
	return uint(v >> 32), uint(v & 0xFFFFFFFF)
}
