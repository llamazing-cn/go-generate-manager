package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type FileCache struct {
	path   string
	hashes map[string]string
	mu     sync.RWMutex
}

func NewFileCache(path string) *FileCache {
	return &FileCache{
		path:   path,
		hashes: make(map[string]string),
	}
}

func (c *FileCache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	content, err := os.ReadFile(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read cache file: %w", err)
	}

	c.hashes = make(map[string]string)
	for _, line := range strings.Split(string(content), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			c.hashes[parts[0]] = parts[1]
		}
	}
	return nil
}

func (c *FileCache) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	file, err := os.Create(c.path)
	if err != nil {
		return fmt.Errorf("create cache file: %w", err)
	}
	defer file.Close()

	for path, hash := range c.hashes {
		if _, err := fmt.Fprintf(file, "%s %s\n", path, hash); err != nil {
			return fmt.Errorf("write cache entry: %w", err)
		}
	}
	return nil
}

func (c *FileCache) Get(path string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	hash, exists := c.hashes[path]
	return hash, exists
}

func (c *FileCache) Set(path, hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hashes[path] = hash
}
