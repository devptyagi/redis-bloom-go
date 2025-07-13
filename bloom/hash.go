package bloom

import (
	"hash/fnv"

	"github.com/cespare/xxhash/v2"
	"github.com/spaolacci/murmur3"
)

// HashStrategy defines the interface for hash functions used in the Bloom Filter
type HashStrategy interface {
	Hash(data []byte, i uint) uint64
}

// XXHashStrategy implements HashStrategy using xxhash (fastest)
type XXHashStrategy struct{}

// NewXXHashStrategy creates a new XXHash strategy instance
func NewXXHashStrategy() HashStrategy {
	return &XXHashStrategy{}
}

// Hash implements HashStrategy using xxhash with seed variation
func (x *XXHashStrategy) Hash(data []byte, i uint) uint64 {
	// Use different seeds for different hash functions
	seed := uint64(i) * 0x9e3779b185ebca87 // golden ratio
	h := xxhash.New()
	h.Write([]byte{byte(seed), byte(seed >> 8), byte(seed >> 16), byte(seed >> 24)})
	h.Write(data)
	return h.Sum64()
}

// Murmur3Strategy implements HashStrategy using Murmur3
type Murmur3Strategy struct{}

// NewMurmur3Strategy creates a new Murmur3 strategy instance
func NewMurmur3Strategy() HashStrategy {
	return &Murmur3Strategy{}
}

// Hash implements HashStrategy using Murmur3 with seed variation
func (m *Murmur3Strategy) Hash(data []byte, i uint) uint64 {
	// Use different seeds for different hash functions
	seed := uint32(i) * 0x9e3779b9 // golden ratio
	return uint64(murmur3.Sum32WithSeed(data, seed))
}

// FNVStrategy implements HashStrategy using FNV-1a
type FNVStrategy struct{}

// NewFNVStrategy creates a new FNV strategy instance
func NewFNVStrategy() HashStrategy {
	return &FNVStrategy{}
}

// Hash implements HashStrategy using FNV-1a with seed variation
func (f *FNVStrategy) Hash(data []byte, i uint) uint64 {
	h := fnv.New64a()

	// Add seed to the data for variation
	seed := []byte{byte(i), byte(i >> 8), byte(i >> 16), byte(i >> 24)}
	h.Write(seed)
	h.Write(data)

	return h.Sum64()
}
