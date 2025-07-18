package clipboard

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Reader defines the interface for reading from clipboard
type Reader interface {
	Read() (string, error)
}

// DefaultReader implements clipboard reading for multiple platforms
type DefaultReader struct{}

// NewReader creates a new clipboard reader
func NewReader() Reader {
	return &DefaultReader{}
}

// Read reads content from the system clipboard
func (r *DefaultReader) Read() (string, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("powershell", "-command", "Get-Clipboard")
	case "darwin":
		cmd = exec.Command("pbpaste")
	case "linux":
		// Try xclip first, then xsel as fallback
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard", "-o")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--output")
		} else {
			return "", fmt.Errorf("clipboard access requires xclip or xsel on Linux")
		}
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute clipboard command: %v", err)
	}

	return strings.TrimSpace(string(output)), nil
}
