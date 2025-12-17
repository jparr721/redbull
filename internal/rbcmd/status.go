package rbcmd

import (
	"fmt"
	config "redbull"
)

type StatusCommand struct{}

func (c *StatusCommand) Help() string {
	return "Display beacon status information: status"
}

func (c *StatusCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	return fmt.Sprintf("Upstream: %s\nProxy: %s\nUsing KRB: %t\nSleep Time: %s", config.UPSTREAM, config.PROXY_URL, config.USE_KRB, *ctx.SleepTime), "", nil
}

