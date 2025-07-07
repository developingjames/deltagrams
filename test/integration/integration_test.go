package integration

import (
	"strings"
	"testing"

	"github.com/developingjames/deltagrams/internal/testutil"
	"github.com/developingjames/deltagrams/pkg/operations"
	"github.com/developingjames/deltagrams/pkg/parser"
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

func TestIntegration_FlexibleBoundaryIdentifiers(t *testing.T) {
	parser := parser.NewParser()
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)

	fs.AddDir("/base/src")

	// Use the flexible boundary identifier format from the example
	identifier := "voice456sample789012345678901234ef"
	deltagramContent := `--====DELTAGRAM_` + identifier + `====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Testing flexible boundary identifiers with non-UUID format.
--====DELTAGRAM_` + identifier + `====
Content-Location: src/test.py
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: create

+++ src/test.py
def test_function():
    return "success"
--====DELTAGRAM_` + identifier + `====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram with flexible identifier: %v", err)
	}

	// Verify the identifier was parsed correctly
	if deltagram.UUID != identifier {
		t.Errorf("Expected identifier '%s', got '%s'", identifier, deltagram.UUID)
	}

	err = applier.Apply(deltagram, "/base")
	if err != nil {
		t.Fatalf("Failed to apply deltagram with flexible identifier: %v", err)
	}

	// Check file was created successfully
	content, err := fs.ReadFile("/base/src/test.py")
	if err != nil {
		t.Fatalf("Failed to read created file: %v", err)
	}

	expected := `def test_function():
    return "success"`
	if string(content) != expected {
		t.Errorf("Expected content %q, got %q", expected, string(content))
	}
}

func TestIntegration_ContentOperations_MultiHunk(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	parser := parser.NewParser()

	// Create initial configuration file
	initialContent := `# API Configuration Guide
# Last updated: 2024-01-15

## 🔧 Server Settings

### Connection Limits
- **Max Connections**: 1000 per server
- **Connection Pool**: 50-100 connections
- **Timeout Settings**: 30 seconds default
- **Retry Attempts**: 3 maximum retries
- **Buffer Size**: 8KB standard

### Performance Monitoring
- **Health Checks**: Every 60 seconds
- **Metrics Collection**: Real-time data gathering
- **Alert Thresholds**: 95% CPU, 90% memory

## 📊 Database Configuration

### Connection Settings
- **Primary DB**: PostgreSQL cluster
- **Read Replicas**: 3 instances minimum
- **Connection Timeout**: 10 seconds
- **Query Timeout**: 30 seconds

### Optimization Guidelines
- **Index Strategy**: Composite indexes preferred
- **Query Caching**: 1-hour TTL for static data
- **Batch Processing**: 500 records per batch

## 🔐 Security Configuration

Security settings and authentication rules go here.`

	fs.AddFile("/test/config/api_guide.md", []byte(initialContent))

	// Apply multi-hunk content operations
	deltagramContent := `--====DELTAGRAM_api_config_update_v2====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Updating API configuration with new limits, monitoring features, and security section.
Changes include: updated connection limits, added real-time monitoring, enhanced security.
--====DELTAGRAM_api_config_update_v2====
Content-Location: config/api_guide.md
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -6,5 +6,5 @@
 ### Connection Limits
 - **Max Connections**: 1000 per server
 - **Connection Pool**: 50-100 connections
-- **Timeout Settings**: 30 seconds default
+- **Timeout Settings**: 45 seconds default
 - **Retry Attempts**: 3 maximum retries
 - **Buffer Size**: 8KB standard
@@ -11,0 +11,1 @@
 - **Buffer Size**: 8KB standard
+- **Load Balancing**: Round-robin algorithm
 
@@ -13,3 +14,4 @@
 ### Performance Monitoring
 - **Health Checks**: Every 60 seconds
+- **Real-time Alerts**: Instant notification system
 - **Metrics Collection**: Real-time data gathering
 - **Alert Thresholds**: 95% CPU, 90% memory
@@ -30,0 +32,12 @@
 
+---
+
+## 🛡️ Enhanced Security Features
+
+### Authentication Methods
+- **OAuth 2.0**: Primary authentication protocol
+- **API Keys**: Secondary access method for services
+- **JWT Tokens**: 15-minute expiration for sessions
+
+### Access Control
+- **Role-based Permissions**: Admin, User, Read-only levels
+- **IP Whitelisting**: Restrict access by network location
+
 ## 🔐 Security Configuration
--====DELTAGRAM_api_config_update_v2====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	err = applier.Apply(deltagram, "/test")
	if err != nil {
		t.Fatalf("Failed to apply deltagram: %v", err)
	}

	// Verify the content was modified correctly
	resultContent, err := fs.ReadFile("/test/config/api_guide.md")
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	result := string(resultContent)

	// Verify key changes were applied
	if !strings.Contains(result, "- **Timeout Settings**: 45 seconds default") {
		t.Error("Timeout setting was not updated correctly")
	}
	if !strings.Contains(result, "- **Load Balancing**: Round-robin algorithm") {
		t.Error("Load balancing line was not inserted")
	}
	if !strings.Contains(result, "- **Real-time Alerts**: Instant notification system") {
		t.Error("Real-time alerts line was not inserted")
	}
	if !strings.Contains(result, "## 🛡️ Enhanced Security Features") {
		t.Error("Enhanced security section was not added")
	}
	if !strings.Contains(result, "- **OAuth 2.0**: Primary authentication protocol") {
		t.Error("OAuth authentication line was not added")
	}

	// Verify structure is maintained
	if !strings.Contains(result, "## 🔐 Security Configuration") {
		t.Error("Original security section should still be present")
	}
}

