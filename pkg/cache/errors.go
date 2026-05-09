package cache

import "errors"

// Error definitions
var (
	ErrCacheMiss = errors.New("cache: key not found")
	ErrCacheErr  = errors.New("cache: operation failed")
)
