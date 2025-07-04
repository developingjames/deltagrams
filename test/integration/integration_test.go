package integration

import (
	"strings"
	"testing"

	"deltagram/internal/testutil"
	"deltagram/pkg/operations"
	"deltagram/pkg/parser"
)

func TestIntegration_FullDeltagramWorkflow(t *testing.T) {
	// Create test components
	parser := parser.NewParser()
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	
	// Add initial file structure
	fs.AddDir("/base/src")
	fs.AddFile("/base/src/original.py", []byte(`def hello():
    print("Hello")
    return True

def main():
    hello()`))
	
	// Complete deltagram with all operation types
	deltagramContent := `--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Integration test: creating, modifying, copying, moving files.
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/new_module.py
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: create

+++ src/new_module.py
class NewClass:
    def __init__(self):
        self.value = 42
    
    def get_value(self):
        return self.value
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/original.py
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -1,6 +1,8 @@
+import sys
+
 def hello():
-    print("Hello")
+    print("Hello, World!")
     return True
 
 def main():
+    print("Starting application...")
     hello()
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/backup.py
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: copy

--- src/original.py
+++ src/backup.py
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/renamed_module.py
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: move

--- src/new_module.py
+++ src/renamed_module.py
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--`

	// Parse deltagram
	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	// Apply deltagram
	err = applier.Apply(deltagram, "/base")
	if err != nil {
		t.Fatalf("Failed to apply deltagram: %v", err)
	}

	// Verify results
	files := fs.GetFiles()
	
	// Check original file was modified
	originalContent, exists := files["/base/src/original.py"]
	if !exists {
		t.Error("Original file should still exist")
	}
	expectedOriginal := `import sys

def hello():
    print("Hello, World!")
    return True

def main():
    print("Starting application...")
    hello()`
	if string(originalContent) != expectedOriginal {
		t.Errorf("Original file content mismatch.\nExpected:\n%s\n\nGot:\n%s", 
			expectedOriginal, string(originalContent))
	}

	// Check backup was created (copy of modified original)
	backupContent, exists := files["/base/src/backup.py"]
	if !exists {
		t.Error("Backup file should exist")
	}
	if string(backupContent) != expectedOriginal {
		t.Errorf("Backup file should match modified original")
	}

	// Check new module was created and then moved
	_, exists = files["/base/src/new_module.py"]
	if exists {
		t.Error("Original new module should not exist after move")
	}
	
	renamedContent, exists := files["/base/src/renamed_module.py"]
	if !exists {
		t.Error("Renamed module should exist")
	}
	expectedRenamed := `class NewClass:
    def __init__(self):
        self.value = 42
    
    def get_value(self):
        return self.value`
	if string(renamedContent) != expectedRenamed {
		t.Errorf("Renamed module content mismatch.\nExpected:\n%s\n\nGot:\n%s", 
			expectedRenamed, string(renamedContent))
	}

	// Verify expected file count
	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d: %v", len(files), files)
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	parser := parser.NewParser()
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	
	// Deltagram that tries to modify non-existent file
	deltagramContent := `--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: nonexistent.txt
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -1,1 +1,1 @@
-old
+new
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	err = applier.Apply(deltagram, "/base")
	if err == nil {
		t.Error("Expected error when trying to modify non-existent file")
	}
}

func TestIntegration_DeleteOperation(t *testing.T) {
	parser := parser.NewParser()
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	
	// Add initial files
	fs.AddFile("/base/file1.txt", []byte("content1"))
	fs.AddFile("/base/file2.txt", []byte("content2"))
	
	deltagramContent := `--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: file1.txt
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: delete

--- file1.txt
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	err = applier.Apply(deltagram, "/base")
	if err != nil {
		t.Fatalf("Failed to apply deltagram: %v", err)
	}

	files := fs.GetFiles()
	
	// file1.txt should be deleted
	if fs.FileExists("/base/file1.txt") {
		t.Error("file1.txt should have been deleted")
	}
	
	// file2.txt should still exist
	if !fs.FileExists("/base/file2.txt") {
		t.Error("file2.txt should still exist")
	}
	
	if len(files) != 1 {
		t.Errorf("Expected 1 file remaining, got %d", len(files))
	}
}

