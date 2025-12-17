package rbcmd

import (
	"fmt"
	"strconv"
	"time"
)

type SleepCommand struct{}

func (c *SleepCommand) Help() string {
	return "Set the sleep time between check-ins (in seconds): sleep <seconds>"
}

func (c *SleepCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	// Check if the cmd can be turned into an int
	sleepTimeInt, err := strconv.Atoi(cmd)
	if err != nil {
		return "", "", fmt.Errorf("failed to convert sleep time to int: %w", err)
	}

	*ctx.SleepTime = time.Duration(sleepTimeInt) * time.Second
	return fmt.Sprintf("set sleep time to %s", *ctx.SleepTime), "", nil
}

