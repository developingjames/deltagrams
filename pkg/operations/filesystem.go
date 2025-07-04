package operations

import (
	"io"
	"os"
)

// RealFileSystem implements FileSystem using actual OS operations
type RealFileSystem struct{}

// NewRealFileSystem creates a new real file system
func NewRealFileSystem() FileSystem {
	return &RealFileSystem{}
}

func (fs *RealFileSystem) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func (fs *RealFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filename, data, perm)
}

func (fs *RealFileSystem) Remove(name string) error {
	return os.Remove(name)
}

func (fs *RealFileSystem) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (fs *RealFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *RealFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *RealFileSystem) Open(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func (fs *RealFileSystem) Create(name string) (io.WriteCloser, error) {
	return os.Create(name)
}