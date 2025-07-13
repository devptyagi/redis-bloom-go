package bloom

import "errors"

// Error definitions for the Bloom Filter library
var (
	ErrInvalidExpectedInsertions = errors.New("expected insertions must be greater than 0")
	ErrInvalidFalsePositiveRate  = errors.New("false positive rate must be between 0 and 1")
	ErrEmptyRedisKey             = errors.New("redis key cannot be empty")
	ErrNilRedisClient            = errors.New("redis client cannot be nil")
)
