package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/developingjames/deltagrams/pkg/operations"
	"github.com/developingjames/deltagrams/pkg/parser"
)

// TestAutomaticOffsetCalculation tests the core scenario where LLMs generate
// deltagrams using original file line numbers, and the system automatically
// handles offset calculations for multiple hunks.
func TestAutomaticOffsetCalculation(t *testing.T) {
	// Create a temporary directory for the test
	tempDir := t.TempDir()

	// Original file content - simulates a typical file an LLM would work with
	originalContent := `package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello World")
	
	result := calculate(5, 3)
	fmt.Printf("Result: %d\n", result)
}

func calculate(a, b int) int {
	return a + b
}

func helper() {
	log.Println("Helper called")
}

// End of file`

	// Write the original file
	filePath := filepath.Join(tempDir, "example.go")
	err := os.WriteFile(filePath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a deltagram that makes multiple changes using ORIGINAL line numbers
	// This simulates what an LLM would generate - it references the original file
	// without manually calculating offsets for subsequent hunks
	deltagramContent := `--====DELTAGRAM_offset_test_12345678====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Integration test: Multiple content operations using original file line numbers.
Tests automatic offset calculation across multiple hunks.
--====DELTAGRAM_offset_test_12345678====
Content-Location: example.go
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -3,2 +3,3 @@
 import (
 	"fmt"
+	"errors"
 	"log"
@@ -8,2 +9,4 @@
 func main() {
+	// Initialize the application
 	fmt.Println("Hello World")
+	fmt.Println("Starting calculations...")
 	
@@ -15,1 +18,3 @@
 func calculate(a, b int) int {
+	if a < 0 || b < 0 {
+		return 0
+	}
 	return a + b
@@ -19,2 +24,4 @@
 func helper() {
+	fmt.Println("Debug: Helper function called")
 	log.Println("Helper called")
+	fmt.Println("Debug: Helper function completed")
--====DELTAGRAM_offset_test_12345678====--`

	// Parse and apply the deltagram
	parser := parser.NewParser()
	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	// Apply the deltagram
	fs := operations.NewRealFileSystem()
	applier := operations.NewApplier(fs)

	err = applier.Apply(deltagram, tempDir)
	if err != nil {
		t.Fatalf("Failed to apply deltagram: %v", err)
	}

	// Read the modified file
	modifiedContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	// Expected result after all automatic offset calculations
	expectedContent := `package main

import (
	"fmt"
	"errors"
	"log"
)

func main() {
	// Initialize the application
	fmt.Println("Hello World")
	fmt.Println("Starting calculations...")
	
	result := calculate(5, 3)
	fmt.Printf("Result: %d\n", result)
}

func calculate(a, b int) int {
	if a < 0 || b < 0 {
		return 0
	}
	return a + b
}

func helper() {
	fmt.Println("Debug: Helper function called")
	log.Println("Helper called")
	fmt.Println("Debug: Helper function completed")
}

// End of file`

	// Compare the results
	if strings.TrimSpace(string(modifiedContent)) != strings.TrimSpace(expectedContent) {
		t.Errorf("Automatic offset calculation failed.\n\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(modifiedContent))
	}
}

// TestOffsetCalculationWithComplexChanges tests more complex scenarios including
// removals, additions, and mixed operations across multiple hunks
func TestOffsetCalculationWithComplexChanges(t *testing.T) {
	tempDir := t.TempDir()

	// Original file with various sections
	originalContent := `# Project Configuration

## Database Settings
host = localhost
port = 5432
database = myapp
user = admin
password = secret

## API Settings
endpoint = https://api.example.com
timeout = 30
retries = 3

## Cache Settings
enabled = true
ttl = 3600
provider = redis

## Logging
level = info
format = json`

	filePath := filepath.Join(tempDir, "config.txt")
	err := os.WriteFile(filePath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Deltagram with complex changes using original line numbers
	deltagramContent := `--====DELTAGRAM_complex_test_87654321====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Complex integration test: Removals, additions, and modifications using original line numbers.
--====DELTAGRAM_complex_test_87654321====
Content-Location: config.txt
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -4,3 +4,4 @@
 ## Database Settings
 host = localhost
+port_backup = 5433
-port = 5432
+port = 3306
 database = myapp
@@ -8,1 +9,0 @@
-password = secret
@@ -12,3 +12,3 @@
 endpoint = https://api.example.com
+version = v2
 timeout = 30
-retries = 3
+retries = 5
@@ -18,1 +18,3 @@
 provider = redis
+host = localhost:6379
+cluster = false
--====DELTAGRAM_complex_test_87654321====--`

	// Parse and apply
	parser := parser.NewParser()
	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse complex deltagram: %v", err)
	}

	fs := operations.NewRealFileSystem()
	applier := operations.NewApplier(fs)

	err = applier.Apply(deltagram, tempDir)
	if err != nil {
		t.Fatalf("Failed to apply complex deltagram: %v", err)
	}

	// Read and verify result
	modifiedContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	expectedContent := `# Project Configuration

## Database Settings
host = localhost
port_backup = 5433
port = 3306
database = myapp
user = admin

## API Settings
endpoint = https://api.example.com
version = v2
timeout = 30
retries = 5

## Cache Settings
enabled = true
ttl = 3600
provider = redis
host = localhost:6379
cluster = false

## Logging
level = info
format = json`

	if strings.TrimSpace(string(modifiedContent)) != strings.TrimSpace(expectedContent) {
		t.Errorf("Complex offset calculation failed.\n\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(modifiedContent))
	}
}

// TestMultipleHunksOriginalLineNumbers tests a complex scenario with multiple hunks
// using original file line numbers to verify automatic offset calculation works correctly
func TestMultipleHunksOriginalLineNumbers(t *testing.T) {
	tempDir := t.TempDir()

	// Generic document content for testing
	originalContent := `# Software Component Documentation
#component #system #architecture

## Overview
This component handles user authentication and session management. It provides secure login functionality and maintains user state across application sessions.

## Features
The system includes **multi-factor authentication** and supports various authentication methods including password-based and token-based approaches.

## Implementation
Clean, modular design with clear separation of concerns. Follows established security patterns and best practices.

## Dependencies
- Core authentication library for password hashing
- Session management utilities for state persistence  
- Token validation services for API access

## Security Model
🔒 Defense in Depth / 🛡️ Zero Trust
Assumes all requests are potentially malicious until proven otherwise. Every operation requires explicit authentication and authorization.

## Core Principle
> "Security through transparency and verification, not obscurity."

## Known Issues
- Memory usage increases with concurrent sessions
- Memory usage increases with concurrent sessions

## Design Goals
This system prioritizes security and reliability over performance. The goal is to provide bulletproof authentication that scales with user growth while maintaining strict security standards.

## Implementation Notes

### Error Handling
> **System**: "Authentication failed. Please check your credentials and try again. If the problem persists, contact system administration."

### Logging Strategy  
> **System**: "All authentication attempts are logged for security auditing. Successful logins are recorded with session details for compliance tracking."

### Monitoring Approach
> **System**: "Real-time monitoring tracks failed login attempts, unusual access patterns, and potential security threats for immediate response."

**Technical Notes**: Uses industry-standard encryption protocols with comprehensive audit logging. Implements rate limiting and suspicious activity detection. Designed for high availability and fault tolerance.

## Development Timeline
- **Phase 1:** Initial authentication framework implementation with basic security features.
- **Phase 2:** Enhanced security features including multi-factor authentication and session management.
- **Phase 3:** Advanced monitoring and analytics capabilities with automated threat detection.
- **Phase 4:** Full deployment with comprehensive security monitoring and compliance reporting.`

	filePath := filepath.Join(tempDir, "component.md")
	err := os.WriteFile(filePath, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Deltagram with multiple hunks using original line numbers
	deltagramContent := `--====DELTAGRAM_generic_test_12345678====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Test deltagram with multiple content changes using original line numbers:
- Fix duplicate line in Known Issues section
- Add new Integration section
- Enhance Error Handling sample
- Add Current Status section after Overview
- Expand Development Timeline with more detail
--====DELTAGRAM_generic_test_12345678====
Content-Location: component.md
Content-Type: application/x-deltagram-content; charset=utf-8; linesep=LF
Delta-Operation: content

@@ -5,6 +5,10 @@
 ## Overview
 This component handles user authentication and session management. It provides secure login functionality and maintains user state across application sessions.
 
+## Current Status
+Production-ready system actively serving authentication requests. Currently handling enterprise-scale deployments with comprehensive monitoring and automated security responses.
+
 ## Features
 The system includes **multi-factor authentication** and supports various authentication methods including password-based and token-based approaches.
 
@@ -26,7 +30,13 @@
 
 ## Known Issues
 - Memory usage increases with concurrent sessions
-- Memory usage increases with concurrent sessions
+- Rate limiting may affect legitimate high-frequency users
+- Token refresh requires network connectivity for validation
+
+## Integration
+- **External Services**: Connects with third-party identity providers for federated authentication
+- **Internal Systems**: Integrates with user management and audit logging components  
+- **API Gateway**: Provides authentication tokens for downstream service authorization
 
 ## Design Goals
 This system prioritizes security and reliability over performance. The goal is to provide bulletproof authentication that scales with user growth while maintaining strict security standards.
@@ -35,6 +45,8 @@
 
 ### Error Handling
 > **System**: "Authentication failed. Please check your credentials and try again. If the problem persists, contact system administration."
+> 
+> **System**: "For security reasons, detailed error information is available in the system logs accessible to administrators only."
 
 ### Logging Strategy  
 > **System**: "All authentication attempts are logged for security auditing. Successful logins are recorded with session details for compliance tracking."
@@ -46,7 +58,9 @@
 
 ## Development Timeline
 - **Phase 1:** Initial authentication framework implementation with basic security features.
-- **Phase 2:** Enhanced security features including multi-factor authentication and session management.
+- **Phase 2:** Enhanced security features including multi-factor authentication and session management. Added comprehensive audit logging and monitoring capabilities.
 - **Phase 3:** Advanced monitoring and analytics capabilities with automated threat detection.
-- **Phase 4:** Full deployment with comprehensive security monitoring and compliance reporting.
+- **Phase 4:** Full deployment with comprehensive security monitoring and compliance reporting. Established disaster recovery procedures and high-availability configuration.
+
+**Evolution Notes**: Transitioned from basic authentication to enterprise-grade security platform with advanced threat detection and automated incident response capabilities.
--====DELTAGRAM_generic_test_12345678====--`

	// Parse and apply
	parser := parser.NewParser()
	deltagram, err := parser.Parse(deltagramContent)
	if err != nil {
		t.Fatalf("Failed to parse deltagram: %v", err)
	}

	fs := operations.NewRealFileSystem()
	applier := operations.NewApplier(fs)

	err = applier.Apply(deltagram, tempDir)
	if err != nil {
		t.Fatalf("Failed to apply deltagram: %v", err)
	}

	// Read and verify the result
	modifiedContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read modified file: %v", err)
	}

	// Expected result after automatic offset calculation
	expectedContent := `# Software Component Documentation
#component #system #architecture

## Overview
This component handles user authentication and session management. It provides secure login functionality and maintains user state across application sessions.

## Current Status
Production-ready system actively serving authentication requests. Currently handling enterprise-scale deployments with comprehensive monitoring and automated security responses.

## Features
The system includes **multi-factor authentication** and supports various authentication methods including password-based and token-based approaches.

## Implementation
Clean, modular design with clear separation of concerns. Follows established security patterns and best practices.

## Dependencies
- Core authentication library for password hashing
- Session management utilities for state persistence  
- Token validation services for API access

## Security Model
🔒 Defense in Depth / 🛡️ Zero Trust
Assumes all requests are potentially malicious until proven otherwise. Every operation requires explicit authentication and authorization.

## Core Principle
> "Security through transparency and verification, not obscurity."

## Known Issues
- Memory usage increases with concurrent sessions
- Rate limiting may affect legitimate high-frequency users
- Token refresh requires network connectivity for validation

## Integration
- **External Services**: Connects with third-party identity providers for federated authentication
- **Internal Systems**: Integrates with user management and audit logging components  
- **API Gateway**: Provides authentication tokens for downstream service authorization

## Design Goals
This system prioritizes security and reliability over performance. The goal is to provide bulletproof authentication that scales with user growth while maintaining strict security standards.

## Implementation Notes

### Error Handling
> **System**: "Authentication failed. Please check your credentials and try again. If the problem persists, contact system administration."
> 
> **System**: "For security reasons, detailed error information is available in the system logs accessible to administrators only."

### Logging Strategy  
> **System**: "All authentication attempts are logged for security auditing. Successful logins are recorded with session details for compliance tracking."

### Monitoring Approach
> **System**: "Real-time monitoring tracks failed login attempts, unusual access patterns, and potential security threats for immediate response."

**Technical Notes**: Uses industry-standard encryption protocols with comprehensive audit logging. Implements rate limiting and suspicious activity detection. Designed for high availability and fault tolerance.

## Development Timeline
- **Phase 1:** Initial authentication framework implementation with basic security features.
- **Phase 2:** Enhanced security features including multi-factor authentication and session management. Added comprehensive audit logging and monitoring capabilities.
- **Phase 3:** Advanced monitoring and analytics capabilities with automated threat detection.
- **Phase 4:** Full deployment with comprehensive security monitoring and compliance reporting. Established disaster recovery procedures and high-availability configuration.

**Evolution Notes**: Transitioned from basic authentication to enterprise-grade security platform with advanced threat detection and automated incident response capabilities.`

	if strings.TrimSpace(string(modifiedContent)) != strings.TrimSpace(expectedContent) {
		t.Errorf("Multiple hunks with original line numbers test failed.\n\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(modifiedContent))
	}

	t.Logf("✅ Multiple hunks with original line numbers test passed - automatic offset calculation working correctly!")
}
