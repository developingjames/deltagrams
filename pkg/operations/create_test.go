package operations

import (
	"testing"

	"github.com/developingjames/deltagrams/internal/testutil"
	"github.com/developingjames/deltagrams/pkg/parser"
)

func TestCreateHandler_Apply(t *testing.T) {
	handler := NewCreateHandler()
	fs := testutil.NewMockFileSystem()
	
	// Ensure directory structure exists
	fs.AddDir("test")
	
	part := parser.DeltagramPart{
		ContentLocation: "test/hello.txt",
		ContentType:     "text/plain; charset=utf-8; linesep=LF",
		DeltaOperation:  "create",
		Content:         "+++ test/hello.txt\nHello, World!\nSecond line",
	}

	err := handler.Apply(fs, "/base", part)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check file was created
	if !fs.FileExists("/base/test/hello.txt") {
		t.Error("Expected file to be created")
	}

	// Check file content
	content, err := fs.ReadFile("/base/test/hello.txt")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := "Hello, World!\nSecond line"
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestCreateHandler_Apply_NoMarker(t *testing.T) {
	handler := NewCreateHandler()
	fs := testutil.NewMockFileSystem()
	
	// Ensure directory structure exists
	fs.AddDir("test")
	
	part := parser.DeltagramPart{
		ContentLocation: "test/simple.txt",
		ContentType:     "text/plain; charset=utf-8; linesep=LF",
		DeltaOperation:  "create",
		Content:         "Simple content without marker",
	}

	err := handler.Apply(fs, "/base", part)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check file content
	content, err := fs.ReadFile("/base/test/simple.txt")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := "Simple content without marker"
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestCreateHandler_CanHandle(t *testing.T) {
	handler := NewCreateHandler()
	
	tests := []struct {
		operation string
		expected  bool
	}{
		{"create", true},
		{"", true}, // Default case
		{"delete", false},
		{"move", false},
		{"copy", false},
		{"content", false},
	}

	for _, test := range tests {
		result := handler.CanHandle(test.operation)
		if result != test.expected {
			t.Errorf("CanHandle(%q) = %v, expected %v", test.operation, result, test.expected)
		}
	}
}