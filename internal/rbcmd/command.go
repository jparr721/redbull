package rbcmd

import (
	"redbull/internal/rbhttp"
	"time"
)

// Command defines the interface that all commands must implement
type Command interface {
	Help() string
	Execute(ctx *Context, cmd string) (string, string, error)
}

// Context holds shared state that commands can access and modify
type Context struct {
	CWD        *string
	SleepTime  *time.Duration
	HttpClient rbhttp.HttpClient
}

// Registry is a map of command names to Command implementations
type Registry map[string]Command

var registry Registry

func init() {
	registry = Registry{
		"shell":    &ShellCommand{},
		"pwd":      &PwdCommand{},
		"cd":       &CdCommand{},
		"ls":       &LsCommand{},
		"status":   &StatusCommand{},
		"sleep":    &SleepCommand{},
		"help":     &HelpCommand{},
		"download": &DownloadCommand{},
		"upload":   &UploadCommand{},
	}
}

// GetRegistry returns the command registry
func GetRegistry() Registry {
	return registry
}
