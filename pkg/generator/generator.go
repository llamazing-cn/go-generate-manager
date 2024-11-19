package generator

import (
	"context"
	"fmt"
	"sync"
)

type DefaultGenerator struct {
	hasher  FileHasher
	cache   Cache
	finder  CommandFinder
	workers int
}

// New 创建新的生成器实例
func New(opts Options) Generator {
	if opts.Workers <= 0 {
		opts.Workers = 1
	}

	return &DefaultGenerator{
		hasher:  opts.Hasher,
		cache:   opts.Cache,
		finder:  opts.Finder,
		workers: opts.Workers,
	}
}

// Generate 实现代码生成逻辑
func (g *DefaultGenerator) Generate(ctx context.Context, dir string) error {
	// 1. 查找所有命令
	commands, err := g.finder.Find(dir)
	if err != nil {
		return fmt.Errorf("find commands: %w", err)
	}

	// 2. 并发处理文件
	var wg sync.WaitGroup
	errChan := make(chan error, len(commands))
	semaphore := make(chan struct{}, g.workers)

	for _, cmd := range commands {
		wg.Add(1)
		go func(cmd Command) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				errChan <- ctx.Err()
				return
			}

			if err := g.processCommand(ctx, cmd); err != nil {
				errChan <- err
			}
		}(cmd)
	}

	// 等待所有任务完成
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 收集错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("generate failed with %d errors: %v", len(errs), errs[0])
	}

	return nil
}

func (g *DefaultGenerator) processCommand(ctx context.Context, cmd Command) error {
	path := cmd.GetFilePath()

	// 1. 检查文件是否需要重新生成
	oldHash, exists := g.cache.Get(path)
	if !exists || g.hasher.IsChanged(path, oldHash) {
		// 2. 执行命令
		if err := cmd.Execute(ctx); err != nil {
			return fmt.Errorf("execute command: %w", err)
		}

		// 3. 更新缓存
		newHash, err := g.hasher.Hash(path)
		if err != nil {
			return fmt.Errorf("calculate hash: %w", err)
		}
		g.cache.Set(path, newHash)
	}

	return nil
}
