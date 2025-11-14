package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"sync"
	"time"
)

// ModuleCache represents a cache for module file hashes and metadata
type ModuleCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
}

// CacheEntry stores metadata about a cached file
type CacheEntry struct {
	Hash         string
	LastModified time.Time
	Size         int64
	Dependencies []string
}

// NewModuleCache creates a new module cache
func NewModuleCache() *ModuleCache {
	return &ModuleCache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get retrieves a cache entry for a file
func (c *ModuleCache) Get(path string) (*CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, exists := c.entries[path]
	return entry, exists
}

// Set stores a cache entry for a file
func (c *ModuleCache) Set(path string, entry *CacheEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[path] = entry
}

// Invalidate removes a cache entry
func (c *ModuleCache) Invalidate(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, path)
}

// IsValid checks if a file's cache entry is still valid
func (c *ModuleCache) IsValid(path string) (bool, error) {
	entry, exists := c.Get(path)
	if !exists {
		return false, nil
	}

	// Check if file still exists and hasn't changed
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	// Compare modification time and size
	if !info.ModTime().Equal(entry.LastModified) || info.Size() != entry.Size {
		return false, nil
	}

	// Compute current hash for verification
	hash, err := ComputeFileHash(path)
	if err != nil {
		return false, err
	}

	return hash == entry.Hash, nil
}

// ComputeFileHash computes SHA-256 hash of a file
func ComputeFileHash(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// UpdateCache updates the cache entry for a file
func (c *ModuleCache) UpdateCache(path string, deps []string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	hash, err := ComputeFileHash(path)
	if err != nil {
		return err
	}

	entry := &CacheEntry{
		Hash:         hash,
		LastModified: info.ModTime(),
		Size:         info.Size(),
		Dependencies: deps,
	}

	c.Set(path, entry)
	return nil
}

// GetStats returns cache statistics
func (c *ModuleCache) GetStats() map[string]int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]int{
		"total_entries": len(c.entries),
	}
}
