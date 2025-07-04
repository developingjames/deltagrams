package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: deltagram <command>")
		fmt.Println("Commands:")
		fmt.Println("  apply    Apply deltagram from clipboard to current directory")
		fmt.Println("  test     Test deltagram parser with test file")
		fmt.Println("  test-apply Test apply operation with test file")
		fmt.Println("  test-full  Test all delta operations (create/modify/move/copy/delete)")
		fmt.Println("  test-content Test content modification with unified diff")
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "apply":
		if err := applyDeltagram(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "test":
		testParser()
	case "test-apply":
		testApply()
	case "test-full":
		testFullOperations()
	case "test-content":
		testContentOperations()
	case "test-simple":
		testSimpleOperations()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func applyDeltagram() error {
	// Read deltagram from clipboard
	content, err := readClipboard()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %v", err)
	}

	// Parse deltagram
	deltagram, err := parseDeltagram(content)
	if err != nil {
		return fmt.Errorf("failed to parse deltagram: %v", err)
	}

	// Apply deltagram to current directory
	if err := applyToDirectory(deltagram); err != nil {
		return fmt.Errorf("failed to apply deltagram: %v", err)
	}

	fmt.Println("Deltagram applied successfully")
	return nil
}

func testParser() {
	fmt.Println("Testing deltagram parser...")
	
	// Read test file
	content, err := os.ReadFile("test_deltagram_full.txt")
	if err != nil {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	// Parse deltagram
	deltagram, err := parseDeltagram(string(content))
	if err != nil {
		fmt.Printf("Error parsing deltagram: %v\n", err)
		return
	}

	fmt.Printf("Successfully parsed deltagram with UUID: %s\n", deltagram.UUID)
	fmt.Printf("Number of parts: %d\n", len(deltagram.Parts))
	
	for i, part := range deltagram.Parts {
		fmt.Printf("Part %d: %s (%s) [%s]\n", i+1, part.ContentLocation, part.ContentType, part.DeltaOperation)
	}
}

func testApply() {
	fmt.Println("Testing apply operation...")
	
	// Read test file
	content, err := os.ReadFile("test_deltagram.txt")
	if err != nil {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	// Parse deltagram
	deltagram, err := parseDeltagram(string(content))
	if err != nil {
		fmt.Printf("Error parsing deltagram: %v\n", err)
		return
	}

	// Create test directory
	testDir := "test_output"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("Error creating test directory: %v\n", err)
		return
	}

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("Error changing to test directory: %v\n", err)
		return
	}
	
	defer os.Chdir(oldDir)

	// Apply deltagram
	if err := applyToDirectory(deltagram); err != nil {
		fmt.Printf("Error applying deltagram: %v\n", err)
		return
	}

	// Verify files were created
	expectedFiles := []string{"test/hello.txt", "src/main.go"}
	for _, file := range expectedFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Error: Expected file %s was not created\n", file)
			return
		}
		
		// Read and display file content
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file %s: %v\n", file, err)
			return
		}
		
		fmt.Printf("File %s:\n%s\n", file, string(content))
	}
	
	fmt.Println("Apply operation test completed successfully!")
}

func testFullOperations() {
	fmt.Println("Testing full delta operations...")
	
	// Read test file
	content, err := os.ReadFile("test_deltagram_full.txt")
	if err != nil {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	// Parse deltagram
	deltagram, err := parseDeltagram(string(content))
	if err != nil {
		fmt.Printf("Error parsing deltagram: %v\n", err)
		return
	}

	// Create test directory
	testDir := "test_full_output"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("Error creating test directory: %v\n", err)
		return
	}

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("Error changing to test directory: %v\n", err)
		return
	}
	
	defer os.Chdir(oldDir)

	// Apply deltagram
	if err := applyToDirectory(deltagram); err != nil {
		fmt.Printf("Error applying deltagram: %v\n", err)
		return
	}
	
	fmt.Println("Full operations test completed successfully!")
}

func testContentOperations() {
	fmt.Println("Testing content modification operations...")
	
	// Read test file
	content, err := os.ReadFile("test_content_diff.txt")
	if err != nil {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	// Parse deltagram
	deltagram, err := parseDeltagram(string(content))
	if err != nil {
		fmt.Printf("Error parsing deltagram: %v\n", err)
		return
	}

	// Create test directory
	testDir := "test_content_output"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("Error creating test directory: %v\n", err)
		return
	}

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("Error changing to test directory: %v\n", err)
		return
	}
	
	defer os.Chdir(oldDir)

	// Apply deltagram
	if err := applyToDirectory(deltagram); err != nil {
		fmt.Printf("Error applying deltagram: %v\n", err)
		return
	}
	
	// Show final file content
	finalContent, err := os.ReadFile("src/example.py")
	if err != nil {
		fmt.Printf("Error reading final file: %v\n", err)
		return
	}
	
	fmt.Println("Final file content:")
	fmt.Println(string(finalContent))
	fmt.Println("Content operations test completed successfully!")
}

func testSimpleOperations() {
	fmt.Println("Testing simple copy operation...")
	
	// Read test file
	content, err := os.ReadFile("test_simple.txt")
	if err != nil {
		fmt.Printf("Error reading test file: %v\n", err)
		return
	}

	// Parse deltagram
	deltagram, err := parseDeltagram(string(content))
	if err != nil {
		fmt.Printf("Error parsing deltagram: %v\n", err)
		return
	}

	// Create test directory
	testDir := "test_simple_output"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("Error creating test directory: %v\n", err)
		return
	}

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting current directory: %v\n", err)
		return
	}
	
	if err := os.Chdir(testDir); err != nil {
		fmt.Printf("Error changing to test directory: %v\n", err)
		return
	}
	
	defer os.Chdir(oldDir)

	// Apply deltagram
	if err := applyToDirectory(deltagram); err != nil {
		fmt.Printf("Error applying deltagram: %v\n", err)
		return
	}
	
	// Verify both files exist
	files := []string{"test/source.txt", "test/copied.txt"}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", file, err)
			return
		}
		fmt.Printf("File %s:\n%s\n", file, string(content))
	}
	
	fmt.Println("Simple operations test completed successfully!")
}