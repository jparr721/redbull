package rbcmd

import (
	"fmt"
	"os"
	"strings"
)

type LsCommand struct{}

func (c *LsCommand) Help() string {
	return "List directory contents: ls <path>"
}

func (c *LsCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	path := cmd
	if path == "" {
		path = *ctx.CWD
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read directory %s: %w", *ctx.CWD, err)
	}

	fileNames := make([]string, 0)
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}

	return strings.Join(fileNames, "\n"), "", nil
}
