package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func applyToDirectory(deltagram *Deltagram) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	// Process operations in the order they appear
	for _, part := range deltagram.Parts {
		// Skip message parts
		if part.ContentLocation == "mimeogram://message" || part.ContentLocation == "deltagram://message" {
			fmt.Printf("Message: %s\n", strings.TrimSpace(part.Content))
			continue
		}

		switch part.DeltaOperation {
		case "delete":
			if err := applyDeleteOperation(cwd, part); err != nil {
				return fmt.Errorf("failed to delete %s: %v", part.ContentLocation, err)
			}
		case "move":
			if err := applyMoveOperation(cwd, part); err != nil {
				return fmt.Errorf("failed to move %s: %v", part.ContentLocation, err)
			}
		case "copy":
			if err := applyCopyOperation(cwd, part); err != nil {
				return fmt.Errorf("failed to copy %s: %v", part.ContentLocation, err)
			}
		case "create":
			if err := applyCreateOperation(cwd, part); err != nil {
				return fmt.Errorf("failed to create %s: %v", part.ContentLocation, err)
			}
		case "content":
			if err := applyContentOperation(cwd, part); err != nil {
				return fmt.Errorf("failed to apply content changes to %s: %v", part.ContentLocation, err)
			}
		default:
			// Default to create for backward compatibility
			if err := applyCreateOperation(cwd, part); err != nil {
				return fmt.Errorf("failed to create %s: %v", part.ContentLocation, err)
			}
		}
	}

	return nil
}

func applyDeleteOperation(baseDir string, part DeltagramPart) error {
	filePath := resolveFilePath(baseDir, part.ContentLocation)
	
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Warning: File %s does not exist (already deleted)\n", part.ContentLocation)
			return nil
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	fmt.Printf("Deleted: %s\n", part.ContentLocation)
	return nil
}

func applyMoveOperation(baseDir string, part DeltagramPart) error {
	// Parse move operation content to get source and destination
	lines := strings.Split(part.Content, "\n")
	var sourcePath, destPath string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "---") {
			sourcePath = strings.TrimSpace(strings.TrimPrefix(line, "---"))
		} else if strings.HasPrefix(line, "+++") {
			destPath = strings.TrimSpace(strings.TrimPrefix(line, "+++"))
		}
	}
	
	if sourcePath == "" || destPath == "" {
		return fmt.Errorf("invalid move operation: missing source or destination path")
	}
	
	sourceFullPath := resolveFilePath(baseDir, sourcePath)
	destFullPath := resolveFilePath(baseDir, destPath)
	
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}
	
	if err := os.Rename(sourceFullPath, destFullPath); err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}

	fmt.Printf("Moved: %s -> %s\n", sourcePath, destPath)
	return nil
}

func applyCopyOperation(baseDir string, part DeltagramPart) error {
	// Parse copy operation content to get source and destination
	lines := strings.Split(part.Content, "\n")
	var sourcePath, destPath string
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "---") {
			sourcePath = strings.TrimSpace(strings.TrimPrefix(line, "---"))
		} else if strings.HasPrefix(line, "+++") {
			destPath = strings.TrimSpace(strings.TrimPrefix(line, "+++"))
		}
	}
	
	if sourcePath == "" || destPath == "" {
		return fmt.Errorf("invalid copy operation: missing source or destination path")
	}
	
	sourceFullPath := resolveFilePath(baseDir, sourcePath)
	destFullPath := resolveFilePath(baseDir, destPath)
	
	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destFullPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}
	
	if err := copyFile(sourceFullPath, destFullPath); err != nil {
		return fmt.Errorf("failed to copy file: %v", err)
	}

	fmt.Printf("Copied: %s -> %s\n", sourcePath, destPath)
	return nil
}

