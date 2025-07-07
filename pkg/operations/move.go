package operations

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/developingjames/deltagrams/pkg/parser"
)

// MoveHandler handles file move/rename operations
type MoveHandler struct{}

// NewMoveHandler creates a new move handler
func NewMoveHandler() OperationHandler {
	return &MoveHandler{}
}

// CanHandle returns true if this handler can process the given operation
func (h *MoveHandler) CanHandle(operation string) bool {
	return operation == "move"
}

// Apply moves/renames a file from source to destination
func (h *MoveHandler) Apply(fs FileSystem, baseDir string, part parser.DeltagramPart) error {
	// Parse move operation content to get source and destination
	lines := strings.Split(part.Content, "\n")
	var sourcePath, destPath string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "---") {
			sourcePath = strings.TrimSpace(strings.TrimPrefix(line, "---"))
		} else if strings.HasPrefix(line, "+++") {
			destPath = strings.TrimSpace(strings.TrimPrefix(line, "+++"))
		}
	}

	if sourcePath == "" || destPath == "" {
		return fmt.Errorf("invalid move operation: missing source or destination path")
	}

	sourceFullPath := ResolveFilePath(baseDir, sourcePath)
	destFullPath := ResolveFilePath(baseDir, destPath)

	// Ensure destination directory exists
	if err := fs.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	if err := fs.Rename(sourceFullPath, destFullPath); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	fmt.Printf("Moved: %s -> %s\n", sourcePath, destPath)
	return nil
}
