package cache

import (
	"path/filepath"
	"testing"
)

func TestFileCache(t *testing.T) {
	tmpDir := t.TempDir()
	cacheFile := filepath.Join(tmpDir, "test.sum")

	cache := NewFileCache(cacheFile)

	// 测试设置和获取
	cache.Set("test.go", "hash1")
	if hash, exists := cache.Get("test.go"); !exists || hash != "hash1" {
		t.Error("cache Set/Get failed")
	}

	// 测试保存
	if err := cache.Save(); err != nil {
		t.Fatalf("unexpected error saving cache: %v", err)
	}

	// 测试加载
	newCache := NewFileCache(cacheFile)
	if err := newCache.Load(); err != nil {
		t.Fatalf("unexpected error loading cache: %v", err)
	}

	if hash, exists := newCache.Get("test.go"); !exists || hash != "hash1" {
		t.Error("cache Load failed")
	}

	// 测试并发安全性
	done := make(chan bool)
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("concurrent.go", "hash")
			cache.Get("concurrent.go")
		}
		done <- true
	}()
	go func() {
		for i := 0; i < 100; i++ {
			cache.Set("concurrent.go", "hash")
			cache.Get("concurrent.go")
		}
		done <- true
	}()

	<-done
	<-done
}
