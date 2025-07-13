package bloom

import (
	"math"
	"time"
)

// Config holds the configuration for creating a new Bloom Filter
type Config struct {
	RedisKey           string
	RedisClient        RedisClient
	ExpectedInsertions uint64
	FalsePositiveRate  float64
	TTL                time.Duration
	HashStrategy       HashStrategy
}

// calculateOptimalParameters calculates the optimal number of bits and hash functions
// using the standard Bloom Filter formulas:
// m = -(n * ln(p)) / (ln(2)^2)  // total bits
// k = (m / n) * ln(2)           // number of hash functions
func calculateOptimalParameters(n uint64, p float64) (uint64, uint) {
	// Calculate optimal bit size
	ln2 := math.Ln2
	ln2Squared := ln2 * ln2
	bitSize := uint64(math.Ceil(-float64(n) * math.Log(p) / ln2Squared))

	// Calculate optimal number of hash functions
	hashCount := uint(math.Ceil(float64(bitSize) / float64(n) * ln2))

	// Ensure hashCount is at least 1
	if hashCount < 1 {
		hashCount = 1
	}

	return bitSize, hashCount
}
