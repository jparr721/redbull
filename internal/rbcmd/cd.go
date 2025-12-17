package rbcmd

import (
	"fmt"
	"os"
	"path/filepath"
)

type CdCommand struct{}

func (c *CdCommand) Help() string {
	return "Change directory: cd <path>"
}

func (c *CdCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	// Join the current directory with the command path
	newCwd := filepath.Join(*ctx.CWD, cmd)

	// Resolve to absolute path (handles relative paths like ../..)
	absPath, err := filepath.Abs(newCwd)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve path '%s': %w", newCwd, err)
	}

	res, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return "", "", fmt.Errorf("invalid path '%s': path does not exist", absPath)
	}

	if !res.IsDir() {
		return "", "", fmt.Errorf("invalid path '%s': not a directory", absPath)
	}

	*ctx.CWD = absPath
	return fmt.Sprintf("changed directory to %s", *ctx.CWD), "", nil
}

