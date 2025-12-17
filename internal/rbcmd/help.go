package rbcmd

import (
	"fmt"
)

type HelpCommand struct{}

func (c *HelpCommand) Help() string {
	return "Display help information: help"
}

func (c *HelpCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	help := "Available commands:\n"
	registry := GetRegistry()
	for name, command := range registry {
		help += fmt.Sprintf("  %s - %s\n", name, command.Help())
	}
	return help, "", nil
}

