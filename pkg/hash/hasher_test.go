package hash

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContentHasher(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	content := []byte("test content")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	hasher := NewContentHasher()

	// 测试哈希计算
	hash1, err := hasher.Hash(testFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 测试相同内容产生相同哈希
	hash2, err := hasher.Hash(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 != hash2 {
		t.Error("same content should produce same hash")
	}

	// 测试内容变化
	if err := os.WriteFile(testFile, []byte("different content"), 0644); err != nil {
		t.Fatal(err)
	}

	hash3, err := hasher.Hash(testFile)
	if err != nil {
		t.Fatal(err)
	}
	if hash1 == hash3 {
		t.Error("different content should produce different hash")
	}

	// 测试 IsChanged
	if !hasher.IsChanged(testFile, hash1) {
		t.Error("IsChanged should return true for different content")
	}
}
