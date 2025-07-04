package operations

import (
	"io"
	"os"

	"deltagram/pkg/parser"
)

// FileSystem abstracts file system operations for testing
type FileSystem interface {
	ReadFile(filename string) ([]byte, error)
	WriteFile(filename string, data []byte, perm os.FileMode) error
	Remove(name string) error
	Rename(oldpath, newpath string) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Open(name string) (io.ReadCloser, error)
	Create(name string) (io.WriteCloser, error)
}

// Applier defines the interface for applying deltagram operations
type Applier interface {
	Apply(deltagram *parser.Deltagram, baseDir string) error
}

// OperationHandler handles specific types of operations
type OperationHandler interface {
	CanHandle(operation string) bool
	Apply(fs FileSystem, baseDir string, part parser.DeltagramPart) error
}