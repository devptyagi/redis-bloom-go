package bloom

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// RedisAdapter wraps Redis clients to provide a unified interface
type RedisAdapter struct {
	client redis.Cmdable
}

var _ RedisClient = (*RedisAdapter)(nil)

// NewRedisAdapter creates a new Redis adapter from a Redis client
func NewRedisAdapter(client redis.Cmdable) RedisClient {
	return &RedisAdapter{client: client}
}

// SetBit sets a bit at the specified offset
func (ra *RedisAdapter) SetBit(ctx context.Context, key string, offset int64, value int) *redis.IntCmd {
	return ra.client.SetBit(ctx, key, offset, value)
}

// GetBit gets a bit at the specified offset
func (ra *RedisAdapter) GetBit(ctx context.Context, key string, offset int64) *redis.IntCmd {
	return ra.client.GetBit(ctx, key, offset)
}

// Pipeline returns a new pipeline
func (ra *RedisAdapter) Pipeline() pipeliner {
	return ra.client.Pipeline()
}

// NewSingleNodeRedisClient creates a Redis adapter for a single-node Redis client
func NewSingleNodeRedisClient(client *redis.Client) RedisClient {
	return NewRedisAdapter(client)
}

// NewClusterRedisClient creates a Redis adapter for a Redis cluster client
func NewClusterRedisClient(client *redis.ClusterClient) RedisClient {
	return NewRedisAdapter(client)
}
