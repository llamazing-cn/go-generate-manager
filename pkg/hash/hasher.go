package hash

import (
	"fmt"
	"hash"
	"os"
	"sync"

	"github.com/cespare/xxhash/v2"
)

type ContentHasher struct {
	pool *sync.Pool
}

func NewContentHasher() *ContentHasher {
	return &ContentHasher{
		pool: &sync.Pool{
			New: func() interface{} {
				return xxhash.New()
			},
		},
	}
}

func (h *ContentHasher) Hash(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	hasher := h.pool.Get().(hash.Hash64)
	defer func() {
		hasher.Reset()
		h.pool.Put(hasher)
	}()

	if _, err := hasher.Write(content); err != nil {
		return "", fmt.Errorf("hash content: %w", err)
	}

	return fmt.Sprintf("xxhash:%x", hasher.Sum64()), nil
}

func (h *ContentHasher) IsChanged(path, oldHash string) bool {
	newHash, err := h.Hash(path)
	if err != nil {
		return true // 如果无法计算新哈希，认为文件已更改
	}
	return newHash != oldHash
}