func TestIntegration_MimeogramBackwardCompatibility(t *testing.T) {
	parser := parser.NewParser()
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	
	fs.AddDir("/base/src")
	
	// Old mimeogram format without Delta-Operation headers
	mimeogramContent := `--====MIMEOGRAM_0123456789abcdef0123456789abcdef====
Content-Location: mimeogram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

This is a backward compatibility test.
--====MIMEOGRAM_0123456789abcdef0123456789abcdef====
Content-Location: src/hello.py
Content-Type: text/x-python; charset=utf-8; linesep=LF

def hello():
    print("Hello from mimeogram!")
--====MIMEOGRAM_0123456789abcdef0123456789abcdef====--`

	deltagram, err := parser.Parse(mimeogramContent)
	if err != nil {
		t.Fatalf("Failed to parse mimeogram: %v", err)
	}

	err = applier.Apply(deltagram, "/base")
	if err != nil {
		t.Fatalf("Failed to apply mimeogram: %v", err)
	}

	// Check file was created with default create operation
	content, err := fs.ReadFile("/base/src/hello.py")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := `def hello():
    print("Hello from mimeogram!")`
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestIntegration_ComplexDiffOperations(t *testing.T) {
	parser := parser.NewParser()
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	
	// Create a more complex file for testing
	originalContent := `#!/usr/bin/env python3
import os
import sys

class Calculator:
    def __init__(self):
        self.history = []
    
    def add(self, a, b):
        result = a + b
        self.history.append(f"{a} + {b} = {result}")
        return result
    
    def subtract(self, a, b):
        result = a - b
        self.history.append(f"{a} - {b} = {result}")
        return result

def main():
    calc = Calculator()
    print(calc.add(5, 3))
    print(calc.subtract(10, 4))

if __name__ == "__main__":
    main()`

	fs.AddFile("/base/calculator.py", []byte(originalContent))
	
	// Complex diff that adds imports, modifies methods, and adds new functionality
	deltagramContent := `--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: calculator.py
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -1,4 +1,6 @@
 #!/usr/bin/env python3
 import os
 import sys
+import math
+from typing import List
 
 class Calculator:
@@ -8,11 +10,17 @@
     
     def add(self, a, b):
         result = a + b
-        self.history.append(f"{a} + {b} = {result}")
+        self.history.append(f"ADD: {a} + {b} = {result}")
         return result
     
     def subtract(self, a, b):
         result = a - b
-        self.history.append(f"{a} - {b} = {result}")
+        self.history.append(f"SUB: {a} - {b} = {result}")
+        return result
+    
+    def multiply(self, a, b):
+        result = a * b
+        self.history.append(f"MUL: {a} * {b} = {result}")
         return result
 
 def main():
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	err = applier.Apply(deltagram, "/base")
	if err != nil {
		t.Fatalf("Failed to apply deltagram: %v", err)
	}

	// Verify the complex modifications
	modifiedContent, err := fs.ReadFile("/base/calculator.py")
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	modified := string(modifiedContent)
	
	// Check that new imports were added
	if !strings.Contains(modified, "import math") {
		t.Error("Should contain 'import math'")
	}
	if !strings.Contains(modified, "from typing import List") {
		t.Error("Should contain 'from typing import List'")
	}
	
	// Check that existing methods were modified
	if !strings.Contains(modified, "ADD: {a} + {b} = {result}") {
		t.Error("Add method should be modified with ADD prefix")
	}
	if !strings.Contains(modified, "SUB: {a} - {b} = {result}") {
		t.Error("Subtract method should be modified with SUB prefix")
	}
	
	// Check that new method was added
	if !strings.Contains(modified, "def multiply(self, a, b):") {
		t.Error("Should contain new multiply method")
	}
	if !strings.Contains(modified, "MUL: {a} * {b} = {result}") {
		t.Error("Multiply method should have correct implementation")
	}
}