func applyCreateOperation(baseDir string, part DeltagramPart) error {
	filePath := resolveFilePath(baseDir, part.ContentLocation)
	
	// Parse create operation content
	lines := strings.Split(part.Content, "\n")
	var content string
	var contentStarted bool
	
	for _, line := range lines {
		if strings.HasPrefix(line, "+++") {
			contentStarted = true
			continue
		}
		if contentStarted {
			if content != "" {
				content += "\n"
			}
			content += line
		}
	}
	
	// If no +++ marker found, use entire content
	if !contentStarted {
		content = part.Content
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}

	// Write file content
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("Created: %s\n", part.ContentLocation)
	return nil
}

func applyContentOperation(baseDir string, part DeltagramPart) error {
	filePath := resolveFilePath(baseDir, part.ContentLocation)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", part.ContentLocation)
	}
	
	// Read existing file
	existingContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read existing file: %v", err)
	}
	
	// Apply unified diff
	modifiedContent, err := applyUnifiedDiff(string(existingContent), part.Content)
	if err != nil {
		return fmt.Errorf("failed to apply diff: %v", err)
	}
	
	// Write modified content back
	if err := os.WriteFile(filePath, []byte(modifiedContent), 0644); err != nil {
		return fmt.Errorf("failed to write modified file: %v", err)
	}

	fmt.Printf("Modified: %s\n", part.ContentLocation)
	return nil
}

func applyUnifiedDiff(original, diff string) (string, error) {
	originalLines := strings.Split(original, "\n")
	diffLines := strings.Split(diff, "\n")
	
	// Parse unified diff hunks
	var result []string
	originalIndex := 0
	
	for i := 0; i < len(diffLines); i++ {
		line := diffLines[i]
		
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header
			hunk, err := parseHunkHeader(line)
			if err != nil {
				return "", fmt.Errorf("invalid hunk header: %v", err)
			}
			
			// Copy lines up to the hunk start
			for originalIndex < hunk.OldStart-1 {
				result = append(result, originalLines[originalIndex])
				originalIndex++
			}
			
			// Process hunk content
			i++ // Skip hunk header
			for i < len(diffLines) && !strings.HasPrefix(diffLines[i], "@@") {
				hunkLine := diffLines[i]
				if len(hunkLine) == 0 {
					i++
					continue
				}
				
				switch hunkLine[0] {
				case '+':
					// Add line
					result = append(result, hunkLine[1:])
				case '-':
					// Remove line (skip it)
					originalIndex++
				case ' ':
					// Context line (unchanged)
					if originalIndex < len(originalLines) {
						result = append(result, originalLines[originalIndex])
						originalIndex++
					}
				}
				i++
			}
			i-- // Adjust for outer loop increment
		}
	}
	
	// Copy remaining lines
	for originalIndex < len(originalLines) {
		result = append(result, originalLines[originalIndex])
		originalIndex++
	}
	
	return strings.Join(result, "\n"), nil
}

type HunkHeader struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
}

func parseHunkHeader(line string) (*HunkHeader, error) {
	// Example: @@ -1,5 +1,8 @@
	re := regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid hunk header format")
	}
	
	oldStart, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid old start: %v", err)
	}
	
	oldCount := 1
	if matches[2] != "" {
		oldCount, err = strconv.Atoi(matches[2])
		if err != nil {
			return nil, fmt.Errorf("invalid old count: %v", err)
		}
	}
	
	newStart, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid new start: %v", err)
	}
	
	newCount := 1
	if len(matches) > 4 && matches[4] != "" {
		newCount, err = strconv.Atoi(matches[4])
		if err != nil {
			return nil, fmt.Errorf("invalid new count: %v", err)
		}
	}
	
	return &HunkHeader{
		OldStart: oldStart,
		OldCount: oldCount,
		NewStart: newStart,
		NewCount: newCount,
	}, nil
}

func resolveFilePath(baseDir, filePath string) string {
	// Handle URLs or absolute paths by converting to relative
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		// Extract filename from URL
		parts := strings.Split(filePath, "/")
		filePath = parts[len(parts)-1]
	} else if filepath.IsAbs(filePath) {
		// Convert absolute path to relative by removing leading slash
		filePath = strings.TrimPrefix(filePath, "/")
	}

	return filepath.Join(baseDir, filePath)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}