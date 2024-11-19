package command

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCommandFinder(t *testing.T) {
	// 创建测试目录和文件
	tmpDir := t.TempDir()

	files := map[string]string{
		"test1.go": "//go:generate mockgen -source=test1.go",
		"test2.go": "//go:generate protoc --go_out=. test2.proto",
		"test3.go": "// normal comment",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name          string
		pattern       string
		expectedCount int
	}{
		{
			name:          "find mockgen commands",
			pattern:       "mockgen",
			expectedCount: 1,
		},
		{
			name:          "find protoc commands",
			pattern:       "protoc",
			expectedCount: 1,
		},
		{
			name:          "find non-existent commands",
			pattern:       "invalid",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFinder(tt.pattern)
			commands, err := finder.Find(tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(commands) != tt.expectedCount {
				t.Errorf("expected %d commands, got %d", tt.expectedCount, len(commands))
			}
		})
	}
}

func TestGoGenCommand(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	cmd := NewCommand(testFile, "echo test")

	// 测试命令执行
	err := cmd.Execute(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// 测试文件路径获取
	if cmd.GetFilePath() != testFile {
		t.Errorf("expected file path %s, got %s", testFile, cmd.GetFilePath())
	}

	// 测试命令字符串
	if cmd.String() != "echo test" {
		t.Errorf("expected command string 'echo test', got '%s'", cmd.String())
	}
}