func TestIntegration_ContentOperations_LineEndingTolerance(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	parser := parser.NewParser()

	// Create file with Windows line endings (CRLF)
	initialContent := "# Development Setup\r\n\r\n## Prerequisites\r\n- Node.js version 18+\r\n- npm or yarn package manager\r\n- Git for version control\r\n\r\n## Installation Steps\r\nFollow these steps to set up the project.\r\n"
	fs.AddFile("/test/setup.md", []byte(initialContent))

	// Apply deltagram with Unix line endings (LF)
	deltagramContent := `--====DELTAGRAM_setup_updates====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Updating development setup requirements.
--====DELTAGRAM_setup_updates====
Content-Location: setup.md
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -3,4 +3,5 @@
 ## Prerequisites
 - Node.js version 18+
+- Python 3.9+ for build scripts
 - npm or yarn package manager
 - Git for version control
--====DELTAGRAM_setup_updates====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	// Should succeed despite line ending differences
	err = applier.Apply(deltagram, "/test")
	if err != nil {
		t.Fatalf("Failed to apply deltagram with mixed line endings: %v", err)
	}

	// Verify the change was applied
	resultContent, err := fs.ReadFile("/test/setup.md")
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	result := string(resultContent)
	if !strings.Contains(result, "- Python 3.9+ for build scripts") {
		t.Error("Python requirement was not inserted correctly")
	}
}

func TestIntegration_ContentOperations_PureInsertion(t *testing.T) {
	fs := testutil.NewMockFileSystem()
	applier := operations.NewApplier(fs)
	parser := parser.NewParser()

	// Create simple file
	initialContent := `# Project Changelog

## Version 1.0.0
- Initial release
- Basic functionality implemented

## Future Plans
Upcoming features and improvements.`

	fs.AddFile("/test/CHANGELOG.md", []byte(initialContent))

	// Apply pure insertion (OldCount=0)
	deltagramContent := `--====DELTAGRAM_changelog_update====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Adding new version entry to changelog.
--====DELTAGRAM_changelog_update====
Content-Location: CHANGELOG.md
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -5,0 +5,5 @@
 - Basic functionality implemented
+
+## Version 1.1.0
+- Added user authentication
+- Improved error handling
+- Performance optimizations
 
 ## Future Plans
--====DELTAGRAM_changelog_update====--`

	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	err = applier.Apply(deltagram, "/test")
	if err != nil {
		t.Fatalf("Failed to apply pure insertion deltagram: %v", err)
	}

	// Verify the insertion was applied correctly
	resultContent, err := fs.ReadFile("/test/CHANGELOG.md")
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	result := string(resultContent)

	// Check that new version was inserted
	if !strings.Contains(result, "## Version 1.1.0") {
		t.Error("New version section was not inserted")
	}
	if !strings.Contains(result, "- Added user authentication") {
		t.Error("Authentication feature was not added")
	}

	// Verify original content is preserved
	if !strings.Contains(result, "## Version 1.0.0") {
		t.Error("Original version section was lost")
	}
	if !strings.Contains(result, "## Future Plans") {
		t.Error("Future plans section was lost")
	}
}
