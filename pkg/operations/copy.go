package operations

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"deltagram/pkg/parser"
)

// CopyHandler handles file copy operations
type CopyHandler struct{}

// NewCopyHandler creates a new copy handler
func NewCopyHandler() OperationHandler {
	return &CopyHandler{}
}

// CanHandle returns true if this handler can process the given operation
func (h *CopyHandler) CanHandle(operation string) bool {
	return operation == "copy"
}

// Apply copies a file from source to destination
func (h *CopyHandler) Apply(fs FileSystem, baseDir string, part parser.DeltagramPart) error {
	// Parse copy operation content to get source and destination
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
		return fmt.Errorf("invalid copy operation: missing source or destination path")
	}
	
	sourceFullPath := ResolveFilePath(baseDir, sourcePath)
	destFullPath := ResolveFilePath(baseDir, destPath)
	
	// Ensure destination directory exists
	if err := fs.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}
	
	if err := h.copyFile(fs, sourceFullPath, destFullPath); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	fmt.Printf("Copied: %s -> %s\n", sourcePath, destPath)
	return nil
}

func (h *CopyHandler) copyFile(fs FileSystem, src, dst string) error {
	sourceFile, err := fs.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := fs.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}