package rbcmd

import (
	"fmt"
	"os"
)

type UploadCommand struct{}

func (c *UploadCommand) Help() string {
	return "Upload a file to the server: upload <file>"
}

func (c *UploadCommand) Execute(ctx *Context, cmd string) (string, string, error) {
	// First, read the file from the command line
	file, err := os.Open(cmd)
	if err != nil {
		return "", "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Then, read the file contents
	// contents, err := io.ReadAll(file)
	// if err != nil {
	// 	return "", "", fmt.Errorf("failed to read file: %w", err)
	// }

	// //

	return fmt.Sprintf("uploaded file %s", cmd), "", nil
}
