package operations

import (
	"testing"

	"github.com/developingjames/deltagrams/internal/testutil"
	"github.com/developingjames/deltagrams/pkg/parser"
)

func TestCopyHandler_Apply(t *testing.T) {
	handler := NewCopyHandler()
	fs := testutil.NewMockFileSystem()

	// Create source file
	fs.AddFile("/base/source.txt", []byte("Original content\nLine 2"))
	fs.AddDir("/base/dest")

	part := parser.DeltagramPart{
		ContentLocation: "dest/copied.txt",
		ContentType:     "application/x-deltagram-fileop; charset=utf-8",
		DeltaOperation:  "copy",
		Content:         "--- source.txt\n+++ dest/copied.txt",
	}

	err := handler.Apply(fs, "/base", part)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check original file still exists
	if !fs.FileExists("/base/source.txt") {
		t.Error("Expected source file to still exist")
	}

	// Check copied file exists
	if !fs.FileExists("/base/dest/copied.txt") {
		t.Error("Expected copied file to exist")
	}

	// Check copied file content
	content, err := fs.ReadFile("/base/dest/copied.txt")
	if err != nil {
		t.Fatalf("Failed to read copied file: %v", err)
	}

	expected := "Original content\nLine 2"
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestCopyHandler_Apply_SourceNotExists(t *testing.T) {
	handler := NewCopyHandler()
	fs := testutil.NewMockFileSystem()

	part := parser.DeltagramPart{
		ContentLocation: "dest/copied.txt",
		ContentType:     "application/x-deltagram-fileop; charset=utf-8",
		DeltaOperation:  "copy",
		Content:         "--- nonexistent.txt\n+++ dest/copied.txt",
	}

	err := handler.Apply(fs, "/base", part)
	if err == nil {
		t.Error("Expected error for nonexistent source file, got none")
	}
}

func TestCopyHandler_Apply_InvalidContent(t *testing.T) {
	handler := NewCopyHandler()
	fs := testutil.NewMockFileSystem()

	part := parser.DeltagramPart{
		ContentLocation: "dest/copied.txt",
		ContentType:     "application/x-deltagram-fileop; charset=utf-8",
		DeltaOperation:  "copy",
		Content:         "invalid content without source/dest markers",
	}

	err := handler.Apply(fs, "/base", part)
	if err == nil {
		t.Error("Expected error for invalid content, got none")
	}
}

func TestCopyHandler_CanHandle(t *testing.T) {
	handler := NewCopyHandler()

	tests := []struct {
		operation string
		expected  bool
	}{
		{"copy", true},
		{"create", false},
		{"delete", false},
		{"move", false},
		{"content", false},
	}

	for _, test := range tests {
		result := handler.CanHandle(test.operation)
		if result != test.expected {
			t.Errorf("CanHandle(%q) = %v, expected %v", test.operation, result, test.expected)
		}
	}
}
