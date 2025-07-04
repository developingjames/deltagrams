package operations

import (
	"path/filepath"
	"strings"
)

// ResolveFilePath converts various path formats to a resolved file path
func ResolveFilePath(baseDir, filePath string) string {
	// Handle URLs or absolute paths by converting to relative
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		// Extract filename from URL
		parts := strings.Split(filePath, "/")
		filePath = parts[len(parts)-1]
	} else if filepath.IsAbs(filePath) {
		// Convert absolute path to relative by removing leading slash
		filePath = strings.TrimPrefix(filePath, "/")
	}

	return filepath.Join(baseDir, filePath)
}