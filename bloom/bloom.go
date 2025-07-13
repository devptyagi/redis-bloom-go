package bloom

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// BloomFilter represents the main interface for Bloom Filter operations
type BloomFilter interface {
	Add(data []byte) error
	Exists(data []byte) (bool, error)
}

// RedisClient interface abstracts both Redis single-node and cluster clients
type RedisClient interface {
	SetBit(ctx context.Context, key string, offset int64, value int) *redis.IntCmd
	GetBit(ctx context.Context, key string, offset int64) *redis.IntCmd
	Pipeline() pipeliner
}

// pipeliner is a minimal interface for pipelining, used for both production and test
// In production, it is satisfied by redis.Pipeliner; in tests, by the minimal mock
// This allows robust, testable code without mocking the full redis.Pipeliner interface
type pipeliner interface {
	SetBit(ctx context.Context, key string, offset int64, value int) *redis.IntCmd
	GetBit(ctx context.Context, key string, offset int64) *redis.IntCmd
	Exec(ctx context.Context) ([]redis.Cmder, error)
}

// bloomFilter implements the BloomFilter interface
type bloomFilter struct {
	config       Config
	bitSize      uint64
	hashCount    uint
	hashStrategy HashStrategy
}

// NewBloomFilter creates a new Bloom Filter instance with the given configuration
func NewBloomFilter(cfg Config) (BloomFilter, error) {
	if cfg.ExpectedInsertions == 0 {
		return nil, ErrInvalidExpectedInsertions
	}
	if cfg.FalsePositiveRate <= 0 || cfg.FalsePositiveRate >= 1 {
		return nil, ErrInvalidFalsePositiveRate
	}
	if cfg.RedisKey == "" {
		return nil, ErrEmptyRedisKey
	}
	if cfg.RedisClient == nil {
		return nil, ErrNilRedisClient
	}

	// Calculate optimal filter size and number of hash functions
	bitSize, hashCount := calculateOptimalParameters(cfg.ExpectedInsertions, cfg.FalsePositiveRate)

	// Set default hash strategy if not provided
	if cfg.HashStrategy == nil {
		cfg.HashStrategy = NewXXHashStrategy()
	}

	return &bloomFilter{
		config:       cfg,
		bitSize:      bitSize,
		hashCount:    hashCount,
		hashStrategy: cfg.HashStrategy,
	}, nil
}

// Add adds an element to the Bloom Filter
func (bf *bloomFilter) Add(data []byte) error {
	ctx := context.Background()
	positions := bf.getHashPositions(data)

	// Use pipeline for efficiency
	pipe, ok := bf.config.RedisClient.Pipeline().(pipeliner)
	if !ok {
		return ErrNilRedisClient
	}
	for _, pos := range positions {
		pipe.SetBit(ctx, bf.config.RedisKey, int64(pos), 1)
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	// Set TTL if configured and greater than zero
	if bf.config.TTL > 0 {
		if adapter, ok := bf.config.RedisClient.(*RedisAdapter); ok {
			adapter.client.Expire(ctx, bf.config.RedisKey, bf.config.TTL)
		}
	}

	return nil
}

// Exists checks if an element exists in the Bloom Filter
func (bf *bloomFilter) Exists(data []byte) (bool, error) {
	ctx := context.Background()
	positions := bf.getHashPositions(data)

	// Use pipeline for efficiency
	pipe, ok := bf.config.RedisClient.Pipeline().(pipeliner)
	if !ok {
		return false, ErrNilRedisClient
	}
	cmds := make([]*redis.IntCmd, len(positions))

	for i, pos := range positions {
		cmds[i] = pipe.GetBit(ctx, bf.config.RedisKey, int64(pos))
	}

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, err
	}

	// Check if all bits are set
	for _, cmd := range cmds {
		if cmd.Val() == 0 {
			return false, nil
		}
	}

	return true, nil
}

// getHashPositions calculates the k hash positions for the given data
// using double hashing technique: position = (h1(data) + i * h2(data)) % m
func (bf *bloomFilter) getHashPositions(data []byte) []uint64 {
	positions := make([]uint64, bf.hashCount)

	// Get two hash values for double hashing
	h1 := bf.hashStrategy.Hash(data, 0)
	h2 := bf.hashStrategy.Hash(data, 1)

	// Ensure h2 is odd for better distribution
	if h2%2 == 0 {
		h2++
	}

	for i := uint(0); i < bf.hashCount; i++ {
		position := (h1 + uint64(i)*h2) % bf.bitSize
		positions[i] = position
	}

	return positions
}
