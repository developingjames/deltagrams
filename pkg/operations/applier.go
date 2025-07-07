package operations

import (
	"fmt"
	"strings"

	"github.com/developingjames/deltagrams/pkg/parser"
)

// DefaultApplier implements the Applier interface
type DefaultApplier struct {
	fs       FileSystem
	handlers []OperationHandler
}

// NewApplier creates a new applier with the given file system
func NewApplier(fs FileSystem) Applier {
	applier := &DefaultApplier{
		fs: fs,
	}

	// Register default handlers
	applier.handlers = []OperationHandler{
		NewCreateHandler(),
		NewDeleteHandler(),
		NewCopyHandler(),
		NewMoveHandler(),
		NewContentHandler(),
	}

	return applier
}

// Apply applies a deltagram to the specified base directory
func (a *DefaultApplier) Apply(deltagram *parser.Deltagram, baseDir string) error {
	// Process operations in the order they appear
	for _, part := range deltagram.Parts {
		// Skip message parts
		if part.ContentLocation == "mimeogram://message" || part.ContentLocation == "deltagram://message" {
			fmt.Printf("Message: %s\n", strings.TrimSpace(part.Content))
			continue
		}

		// Find appropriate handler
		var handler OperationHandler
		for _, h := range a.handlers {
			if h.CanHandle(part.DeltaOperation) {
				handler = h
				break
			}
		}

		if handler == nil {
			// Default to create for backward compatibility
			handler = NewCreateHandler()
		}

		if err := handler.Apply(a.fs, baseDir, part); err != nil {
			return fmt.Errorf("failed to apply %s operation to %s: %v", part.DeltaOperation, part.ContentLocation, err)
		}
	}

	return nil
}
