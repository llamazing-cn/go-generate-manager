package generator

import "context"

// Generator 定义代码生成器的核心接口
type Generator interface {
	Generate(ctx context.Context, dir string) error
}

// FileHasher 定义文件哈希计算接口
type FileHasher interface {
	Hash(path string) (string, error)
	IsChanged(path, oldHash string) bool
}

// Cache 定义缓存接口
type Cache interface {
	Load() error
	Save() error
	Get(path string) (string, bool)
	Set(path, hash string)
}

// Command 定义命令接口
type Command interface {
	Execute(ctx context.Context) error
	GetFilePath() string
	String() string
}

// CommandFinder 定义命令查找接口
type CommandFinder interface {
	Find(dir string) ([]Command, error)
}

// Options 定义生成器配置选项
type Options struct {
	Hasher  FileHasher
	Cache   Cache
	Finder  CommandFinder
	Workers int
}
