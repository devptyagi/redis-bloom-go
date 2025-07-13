# Redis-Backed Bloom Filter Library for Go

A lightweight, production-grade, Redis-backed Bloom Filter implementation in Go, optimized for distributed deployments like AWS ElastiCache Redis Cluster.

## Features

✅ **Distributed Bloom Filter** - Backed by Redis or Redis Cluster  
✅ **Configurable Parameters** - Automatic optimal filter size and hash function calculation  
✅ **Pipelined Operations** - High-performance Redis operations  
✅ **Multiple Hash Strategies** - XXHash (default), Murmur3, and FNV  
✅ **TTL Support** - Expirable filters for temporary data  
✅ **Clean API** - Simple, intuitive interface  
✅ **Production Ready** - Comprehensive testing and error handling  

## Installation

```bash
go get github.com/devptyagi/redis-bloom-go
```

## Quick Start

```go
package main

import (
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/devptyagi/redis-bloom-go/bloom"
)

func main() {
    // Create Redis client (for Docker Compose, use service name)
    client := redis.NewClient(&redis.Options{
        Addr: "redis:6379",
    })

    // Create Bloom Filter
    bf, err := bloom.NewBloomFilter(bloom.Config{
        RedisKey:           "user:emails",
        RedisClient:        bloom.NewSingleNodeRedisClient(client),
        ExpectedInsertions: 1_000_000,
        FalsePositiveRate:  0.01, // 1%
        TTL:                24 * time.Hour,
    })
    if err != nil {
        panic(err)
    }

    // Add an email
    err = bf.Add([]byte("user@example.com"))
    if err != nil {
        panic(err)
    }

    // Check if email exists
    exists, err := bf.Exists([]byte("user@example.com"))
    if err != nil {
        panic(err)
    }
    println("Email exists:", exists) // true
}
```

## API Reference

### Configuration

```go
type Config struct {
    RedisKey           string        // Redis key for the Bloom Filter
    RedisClient        RedisClient   // Redis client (single-node or cluster)
    ExpectedInsertions uint64        // Expected number of insertions
    FalsePositiveRate  float64       // Desired false positive rate (0.0-1.0)
    TTL                time.Duration // Optional TTL for the filter
    HashStrategy       HashStrategy  // Optional hash strategy (defaults to XXHash)
}
```

### Bloom Filter Interface

```go
type BloomFilter interface {
    Add(data []byte) error           // Add an element to the filter
    Exists(data []byte) (bool, error) // Check if an element exists
}
```

### Hash Strategies

```go
// XXHash (fastest, default)
strategy := bloom.NewXXHashStrategy()

// Murmur3
strategy := bloom.NewMurmur3Strategy()

// FNV
strategy := bloom.NewFNVStrategy()
```

### Redis Client Adapters

```go
// Single-node Redis (for Docker Compose, use service name)
client := redis.NewClient(&redis.Options{Addr: "redis:6379"})
redisClient := bloom.NewSingleNodeRedisClient(client)

// Redis Cluster (for Docker Compose, use service name and internal ports)
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "redis-cluster:7000",
        "redis-cluster:7001",
        "redis-cluster:7002",
        "redis-cluster:7003",
        "redis-cluster:7004",
        "redis-cluster:7005",
    },
})
redisClient := bloom.NewClusterRedisClient(clusterClient)
```

## Advanced Examples

### Redis Cluster with Hash Tags

```go
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "redis-cluster:7000",
        "redis-cluster:7001",
        "redis-cluster:7002",
        "redis-cluster:7003",
        "redis-cluster:7004",
        "redis-cluster:7005",
    },
})

bf, err := bloom.NewBloomFilter(bloom.Config{
    RedisKey:           "bloom:{user:emails}", // Hash tag for cluster distribution
    RedisClient:        bloom.NewClusterRedisClient(clusterClient),
    ExpectedInsertions: 500_000,
    FalsePositiveRate:  0.005, // 0.5%
    TTL:                12 * time.Hour,
})
```

### Custom Hash Strategy

```go
bf, err := bloom.NewBloomFilter(bloom.Config{
    RedisKey:           "custom:hash:test",
    RedisClient:        redisClient,
    ExpectedInsertions: 100_000,
    FalsePositiveRate:  0.01,
    HashStrategy:       bloom.NewMurmur3Strategy(),
})
```

### TTL for Temporary Data

```go
bf, err := bloom.NewBloomFilter(bloom.Config{
    RedisKey:           "temp:session:ids",
    RedisClient:        redisClient,
    ExpectedInsertions: 10_000,
    FalsePositiveRate:  0.01,
    TTL:                30 * time.Minute, // Auto-expire after 30 minutes
})
```

## Bloom Filter Theory

The library automatically calculates optimal parameters using standard Bloom Filter formulas:

```
m = -(n * ln(p)) / (ln(2)^2)  // total bits
k = (m / n) * ln(2)           // number of hash functions
```

Where:
- `n` = expected number of insertions
- `p` = false positive probability
- `m` = total bits in the filter
- `k` = number of hash functions

The library uses double hashing to derive k hash positions:
```
position = (h1(data) + i * h2(data)) % m
```

## Testing

### End-to-End Testing (Docker Only)

**All tests are run inside Docker Compose using the test service. This is the only supported method.**

```bash
# Install dependencies
make install

# Run all integration tests (single-node and cluster)
make test

# Clean up
make clean
```

- This will start both Redis and Redis Cluster containers, then run all integration tests inside a Go container on the same Docker network.
- All Redis addresses in your code/tests should use service names (e.g., `redis:6379`, `redis-cluster:7000`).
- No host-based or manual testing is supported.

### Makefile

```bash
make help   # Show all available commands
make test   # Run all integration tests (inside Docker)
make clean  # Clean up containers and build artifacts
```

**Note:** There are no unit tests or mocks; all tests are integration/e2e and run inside Docker Compose.

## Performance

The library is optimized for high-throughput scenarios:

- **Pipelined Redis Operations** - Multiple SETBIT/GETBIT commands are batched
- **Efficient Hash Functions** - XXHash provides excellent performance
- **Minimal Memory Overhead** - Only stores the Bloom Filter bits in Redis
- **Configurable Parameters** - Optimize for your specific use case

## Use Cases

- **Duplicate Detection** - Prevent processing the same data multiple times
- **Cache Warming** - Check if data exists before expensive operations
- **Rate Limiting** - Track unique users/requests over time windows
- **Spam Filtering** - Check if email/URL has been seen before
- **Session Management** - Track active sessions with TTL

## Error Handling

The library provides comprehensive error handling:

```go
bf, err := bloom.NewBloomFilter(config)
if err != nil {
    switch err {
    case bloom.ErrInvalidExpectedInsertions:
        // Expected insertions must be > 0
    case bloom.ErrInvalidFalsePositiveRate:
        // False positive rate must be between 0 and 1
    case bloom.ErrEmptyRedisKey:
        // Redis key cannot be empty
    case bloom.ErrNilRedisClient:
        // Redis client cannot be nil
    }
}
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [Redis Go Client](https://github.com/redis/go-redis) - Redis client library
- [XXHash](https://github.com/cespare/xxhash) - Fast hash function
- [Murmur3](https://github.com/spaolacci/murmur3) - Murmur3 hash implementation 