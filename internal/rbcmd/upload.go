package rbcmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	config "redbull"
	"strings"
)

type UploadCommand struct{}

func (c *UploadCommand) Help() string {
	return "Upload a file from the server: upload <filename> <desired-filename>"
}

func (c *UploadCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	commandGroups := strings.Split(cmd, " ")
	if len(commandGroups) != 2 {
		return "", "", fmt.Errorf("invalid command: %s", cmd)
	}

	filename := commandGroups[0]
	desiredFilename := commandGroups[1]

	// Fetch the file from the server
	fileUrl := fmt.Sprintf("%s/files/%s", config.UPSTREAM, filename)
	response, err := ctx.HttpClient.Get(fileUrl)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch file from server: %w", err)
	}
	defer response.Body.Close()

	// Check for HTTP errors
	if response.StatusCode != 200 {
		return "", "", fmt.Errorf("server returned status %d for file '%s' (requested: %s)", response.StatusCode, cmd, fileUrl)
	}

	// Read the file contents
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file contents: %w", err)
	}

	// Save the file to the current working directory with UUID filename first
	tempFilePath := filepath.Join(*ctx.CWD, filename)
	destFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer destFile.Close()

	// Write the contents to the file
	_, err = destFile.Write(contents)
	if err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}
	destFile.Close() // Close before renaming

	// Rename to the desired filename
	desiredFilePath := filepath.Join(*ctx.CWD, desiredFilename)
	if err := os.Rename(tempFilePath, desiredFilePath); err != nil {
		return "", "", fmt.Errorf("failed to rename file: %w", err)
	}

	return fmt.Sprintf("uploaded file %s as %s", filename, desiredFilename), "", nil
}
