package tests

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/llamazing-cn/go-generate-manager/pkg/cache"
	"github.com/llamazing-cn/go-generate-manager/pkg/command"
	"github.com/llamazing-cn/go-generate-manager/pkg/generator"
	"github.com/llamazing-cn/go-generate-manager/pkg/hash"
)

func BenchmarkGenerate(b *testing.B) {
	// 获取 GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		b.Skip("GOPATH not set")
	}

	// 在 GOPATH/src 下创建测试目录
	testDir := filepath.Join(gopath, "src", "llamazing.cn/go-genreate-manager/tests/testdata/bench")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		b.Fatalf("failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	setupTestFiles(b, testDir)

	workerCounts := []int{1, 2, 4, 8, 16, 32, 64, 128}
	for _, w := range workerCounts {
		b.Run(fmt.Sprintf("workers-%d", w), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				cacheFile := filepath.Join(testDir, fmt.Sprintf("mockgen-%d.sum", i))
				cache := cache.NewFileCache(cacheFile)

				gen := generator.New(generator.Options{
					Hasher:  hash.NewContentHasher(),
					Cache:   cache,
					Finder:  command.NewFinder("mockgen"),
					Workers: w,
				})

				err := gen.Generate(context.Background(), testDir)
				if err != nil {
					b.Fatalf("generation failed: %v", err)
				}
			}
		})
	}
}

func setupTestFiles(b *testing.B, dir string) {
	// 创建20个包，每个包有5个接口文件
	for i := 1; i <= 20; i++ {
		pkgName := fmt.Sprintf("pkg%d", i)
		files := map[string]string{
			filepath.Join(pkgName, "service.go"): fmt.Sprintf(`package %s

import (
    "context"
    "time"
)

//go:generate mockgen -source=service.go -destination=mock_service.go -package=%s
type Service interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    GetStatus() (string, error)
    Configure(config map[string]interface{}) error
    UpdateMetrics(metrics map[string]float64) error
    HandleRequest(req interface{}) (interface{}, error)
    ValidateInput(input interface{}) error
    ProcessOutput(output interface{}) error
    GetVersion() string
    IsHealthy() bool
}`, pkgName, pkgName),

			filepath.Join(pkgName, "repository.go"): fmt.Sprintf(`package %s

import (
    "context"
    "time"
)

//go:generate mockgen -source=repository.go -destination=mock_repository.go -package=%s
type Repository interface {
    Find(ctx context.Context, id string) (interface{}, error)
    FindAll(ctx context.Context, filter map[string]interface{}) ([]interface{}, error)
    Create(ctx context.Context, data interface{}) error
    Update(ctx context.Context, id string, data interface{}) error
    Delete(ctx context.Context, id string) error
    Count(ctx context.Context, filter map[string]interface{}) (int64, error)
    Exists(ctx context.Context, id string) (bool, error)
    Transaction(ctx context.Context, fn func(tx interface{}) error) error
    Backup(ctx context.Context) error
    Restore(ctx context.Context, backup interface{}) error
}`, pkgName, pkgName),

			filepath.Join(pkgName, "handler.go"): fmt.Sprintf(`package %s

import (
    "context"
    "time"
)

//go:generate mockgen -source=handler.go -destination=mock_handler.go -package=%s
type Handler interface {
    Handle(ctx context.Context, event interface{}) error
    HandleBatch(ctx context.Context, events []interface{}) []error
    ValidateEvent(event interface{}) error
    ProcessEvent(event interface{}) (interface{}, error)
    GetEventType() string
    IsEventValid(event interface{}) bool
    GetEventMetadata(event interface{}) map[string]interface{}
    SetEventHandler(handlerFunc func(interface{}) error)
    GetEventHistory() []interface{}
    ClearEventHistory() error
}`, pkgName, pkgName),

			filepath.Join(pkgName, "cache.go"): fmt.Sprintf(`package %s

import (
    "context"
    "time"
)

//go:generate mockgen -source=cache.go -destination=mock_cache.go -package=%s
type Cache interface {
    Get(ctx context.Context, key string) (interface{}, error)
    GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    SetMulti(ctx context.Context, values map[string]interface{}, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
    DeleteMulti(ctx context.Context, keys []string) error
    Exists(ctx context.Context, key string) (bool, error)
    Increment(ctx context.Context, key string, delta int64) (int64, error)
    Flush(ctx context.Context) error
    GetStats() map[string]interface{}
}`, pkgName, pkgName),

			filepath.Join(pkgName, "worker.go"): fmt.Sprintf(`package %s

import (
    "context"
    "time"
)

//go:generate mockgen -source=worker.go -destination=mock_worker.go -package=%s
type WorkerStatus int

const (
    StatusIdle WorkerStatus = iota
    StatusBusy
    StatusError
)

type Worker interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Process(ctx context.Context, job interface{}) error
    ProcessBatch(ctx context.Context, jobs []interface{}) []error
    GetStatus() WorkerStatus
    GetMetrics() map[string]float64
    SetConcurrency(n int) error
    IsBusy() bool
    WaitForCompletion(ctx context.Context) error
    Reset() error
}`, pkgName, pkgName),
		}

		for path, content := range files {
			fullPath := filepath.Join(dir, path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				b.Fatalf("failed to create directory: %v", err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				b.Fatalf("failed to write file: %v", err)
			}
		}
	}
}
