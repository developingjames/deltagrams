package operations

import (
	"fmt"
	"os"

	"github.com/developingjames/deltagrams/pkg/parser"
)

// DeleteHandler handles file deletion operations
type DeleteHandler struct{}

// NewDeleteHandler creates a new delete handler
func NewDeleteHandler() OperationHandler {
	return &DeleteHandler{}
}

// CanHandle returns true if this handler can process the given operation
func (h *DeleteHandler) CanHandle(operation string) bool {
	return operation == "delete"
}

// Apply deletes the specified file
func (h *DeleteHandler) Apply(fs FileSystem, baseDir string, part parser.DeltagramPart) error {
	filePath := ResolveFilePath(baseDir, part.ContentLocation)
	
	if err := fs.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: File %s does not exist (already deleted)\n", part.ContentLocation)
			return nil
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	fmt.Printf("Deleted: %s\n", part.ContentLocation)
	return nil
}