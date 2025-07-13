//go:build integration
// +build integration

// NOTE: These tests are designed to run inside a Docker container on the same Docker Compose network as the Redis services.
// Use service names as hostnames (e.g., 'redis', 'redis-cluster') and internal ports (6379, 7000-7005).

package bloom

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// Utility: clean up a Redis key before/after a test
func cleanupKey(client *redis.Client, key string) {
	client.Del(context.Background(), key)
}

func TestIntegrationWithRealRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available, skipping integration test: %v", err)
	}
	defer client.Close()
	redisClient := NewSingleNodeRedisClient(client)

	t.Run("BasicAddAndExists", func(t *testing.T) {
		key := "integration:test:basic"
		cleanupKey(client, key)
		defer cleanupKey(client, key)
		bf, err := NewBloomFilter(Config{
			RedisKey:           key,
			RedisClient:        redisClient,
			ExpectedInsertions: 1000,
			FalsePositiveRate:  0.01,
		})
		if err != nil {
			t.Fatalf("Failed to create Bloom Filter: %v", err)
		}
		testData := []byte("integration_test_data")
		exists, err := bf.Exists(testData)
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Data should not exist initially")
		}
		err = bf.Add(testData)
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
		exists, err = bf.Exists(testData)
		if err != nil {
			t.Fatalf("Failed to check existence after adding: %v", err)
		}
		if !exists {
			t.Error("Data should exist after adding")
		}
	})

	t.Run("FalsePositiveRate", func(t *testing.T) {
		key := "integration:test:fpr"
		cleanupKey(client, key)
		defer cleanupKey(client, key)
		bf, err := NewBloomFilter(Config{
			RedisKey:           key,
			RedisClient:        redisClient,
			ExpectedInsertions: 1000,
			FalsePositiveRate:  0.01,
		})
		if err != nil {
			t.Fatalf("Failed to create Bloom Filter: %v", err)
		}
		addedElements := make([][]byte, 1000)
		for i := 0; i < 1000; i++ {
			data := []byte(fmt.Sprintf("integration_element_%d", i))
			addedElements[i] = data
			err := bf.Add(data)
			if err != nil {
				t.Fatalf("Failed to add element %d: %v", i, err)
			}
		}
		for i, data := range addedElements {
			exists, err := bf.Exists(data)
			if err != nil {
				t.Fatalf("Failed to check element %d: %v", i, err)
			}
			if !exists {
				t.Errorf("Added element %d should exist", i)
			}
		}
		falsePositives := 0
		for i := 0; i < 1000; i++ {
			data := []byte(fmt.Sprintf("integration_unseen_element_%d", i))
			exists, err := bf.Exists(data)
			if err != nil {
				t.Fatalf("Failed to check unseen element %d: %v", i, err)
			}
			if exists {
				falsePositives++
			}
		}
		falsePositiveRate := float64(falsePositives) / 1000.0
		expectedMaxRate := 0.01
		if falsePositiveRate > expectedMaxRate*2 {
			t.Errorf("False positive rate %f exceeds expected maximum %f", falsePositiveRate, expectedMaxRate)
		}
		t.Logf("Observed false positive rate: %f (expected max: %f)", falsePositiveRate, expectedMaxRate)
	})

	t.Run("HashStrategies", func(t *testing.T) {
		testData := []byte("integration_hash_test")
		for _, strategy := range []struct {
			name     string
			strategy HashStrategy
		}{
			{"XXHash", NewXXHashStrategy()},
			{"Murmur3", NewMurmur3Strategy()},
			{"FNV", NewFNVStrategy()},
		} {
			key := fmt.Sprintf("integration:hash:%s", strategy.name)
			cleanupKey(client, key)
			defer cleanupKey(client, key)
			bf, err := NewBloomFilter(Config{
				RedisKey:           key,
				RedisClient:        redisClient,
				ExpectedInsertions: 1000,
				FalsePositiveRate:  0.01,
				HashStrategy:       strategy.strategy,
			})
			if err != nil {
				t.Fatalf("Failed to create Bloom Filter with %s: %v", strategy.name, err)
			}
			err = bf.Add(testData)
			if err != nil {
				t.Fatalf("Failed to add data with %s: %v", strategy.name, err)
			}
			exists, err := bf.Exists(testData)
			if err != nil {
				t.Fatalf("Failed to check data with %s: %v", strategy.name, err)
			}
			if !exists {
				t.Errorf("Data should exist with %s strategy", strategy.name)
			}
		}
	})

	t.Run("TTL", func(t *testing.T) {
		key := "integration:test:ttl"
		cleanupKey(client, key)
		defer cleanupKey(client, key)
		bf, err := NewBloomFilter(Config{
			RedisKey:           key,
			RedisClient:        redisClient,
			ExpectedInsertions: 1000,
			FalsePositiveRate:  0.01,
			TTL:                2 * time.Second,
		})
		if err != nil {
			t.Fatalf("Failed to create Bloom Filter: %v", err)
		}
		testData := []byte("integration_ttl_test")
		err = bf.Add(testData)
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
		exists, err := bf.Exists(testData)
		if err != nil {
			t.Fatalf("Failed to check data: %v", err)
		}
		if !exists {
			t.Error("Data should exist immediately after adding")
		}
		t.Log("Waiting for TTL to expire...")
		time.Sleep(3 * time.Second)
		exists, err = bf.Exists(testData)
		if err != nil {
			t.Fatalf("Failed to check data after TTL: %v", err)
		}
		if exists {
			t.Error("Data should not exist after TTL expiration")
		}
	})
}

