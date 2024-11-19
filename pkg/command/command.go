package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"llamazing.cn/go-genreate-manager/pkg/generator"
)

// GoGenCommand 实现了 generator.Command 接口
type GoGenCommand struct {
	filePath string
	cmdStr   string
}

func NewCommand(path, cmdStr string) *GoGenCommand {
	return &GoGenCommand{
		filePath: path,
		cmdStr:   cmdStr,
	}
}

func (c *GoGenCommand) Execute(ctx context.Context) error {
	args := strings.Fields(c.cmdStr)
	if len(args) == 0 {
		return nil
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = filepath.Dir(c.filePath)

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("execute command failed: %s: %w", out, err)
	}

	return nil
}

func (c *GoGenCommand) GetFilePath() string {
	return c.filePath
}

func (c *GoGenCommand) String() string {
	return c.cmdStr
}

// CommandFinder 实现命令查找功能
type CommandFinder struct {
	pattern string
}

func NewFinder(pattern string) generator.CommandFinder {
	return &CommandFinder{pattern: pattern}
}

func (f *CommandFinder) Find(dir string) ([]generator.Command, error) {
	var commands []generator.Command
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".go" {
			cmd, err := f.findInFile(path)
			if err != nil {
				return err
			}
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
		return nil
	})
	return commands, err
}

func (f *CommandFinder) findInFile(path string) (generator.Command, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(fmt.Sprintf(`//go:generate (%s.*)`, regexp.QuoteMeta(f.pattern)))
	matches := re.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return NewCommand(path, matches[1]), nil
	}
	return nil, nil
}
