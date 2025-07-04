package testutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MockFileSystem implements a fake file system for testing
type MockFileSystem struct {
	mu    sync.RWMutex
	files map[string][]byte
	dirs  map[string]bool
}

// NewMockFileSystem creates a new mock file system
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}
}

// AddFile adds a file to the mock file system
func (fs *MockFileSystem) AddFile(path string, content []byte) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	fs.files[path] = content
	
	// Ensure directories exist
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		fs.dirs[dir] = true
		dir = filepath.Dir(dir)
	}
}

// AddDir adds a directory to the mock file system
func (fs *MockFileSystem) AddDir(path string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.dirs[path] = true
}

// GetFiles returns all files in the mock file system
func (fs *MockFileSystem) GetFiles() map[string][]byte {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	result := make(map[string][]byte)
	for k, v := range fs.files {
		result[k] = make([]byte, len(v))
		copy(result[k], v)
	}
	return result
}

// FileExists checks if a file exists
func (fs *MockFileSystem) FileExists(path string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	_, exists := fs.files[path]
	return exists
}

// ReadFile reads a file from the mock file system
func (fs *MockFileSystem) ReadFile(filename string) ([]byte, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	content, exists := fs.files[filename]
	if !exists {
		return nil, fmt.Errorf("file not found: %s", filename)
	}
	
	result := make([]byte, len(content))
	copy(result, content)
	return result, nil
}

// WriteFile writes a file to the mock file system
func (fs *MockFileSystem) WriteFile(filename string, data []byte, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	// Ensure directory exists
	dir := filepath.Dir(filename)
	if dir != "." && dir != "/" && !fs.dirs[dir] {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	
	fs.files[filename] = make([]byte, len(data))
	copy(fs.files[filename], data)
	return nil
}

// Remove removes a file from the mock file system
func (fs *MockFileSystem) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if _, exists := fs.files[name]; !exists {
		return fmt.Errorf("file not found: %s", name)
	}
	
	delete(fs.files, name)
	return nil
}

// Rename renames a file in the mock file system
func (fs *MockFileSystem) Rename(oldpath, newpath string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	content, exists := fs.files[oldpath]
	if !exists {
		return fmt.Errorf("file not found: %s", oldpath)
	}
	
	// Ensure destination directory exists
	dir := filepath.Dir(newpath)
	if dir != "." && dir != "/" && !fs.dirs[dir] {
		return fmt.Errorf("directory does not exist: %s", dir)
	}
	
	fs.files[newpath] = content
	delete(fs.files, oldpath)
	return nil
}

// MkdirAll creates directories in the mock file system
func (fs *MockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	// Create all parent directories
	current := path
	for current != "." && current != "/" {
		fs.dirs[current] = true
		current = filepath.Dir(current)
	}
	return nil
}

// Stat returns file info for the mock file system
func (fs *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	if _, exists := fs.files[name]; exists {
		return &mockFileInfo{name: filepath.Base(name), isDir: false}, nil
	}
	
	if fs.dirs[name] {
		return &mockFileInfo{name: filepath.Base(name), isDir: true}, nil
	}
	
	return nil, os.ErrNotExist
}

// Open opens a file in the mock file system
func (fs *MockFileSystem) Open(name string) (io.ReadCloser, error) {
	content, err := fs.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &mockFile{content: content, reader: strings.NewReader(string(content))}, nil
}

// Create creates a file in the mock file system
func (fs *MockFileSystem) Create(name string) (io.WriteCloser, error) {
	return &mockWriteFile{fs: fs, name: name}, nil
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name  string
	isDir bool
}

func (fi *mockFileInfo) Name() string       { return fi.name }
func (fi *mockFileInfo) Size() int64        { return 0 }
func (fi *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (fi *mockFileInfo) ModTime() time.Time { return time.Time{} }
func (fi *mockFileInfo) IsDir() bool        { return fi.isDir }
func (fi *mockFileInfo) Sys() interface{}   { return nil }

// mockFile implements io.ReadCloser for testing
type mockFile struct {
	content []byte
	reader  *strings.Reader
}

func (f *mockFile) Read(p []byte) (n int, err error) {
	return f.reader.Read(p)
}

func (f *mockFile) Close() error {
	return nil
}

// mockWriteFile implements io.WriteCloser for testing
type mockWriteFile struct {
	fs     *MockFileSystem
	name   string
	buffer strings.Builder
}

func (f *mockWriteFile) Write(p []byte) (n int, err error) {
	return f.buffer.Write(p)
}

func (f *mockWriteFile) Close() error {
	return f.fs.WriteFile(f.name, []byte(f.buffer.String()), 0644)
}