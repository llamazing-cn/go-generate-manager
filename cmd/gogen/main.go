package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/llamazing-cn/go-generate-manager/pkg/cache"
	"github.com/llamazing-cn/go-generate-manager/pkg/command"
	"github.com/llamazing-cn/go-generate-manager/pkg/generator"
	"github.com/llamazing-cn/go-generate-manager/pkg/hash"
)

const usage = `Usage: gogen [options]

Options:
  -d, --dir     <path>     directory to generate files
  -c, --cmd     <command>  command to use for code generation
  -o, --output  <path>     directory to output files
  -w, --workers <number>   number of worker goroutines (default: number of CPUs)
  -h, --help              show this help message

Example:
  gogen -d ./src -c mockgen -o ./gen
  gogen --dir=./src --cmd=mockgen --output=./gen --workers=4
`

func main() {
	log.Println("starting generation process")
	start := time.Now()

	cfg := parseFlags()
	if cfg == nil {
		return
	}

	if cfg.dir == "..." {
		var err error
		cfg.dir, err = os.Getwd()
		if err != nil {
			log.Fatalf("get working directory failed: %v", err)
		}
	}

	if cfg.output == "" {
		cfg.output = cfg.dir
	}

	cacheFile := filepath.Join(cfg.output, cfg.cmd+".sum")
	cache := cache.NewFileCache(cacheFile)
	if err := cache.Load(); err != nil {
		log.Fatalf("load cache failed: %v", err)
	}
	defer func() {
		if err := cache.Save(); err != nil {
			log.Printf("save cache failed: %v", err)
		}
	}()

	gen := generator.New(generator.Options{
		Hasher:  hash.NewContentHasher(),
		Cache:   cache,
		Finder:  command.NewFinder(cfg.cmd),
		Workers: cfg.workers,
	})

	ctx := context.Background()
	if err := gen.Generate(ctx, cfg.dir); err != nil {
		log.Fatalf("generation failed: %v", err)
	}

	elapsed := time.Since(start)
	log.Printf("generation completed in %s", elapsed)
}

type config struct {
	dir     string
	cmd     string
	output  string
	workers int
	help    bool
}

func parseFlags() *config {
	cfg := &config{}

	flag.StringVar(&cfg.dir, "d", "...", "directory to generate files (use ... for current directory)")
	flag.StringVar(&cfg.cmd, "c", "", "command to use for code generation")
	flag.StringVar(&cfg.output, "o", "", "directory to output files (default: same as source)")
	workers := runtime.NumCPU() * 2
	if workers > 8 {
		workers = 8
	}
	flag.IntVar(&cfg.workers, "w", workers, "number of worker goroutines")
	flag.BoolVar(&cfg.help, "h", false, "show help message")

	flag.StringVar(&cfg.dir, "dir", "...", "directory to generate files (use ... for current directory)")
	flag.StringVar(&cfg.cmd, "cmd", "", "command to use for code generation")
	flag.StringVar(&cfg.output, "output", "", "directory to output files (default: same as source)")
	flag.IntVar(&cfg.workers, "workers", runtime.NumCPU()*4, "number of worker goroutines")
	flag.BoolVar(&cfg.help, "help", false, "show help message")

	flag.Usage = func() {
		log.Print(usage)
	}

	flag.Parse()

	if cfg.help {
		flag.Usage()
		return nil
	}

	if !cfg.validate() {
		flag.Usage()
		return nil
	}

	return cfg
}

func (c *config) validate() bool {
	if c.cmd == "" {
		log.Println("Error: required flag -cmd must be set")
		return false
	}

	if c.workers < 1 {
		log.Printf("Warning: invalid worker count %d, using 1 instead", c.workers)
		c.workers = 1
	}

	return true
}