func TestIntegrationWithRedisCluster(t *testing.T) {
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"redis-cluster:7000",
			"redis-cluster:7001",
			"redis-cluster:7002",
			"redis-cluster:7003",
			"redis-cluster:7004",
			"redis-cluster:7005",
		},
		RouteRandomly:  true,
		RouteByLatency: true,
	})
	ctx := context.Background()
	if err := clusterClient.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis Cluster not available, skipping cluster test: %v", err)
	}
	defer clusterClient.Close()
	redisClient := NewClusterRedisClient(clusterClient)
	t.Run("ClusterBasicFunctionality", func(t *testing.T) {
		key := "bloom:{integration:cluster:test}"
		clusterClient.Del(ctx, key)
		defer clusterClient.Del(ctx, key)
		bf, err := NewBloomFilter(Config{
			RedisKey:           key,
			RedisClient:        redisClient,
			ExpectedInsertions: 1000,
			FalsePositiveRate:  0.01,
		})
		if err != nil {
			t.Fatalf("Failed to create Bloom Filter: %v", err)
		}
		testData := []byte("integration_cluster_test_data")
		exists, err := bf.Exists(testData)
		if err != nil {
			t.Fatalf("Failed to check existence: %v", err)
		}
		if exists {
			t.Error("Data should not exist initially")
		}
		err = bf.Add(testData)
		if err != nil {
			t.Fatalf("Failed to add data: %v", err)
		}
		exists, err = bf.Exists(testData)
		if err != nil {
			t.Fatalf("Failed to check existence after adding: %v", err)
		}
		if !exists {
			t.Error("Data should exist after adding")
		}
	})
}

// Benchmark tests for performance
func BenchmarkBloomFilterAdd(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available, skipping benchmark: %v", err)
	}
	defer client.Close()
	redisClient := NewSingleNodeRedisClient(client)
	bf, err := NewBloomFilter(Config{
		RedisKey:           "benchmark:add",
		RedisClient:        redisClient,
		ExpectedInsertions: 1000000,
		FalsePositiveRate:  0.01,
	})
	if err != nil {
		b.Fatalf("Failed to create Bloom Filter: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := []byte(fmt.Sprintf("benchmark_data_%d", i))
		err := bf.Add(data)
		if err != nil {
			b.Fatalf("Failed to add data: %v", err)
		}
	}
}

func BenchmarkBloomFilterExists(b *testing.B) {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		b.Skipf("Redis not available, skipping benchmark: %v", err)
	}
	defer client.Close()
	redisClient := NewSingleNodeRedisClient(client)
	bf, err := NewBloomFilter(Config{
		RedisKey:           "benchmark:exists",
		RedisClient:        redisClient,
		ExpectedInsertions: 1000000,
		FalsePositiveRate:  0.01,
	})
	if err != nil {
		b.Fatalf("Failed to create Bloom Filter: %v", err)
	}
	// Pre-add some data
	testData := []byte("benchmark_test_data")
	err = bf.Add(testData)
	if err != nil {
		b.Fatalf("Failed to add test data: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bf.Exists(testData)
		if err != nil {
			b.Fatalf("Failed to check existence: %v", err)
		}
	}
}
