package main

import (
	"fmt"
	"os"

	"github.com/developingjames/deltagrams/pkg/clipboard"
	"github.com/developingjames/deltagrams/pkg/operations"
	"github.com/developingjames/deltagrams/pkg/parser"
)

// Version information (set by build flags)
var (
	Version    = "dev"
	CommitHash = "unknown"
	BuildTime  = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		showUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "apply":
		if err := applyDeltagram(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "version", "--version", "-v":
		showVersion()
	case "help", "--help", "-h":
		showUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		showUsage()
		os.Exit(1)
	}
}

func applyDeltagram() error {
	// Create dependencies
	clipboardReader := clipboard.NewReader()
	parser := parser.NewParser()
	fs := operations.NewRealFileSystem()
	applier := operations.NewApplier(fs)

	// Read deltagram from clipboard
	content, err := clipboardReader.Read()
	if err != nil {
		return fmt.Errorf("failed to read clipboard: %v", err)
	}

	// Parse deltagram
	deltagram, err := parser.Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse deltagram: %v", err)
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	// Apply deltagram to current directory
	if err := applier.Apply(deltagram, cwd); err != nil {
		return fmt.Errorf("failed to apply deltagram: %v", err)
	}

	fmt.Println("Deltagram applied successfully")
	return nil
}

func showUsage() {
	fmt.Println("Usage: deltagram <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  apply           Apply deltagram from clipboard to current directory")
	fmt.Println("  version, -v     Show version information")
	fmt.Println("  help, -h        Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  deltagram apply    # Apply deltagram from clipboard")
	fmt.Println("  deltagram version  # Show version")
}

func showVersion() {
	fmt.Printf("deltagram %s\n", Version)
	fmt.Printf("  commit: %s\n", CommitHash)
	fmt.Printf("  built:  %s\n", BuildTime)
}