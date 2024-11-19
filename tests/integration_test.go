package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"llamazing.cn/go-genreate-manager/pkg/cache"
	"llamazing.cn/go-genreate-manager/pkg/command"
	"llamazing.cn/go-genreate-manager/pkg/generator"
	"llamazing.cn/go-genreate-manager/pkg/hash"
)

func TestIntegration(t *testing.T) {
	// 检查必要的工具是否安装
	if _, err := exec.LookPath("mockgen"); err != nil {
		t.Skip("mockgen not installed, skipping tests")
	}

	// 获取测试数据的绝对路径
	testDataDir, err := filepath.Abs("testdata")
	if err != nil {
		t.Fatalf("failed to get absolute path: %v", err)
	}

	// 创建临时输出目录
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		dir     string
		cmd     string
		wantErr bool
		skip    bool
	}{
		{
			name:    "generate mocks for simple package",
			dir:     filepath.Join(testDataDir, "simple"),
			cmd:     "mockgen",
			wantErr: false,
		},
		{
			name:    "generate protos for simple package",
			dir:     filepath.Join(testDataDir, "simple"),
			cmd:     "protoc",
			wantErr: false,
			skip:    true, // 跳过 protoc 测试
		},
		{
			name:    "generate mocks for nested packages",
			dir:     filepath.Join(testDataDir, "nested"),
			cmd:     "mockgen",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("skipping test requiring protoc")
			}

			// 设置缓存文件路径
			cacheFile := filepath.Join(tmpDir, tt.cmd+".sum")
			cache := cache.NewFileCache(cacheFile)
			if err := cache.Load(); err != nil {
				t.Fatalf("load cache failed: %v", err)
			}

			// 创建生成器
			gen := generator.New(generator.Options{
				Hasher:  hash.NewContentHasher(),
				Cache:   cache,
				Finder:  command.NewFinder(tt.cmd),
				Workers: 1,
			})

			// 执行生成
			err := gen.Generate(context.Background(), tt.dir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// 验证生成的文件
			if err == nil {
				verifyGeneratedFiles(t, tt.dir, tt.cmd)
			}

			// 保存缓存
			if err := cache.Save(); err != nil {
				t.Errorf("save cache failed: %v", err)
			}
		})
	}
}

func verifyGeneratedFiles(t *testing.T, dir, cmd string) {
	var foundFiles []string

	// 递归查找所有生成的文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			filename := filepath.Base(path)
			// 检查是否是生成的文件
			if cmd == "mockgen" && strings.HasPrefix(filename, "mock_") && strings.HasSuffix(filename, ".go") {
				foundFiles = append(foundFiles, path)
			} else if cmd == "protoc" && strings.HasSuffix(filename, ".pb.go") {
				foundFiles = append(foundFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		t.Errorf("walk directory failed: %v", err)
		return
	}

	// 检查是否找到了生成的文件
	if len(foundFiles) == 0 {
		t.Errorf("no generated files found in %s for command %s", dir, cmd)
		return
	}

	// 打印找到的文件（用于调试）
	t.Logf("Found generated files:")
	for _, file := range foundFiles {
		relPath, _ := filepath.Rel(dir, file)
		t.Logf("  %s", relPath)
	}
}
