package operations

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/developingjames/deltagrams/pkg/parser"
)

// CreateHandler handles file creation operations
type CreateHandler struct{}

// NewCreateHandler creates a new create handler
func NewCreateHandler() OperationHandler {
	return &CreateHandler{}
}

// CanHandle returns true if this handler can process the given operation
func (h *CreateHandler) CanHandle(operation string) bool {
	return operation == "create" || operation == ""
}

// Apply creates a new file with the specified content
func (h *CreateHandler) Apply(fs FileSystem, baseDir string, part parser.DeltagramPart) error {
	filePath := ResolveFilePath(baseDir, part.ContentLocation)

	// Parse create operation content
	lines := strings.Split(part.Content, "\n")
	var content string
	var contentStarted bool

	for _, line := range lines {
		if strings.HasPrefix(line, "+++") {
			contentStarted = true
			continue
		}
		if contentStarted {
			if content != "" {
				content += "\n"
			}
			content += line
		}
	}

	// If no +++ marker found, use entire content
	if !contentStarted {
		content = part.Content
	}

	// Ensure directory exists
	if err := fs.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write file content
	if err := fs.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("Created: %s\n", part.ContentLocation)
	return nil
}
