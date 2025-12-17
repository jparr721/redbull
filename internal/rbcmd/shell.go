package rbcmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ShellCommand struct{}

func (c *ShellCommand) Help() string {
	return "Execute a shell command: shell <command>"
}

func (c *ShellCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	timeout := 100 * time.Second
	ctxTimeout, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	execCmd := exec.CommandContext(ctxTimeout, "bash", "-c", cmd)

	var outBuf, errBuf bytes.Buffer
	execCmd.Stdout = &outBuf
	execCmd.Stderr = &errBuf

	err := execCmd.Run()
	if ctxTimeout.Err() == context.DeadlineExceeded {
		return outBuf.String(), errBuf.String(), fmt.Errorf("connection timed out after %v", timeout)
	}

	return outBuf.String(), errBuf.String(), err
}

