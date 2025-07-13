# Redis Bloom Filter Library - Setup Guide

This guide explains how to set up and test the Redis Bloom Filter library using Docker and the provided Makefile.

## Prerequisites

- **Docker** - For running Redis containers
- **Docker Compose** - For managing multi-container setups
- **Go 1.21+** - For running the library and tests
- **Make** - For using the provided Makefile (optional but recommended)

## Quick Start

### 1. Clone and Setup

```bash
git clone <repository-url>
cd redis-bloom-go
make install
```

### 2. Run Tests

```bash
# Run all integration tests (this will start Redis services automatically)
make test
```

This will:
- Start Redis single-node and Redis cluster containers
- Run comprehensive integration tests
- Show you that everything is working

### 3. Clean Up

```bash
make clean
```

## Development Workflow

### Using the Makefile

The project includes a simple Makefile with essential commands:

```bash
# Show all available commands
make help

# Development
make install        # Install dependencies (go mod tidy)
make test           # Run all integration tests (requires Docker)
make clean          # Remove build artifacts and containers
```

### Manual Docker Commands

If you prefer to use Docker Compose directly:

```bash
# Start all services and run tests
docker-compose up --build --abort-on-container-exit test

# Start services only (without running tests)
docker-compose up -d redis redis-cluster

# Stop all containers
docker-compose down -v
```

## Testing

### Integration Tests

The project uses Docker Compose to run integration tests with real Redis instances:

```bash
make test
```

This command:
- Starts a single-node Redis container (`redis:6379`)
- Starts a Redis cluster container (`redis-cluster:7000-7005`)
- Runs the test service with Go 1.22
- Executes all integration tests with the `integration` build tag

### Test Coverage

The integration tests cover:
- Basic add and exists operations
- False positive rate validation
- Multiple hash strategies (XXHash, Murmur3, FNV)
- TTL (Time To Live) functionality
- Redis cluster compatibility
- Performance benchmarks

### Test Environment

Tests run inside a Docker container with:
- **Go 1.22** runtime
- Access to both single-node Redis and Redis cluster
- Network connectivity to Redis services via Docker Compose
- Volume mounting for source code

## Project Structure

```
redis-bloom-go/
├── bloom/                    # Main library code
│   ├── bloom.go             # Core Bloom filter implementation
│   ├── redis.go             # Redis client adapters
│   ├── hash.go              # Hash strategy implementations
│   ├── config.go            # Configuration structures
│   ├── errors.go            # Error definitions
│   └── bloom_integration_test.go  # Integration tests
├── examples/                # Usage examples
│   └── main.go             # Comprehensive examples
├── docker-compose.yaml      # Docker services configuration
├── Makefile                 # Development commands
├── go.mod                   # Go module definition
└── README.md               # Project documentation
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "time"
    "github.com/devptyagi/redis-bloom-go/bloom"
    "github.com/redis/go-redis/v9"
)

func main() {
    // Create Redis client
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Create Bloom filter
    bf, err := bloom.NewBloomFilter(bloom.Config{
        RedisKey:           "user:emails",
        RedisClient:        bloom.NewSingleNodeRedisClient(client),
        ExpectedInsertions: 1000000,
        FalsePositiveRate:  0.01,
        TTL:                24 * time.Hour,
    })
    
    // Add data
    bf.Add([]byte("user@example.com"))
    
    // Check existence
    exists, _ := bf.Exists([]byte("user@example.com"))
}
```

### Redis Cluster Usage

```go
// Create cluster client
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{"localhost:7000", "localhost:7001", "localhost:7002"},
})

// Use hash tags for cluster distribution
bf, err := bloom.NewBloomFilter(bloom.Config{
    RedisKey:           "bloom:{user:emails}", // Hash tag for distribution
    RedisClient:        bloom.NewClusterRedisClient(clusterClient),
    ExpectedInsertions: 500000,
    FalsePositiveRate:  0.005,
})
```

## Troubleshooting

### Common Issues

1. **Docker not running**
   ```
   Error: Docker is not running
   ```
   Solution: Start Docker Desktop or Docker daemon

2. **Port conflicts**
   ```
   Error: Port 6379 or 7000-7005 is already in use
   ```
   Solution: Stop any existing Redis instances or change ports in docker-compose.yaml

3. **Integration tests failing**
   ```
   Error: connection refused
   ```
   Solution: Ensure Docker containers are running with `docker-compose ps`

4. **Go version issues**
   ```
   Error: go: go.mod requires go 1.21
   ```
   Solution: Update to Go 1.21 or later

### Checking Status

```bash
# Check container status
docker-compose ps

# Check container logs
docker-compose logs redis
docker-compose logs redis-cluster
docker-compose logs test
```

### Reset Everything

```bash
make clean
```

This will stop all containers and clean build artifacts.

## Production Deployment

For production deployment:

1. **Use a production Redis instance** (AWS ElastiCache, Redis Cloud, etc.)
2. **Configure proper connection settings** (authentication, SSL, etc.)
3. **Set appropriate TTL values** for your use case
4. **Monitor performance** using the benchmark tests

### Example Production Configuration

```go
// AWS ElastiCache Redis Cluster
clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
    Addrs: []string{
        "your-cluster-endpoint-1:6379",
        "your-cluster-endpoint-2:6379",
        "your-cluster-endpoint-3:6379",
    },
    Password: "your-password",
    TLSConfig: &tls.Config{}, // If using SSL
})

bf, err := bloom.NewBloomFilter(bloom.Config{
    RedisKey:           "bloom:{your-app:filter}",
    RedisClient:        bloom.NewClusterRedisClient(clusterClient),
    ExpectedInsertions: 1_000_000,
    FalsePositiveRate:  0.01,
    TTL:                24 * time.Hour,
})
```

## Contributing

When contributing to the project:

1. **Run the full test suite**:
   ```bash
   make test
   ```

2. **Add tests** for new functionality:
   - Integration tests for Redis functionality
   - Unit tests for core logic

3. **Update documentation** if adding new features

4. **Test with both single-node and cluster** Redis

## Support

If you encounter issues:

1. Check the troubleshooting section above
2. Review the integration test logs: `docker-compose logs test`
3. Ensure Docker and Go versions meet requirements
4. Check that all required ports are available 