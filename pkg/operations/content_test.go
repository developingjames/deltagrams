package operations

import (
	"testing"

	"deltagram/internal/testutil"
	"deltagram/pkg/parser"
)

func TestContentHandler_Apply(t *testing.T) {
	handler := NewContentHandler()
	fs := testutil.NewMockFileSystem()
	
	// Create initial file
	originalContent := `def hello():
    print("Hello")
    return True

def main():
    hello()`
	
	fs.AddFile("/base/src/example.py", []byte(originalContent))
	
	part := parser.DeltagramPart{
		ContentLocation: "src/example.py",
		ContentType:     "application/x-deltagram-content; charset=utf-8; linesep=LF",
		DeltaOperation:  "content",
		Content: `@@ -1,6 +1,8 @@
+import sys
+
 def hello():
-    print("Hello")
+    print("Hello, World!")
     return True
 
 def main():
+    print("Starting...")
     hello()`,
	}

	err := handler.Apply(fs, "/base", part)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check file content
	content, err := fs.ReadFile("/base/src/example.py")
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := `import sys

def hello():
    print("Hello, World!")
    return True

def main():
    print("Starting...")
    hello()`

	if string(content) != expected {
		t.Errorf("Expected content:\n%q\n\nGot:\n%q", expected, string(content))
	}
}

func TestContentHandler_Apply_FileNotExists(t *testing.T) {
	handler := NewContentHandler()
	fs := testutil.NewMockFileSystem()
	
	part := parser.DeltagramPart{
		ContentLocation: "nonexistent.txt",
		ContentType:     "application/x-deltagram-content; charset=utf-8; linesep=LF",
		DeltaOperation:  "content",
		Content:         "@@ -1,1 +1,1 @@\n-old\n+new",
	}

	err := handler.Apply(fs, "/base", part)
	if err == nil {
		t.Error("Expected error for nonexistent file, got none")
	}
}

func TestContentHandler_parseHunkHeader(t *testing.T) {
	handler := &ContentHandler{}
	
	tests := []struct {
		line     string
		expected *HunkHeader
		hasError bool
	}{
		{
			line: "@@ -1,5 +1,8 @@",
			expected: &HunkHeader{
				OldStart: 1, OldCount: 5,
				NewStart: 1, NewCount: 8,
			},
		},
		{
			line: "@@ -1 +1,2 @@",
			expected: &HunkHeader{
				OldStart: 1, OldCount: 1,
				NewStart: 1, NewCount: 2,
			},
		},
		{
			line: "@@ -10,3 +15 @@",
			expected: &HunkHeader{
				OldStart: 10, OldCount: 3,
				NewStart: 15, NewCount: 1,
			},
		},
		{
			line:     "invalid header",
			hasError: true,
		},
	}

	for _, test := range tests {
		result, err := handler.parseHunkHeader(test.line)
		
		if test.hasError {
			if err == nil {
				t.Errorf("Expected error for line %q, got none", test.line)
			}
			continue
		}
		
		if err != nil {
			t.Errorf("Unexpected error for line %q: %v", test.line, err)
			continue
		}
		
		if result.OldStart != test.expected.OldStart ||
			result.OldCount != test.expected.OldCount ||
			result.NewStart != test.expected.NewStart ||
			result.NewCount != test.expected.NewCount {
			t.Errorf("For line %q, expected %+v, got %+v", 
				test.line, test.expected, result)
		}
	}
}