package operations

import (
	"strings"
	"testing"

	"github.com/developingjames/deltagrams/internal/testutil"
	"github.com/developingjames/deltagrams/pkg/parser"
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
	
	expectedMsg := "cannot apply content operation to non-existent file"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
	}
}

func TestContentHandler_Apply_HunkBeyondFileEnd(t *testing.T) {
	handler := NewContentHandler()
	fs := testutil.NewMockFileSystem()
	
	// Create a short file
	originalContent := "line 1\nline 2"
	fs.AddFile("/base/short.txt", []byte(originalContent))
	
	// Try to apply diff that references line 10
	part := parser.DeltagramPart{
		ContentLocation: "short.txt",
		ContentType:     "application/x-deltagram-content; charset=utf-8; linesep=LF",
		DeltaOperation:  "content",
		Content:         "@@ -10,1 +10,1 @@\n-nonexistent\n+replacement",
	}

	err := handler.Apply(fs, "/base", part)
	if err == nil {
		t.Error("Expected error for hunk beyond file end, got none")
	}
	
	expectedMsg := "hunk refers to line 10 but original file has 2 lines"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
	}
}

func TestContentHandler_Apply_RemoveLineBeyondFileEnd(t *testing.T) {
	handler := NewContentHandler()
	fs := testutil.NewMockFileSystem()
	
	// Create a short file
	originalContent := "line 1\nline 2"
	fs.AddFile("/base/short.txt", []byte(originalContent))
	
	// Try to remove more lines than exist
	part := parser.DeltagramPart{
		ContentLocation: "short.txt",
		ContentType:     "application/x-deltagram-content; charset=utf-8; linesep=LF",
		DeltaOperation:  "content",
		Content:         "@@ -1,5 +1,1 @@\n-line 1\n-line 2\n-line 3\n-line 4\n-line 5\n+single line",
	}

	err := handler.Apply(fs, "/base", part)
	if err == nil {
		t.Error("Expected error for removing too many lines, got none")
	}
	
	expectedMsg := "line to remove extends beyond original file"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error message to contain %q, got: %v", expectedMsg, err)
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

func TestContentHandler_Apply_MultiHunk(t *testing.T) {
	handler := NewContentHandler()
	fs := testutil.NewMockFileSystem()
	
	// Create file with content similar to the deltagram issue
	originalContent := `### Target Lengths by File Type
- **Core Concept Files**: 800-1,500 words max
- **Faction/Organization Files**: 1,000-2,000 words max  
- **Character Files**: 300-800 words max
- **Technology/System Files**: 600-1,200 words max
- **Reference/Summary Files**: 500-1,000 words max

### Regular Audits
- **Quarterly Review**: Check for file bloat and redundancy
- **Story Relevance**: Cut details that don't serve current or planned narrative
- **Consolidation Opportunities**: Merge overlapping files`
	
	fs.AddFile("/base/test.md", []byte(originalContent))
	
	part := parser.DeltagramPart{
		ContentLocation: "test.md",
		ContentType:     "application/x-deltagram-content; charset=utf-8; linesep=LF",
		DeltaOperation:  "content",
		Content: `@@ -1,5 +1,5 @@
 ### Target Lengths by File Type
 - **Core Concept Files**: 800-1,500 words max
 - **Faction/Organization Files**: 1,000-2,000 words max  
-- **Character Files**: 300-800 words max
+- **Character Files**: 400-900 words max
 - **Technology/System Files**: 600-1,200 words max
@@ -6,0 +6,1 @@
+- **Worldbuilding Index Files**: 300-600 words max
 - **Reference/Summary Files**: 500-1,000 words max
@@ -8,3 +9,4 @@
 ### Regular Audits
 - **Quarterly Review**: Check for file bloat and redundancy
+- **Monthly Quick Scan**: Identify files exceeding target lengths
 - **Story Relevance**: Cut details that don't serve current or planned narrative
 - **Consolidation Opportunities**: Merge overlapping files`,
	}

	err := handler.Apply(fs, "/base", part)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check file content
	content, err := fs.ReadFile("/base/test.md")
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expected := `### Target Lengths by File Type
- **Core Concept Files**: 800-1,500 words max
- **Faction/Organization Files**: 1,000-2,000 words max  
- **Character Files**: 400-900 words max
- **Technology/System Files**: 600-1,200 words max
- **Worldbuilding Index Files**: 300-600 words max
- **Reference/Summary Files**: 500-1,000 words max

### Regular Audits
- **Quarterly Review**: Check for file bloat and redundancy
- **Monthly Quick Scan**: Identify files exceeding target lengths
- **Story Relevance**: Cut details that don't serve current or planned narrative
- **Consolidation Opportunities**: Merge overlapping files`

	if string(content) != expected {
		t.Errorf("Expected content:\n%q\n\nGot:\n%q", expected, string(content))
	}
}