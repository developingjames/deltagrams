package operations

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"deltagram/pkg/parser"
)

// ContentHandler handles content modification operations using unified diff
type ContentHandler struct{}

// NewContentHandler creates a new content handler
func NewContentHandler() OperationHandler {
	return &ContentHandler{}
}

// CanHandle returns true if this handler can process the given operation
func (h *ContentHandler) CanHandle(operation string) bool {
	return operation == "content"
}

// Apply applies content modifications using unified diff format
func (h *ContentHandler) Apply(fs FileSystem, baseDir string, part parser.DeltagramPart) error {
	filePath := ResolveFilePath(baseDir, part.ContentLocation)
	
	// Check if file exists
	if _, err := fs.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", part.ContentLocation)
	}
	
	// Read existing file
	existingContent, err := fs.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read existing file: %v", err)
	}
	
	// Apply unified diff
	modifiedContent, err := h.applyUnifiedDiff(string(existingContent), part.Content)
	if err != nil {
		return fmt.Errorf("failed to apply diff: %v", err)
	}
	
	// Write modified content back
	if err := fs.WriteFile(filePath, []byte(modifiedContent), 0644); err != nil {
		return fmt.Errorf("failed to write modified file: %v", err)
	}

	fmt.Printf("Modified: %s\n", part.ContentLocation)
	return nil
}

func (h *ContentHandler) applyUnifiedDiff(original, diff string) (string, error) {
	originalLines := strings.Split(original, "\n")
	diffLines := strings.Split(diff, "\n")
	
	// Parse unified diff hunks
	var result []string
	originalIndex := 0
	
	for i := 0; i < len(diffLines); i++ {
		line := diffLines[i]
		
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header
			hunk, err := h.parseHunkHeader(line)
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

// HunkHeader represents a parsed unified diff hunk header
type HunkHeader struct {
	OldStart int
	OldCount int
	NewStart int
	NewCount int
}

func (h *ContentHandler) parseHunkHeader(line string) (*HunkHeader, error) {
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