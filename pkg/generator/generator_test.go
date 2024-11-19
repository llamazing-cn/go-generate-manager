package generator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

type mockHasher struct {
	hashes map[string]string
}

func (h *mockHasher) Hash(path string) (string, error) {
	return h.hashes[path], nil
}

func (h *mockHasher) IsChanged(path, oldHash string) bool {
	newHash, _ := h.Hash(path)
	return newHash != oldHash
}

type mockCache struct {
	data map[string]string
}

func (c *mockCache) Load() error                    { return nil }
func (c *mockCache) Save() error                    { return nil }
func (c *mockCache) Get(path string) (string, bool) { h, ok := c.data[path]; return h, ok }
func (c *mockCache) Set(path, hash string)          { c.data[path] = hash }

type mockCommand struct {
	path     string
	executed bool
}

func (c *mockCommand) Execute(ctx context.Context) error { c.executed = true; return nil }
func (c *mockCommand) GetFilePath() string               { return c.path }
func (c *mockCommand) String() string                    { return "mock command" }

type mockFinder struct {
	commands []Command
}

func (f *mockFinder) Find(dir string) ([]Command, error) {
	return f.commands, nil
}

func TestGenerator(t *testing.T) {
	// 创建测试目录结构
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("//go:generate mockgen"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		setupMocks    func() (FileHasher, Cache, CommandFinder)
		expectedCalls int
		expectError   bool
	}{
		{
			name: "should regenerate when file changed",
			setupMocks: func() (FileHasher, Cache, CommandFinder) {
				cmd := &mockCommand{path: testFile}
				return &mockHasher{
						hashes: map[string]string{testFile: "new-hash"},
					},
					&mockCache{
						data: map[string]string{testFile: "old-hash"},
					},
					&mockFinder{
						commands: []Command{cmd},
					}
			},
			expectedCalls: 1,
			expectError:   false,
		},
		{
			name: "should not regenerate when file unchanged",
			setupMocks: func() (FileHasher, Cache, CommandFinder) {
				cmd := &mockCommand{path: testFile}
				hash := "same-hash"
				return &mockHasher{
						hashes: map[string]string{testFile: hash},
					},
					&mockCache{
						data: map[string]string{testFile: hash},
					},
					&mockFinder{
						commands: []Command{cmd},
					}
			},
			expectedCalls: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasher, cache, finder := tt.setupMocks()
			gen := New(Options{
				Hasher:  hasher,
				Cache:   cache,
				Finder:  finder,
				Workers: 1,
			})

			err := gen.Generate(context.Background(), tmpDir)
			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// 验证命令执行次数
			commands, _ := finder.Find("")
			executedCount := 0
			for _, cmd := range commands {
				if mc, ok := cmd.(*mockCommand); ok && mc.executed {
					executedCount++
				}
			}
			if executedCount != tt.expectedCalls {
				t.Errorf("expected %d commands to be executed, got %d", tt.expectedCalls, executedCount)
			}
		})
	}
}
