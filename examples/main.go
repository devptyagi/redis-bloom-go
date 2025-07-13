package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/devptyagi/redis-bloom-go/bloom"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Example 1: Single-node Redis
	exampleSingleNode()

	// Example 2: Redis Cluster
	exampleCluster()

	// Example 3: Different hash strategies
	exampleHashStrategies()

	// Example 4: TTL support
	exampleTTL()
}

func exampleSingleNode() {
	fmt.Println("=== Single Node Redis Example ===")

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Println("Skipping single node example...")
		return
	}

	// Create Redis adapter
	redisClient := bloom.NewSingleNodeRedisClient(client)

	// Create Bloom Filter
	bf, err := bloom.NewBloomFilter(bloom.Config{
		RedisKey:           "user:emails:single",
		RedisClient:        redisClient,
		ExpectedInsertions: 1000000,
		FalsePositiveRate:  0.01,
		TTL:                24 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create Bloom Filter: %v", err)
	}

	// Add some emails
	emails := []string{
		"user1@example.com",
		"user2@example.com",
		"user3@example.com",
	}

	for _, email := range emails {
		if err := bf.Add([]byte(email)); err != nil {
			log.Printf("Failed to add %s: %v", email, err)
		} else {
			fmt.Printf("Added: %s\n", email)
		}
	}

	// Check if emails exist
	for _, email := range emails {
		exists, err := bf.Exists([]byte(email))
		if err != nil {
			log.Printf("Failed to check %s: %v", email, err)
		} else {
			fmt.Printf("Exists %s: %t\n", email, exists)
		}
	}

	// Check non-existent email
	exists, err := bf.Exists([]byte("nonexistent@example.com"))
	if err != nil {
		log.Printf("Failed to check nonexistent email: %v", err)
	} else {
		fmt.Printf("Exists nonexistent@example.com: %t\n", exists)
	}

	client.Close()
}

func exampleCluster() {
	fmt.Println("\n=== Redis Cluster Example ===")

	// Create Redis cluster client
	clusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: []string{
			"localhost:7000",
			"localhost:7001",
			"localhost:7002",
		},
	})

	// Test connection
	ctx := context.Background()
	if err := clusterClient.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Could not connect to Redis Cluster: %v", err)
		log.Println("Skipping cluster example...")
		return
	}

	// Create Redis adapter
	redisClient := bloom.NewClusterRedisClient(clusterClient)

	// Create Bloom Filter with hash tags for cluster distribution
	bf, err := bloom.NewBloomFilter(bloom.Config{
		RedisKey:           "bloom:{user:emails:cluster}", // Hash tag for cluster distribution
		RedisClient:        redisClient,
		ExpectedInsertions: 500000,
		FalsePositiveRate:  0.005,
		TTL:                12 * time.Hour,
	})
	if err != nil {
		log.Fatalf("Failed to create Bloom Filter: %v", err)
	}

	// Add some user IDs
	userIDs := []string{
		"user_12345",
		"user_67890",
		"user_11111",
	}

	for _, userID := range userIDs {
		if err := bf.Add([]byte(userID)); err != nil {
			log.Printf("Failed to add %s: %v", userID, err)
		} else {
			fmt.Printf("Added: %s\n", userID)
		}
	}

	// Check if user IDs exist
	for _, userID := range userIDs {
		exists, err := bf.Exists([]byte(userID))
		if err != nil {
			log.Printf("Failed to check %s: %v", userID, err)
		} else {
			fmt.Printf("Exists %s: %t\n", userID, exists)
		}
	}

	clusterClient.Close()
}

func exampleHashStrategies() {
	fmt.Println("\n=== Hash Strategies Example ===")

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Println("Skipping hash strategies example...")
		return
	}

	redisClient := bloom.NewSingleNodeRedisClient(client)

	// Test different hash strategies
	strategies := map[string]bloom.HashStrategy{
		"XXHash":  bloom.NewXXHashStrategy(),
		"Murmur3": bloom.NewMurmur3Strategy(),
		"FNV":     bloom.NewFNVStrategy(),
	}

	testData := []byte("test@example.com")

	for name, strategy := range strategies {
		bf, err := bloom.NewBloomFilter(bloom.Config{
			RedisKey:           fmt.Sprintf("hash:test:%s", name),
			RedisClient:        redisClient,
			ExpectedInsertions: 10000,
			FalsePositiveRate:  0.01,
			HashStrategy:       strategy,
		})
		if err != nil {
			log.Printf("Failed to create Bloom Filter with %s: %v", name, err)
			continue
		}

		// Add test data
		if err := bf.Add(testData); err != nil {
			log.Printf("Failed to add data with %s: %v", name, err)
			continue
		}

		// Check if data exists
		exists, err := bf.Exists(testData)
		if err != nil {
			log.Printf("Failed to check data with %s: %v", name, err)
		} else {
			fmt.Printf("%s strategy - Exists: %t\n", name, exists)
		}
	}

	client.Close()
}

func exampleTTL() {
	fmt.Println("\n=== TTL Example ===")

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Warning: Could not connect to Redis: %v", err)
		log.Println("Skipping TTL example...")
		return
	}

	redisClient := bloom.NewSingleNodeRedisClient(client)

	// Create Bloom Filter with short TTL
	bf, err := bloom.NewBloomFilter(bloom.Config{
		RedisKey:           "ttl:test",
		RedisClient:        redisClient,
		ExpectedInsertions: 1000,
		FalsePositiveRate:  0.01,
		TTL:                5 * time.Second, // Short TTL for demo
	})
	if err != nil {
		log.Fatalf("Failed to create Bloom Filter: %v", err)
	}

	// Add data
	testData := []byte("temp@example.com")
	if err := bf.Add(testData); err != nil {
		log.Printf("Failed to add data: %v", err)
		return
	}

	fmt.Printf("Added data with 5-second TTL\n")

	// Check immediately
	exists, err := bf.Exists(testData)
	if err != nil {
		log.Printf("Failed to check data: %v", err)
	} else {
		fmt.Printf("Immediately after adding - Exists: %t\n", exists)
	}

	// Wait for TTL to expire
	fmt.Println("Waiting 6 seconds for TTL to expire...")
	time.Sleep(6 * time.Second)

	// Check after TTL expiration
	exists, err = bf.Exists(testData)
	if err != nil {
		log.Printf("Failed to check data after TTL: %v", err)
	} else {
		fmt.Printf("After TTL expiration - Exists: %t\n", exists)
	}

	client.Close()
}
