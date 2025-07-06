package operations

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/developingjames/deltagrams/pkg/parser"
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
		return fmt.Errorf("cannot apply content operation to non-existent file: %s (use 'create' operation instead)", part.ContentLocation)
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
	
	// Parse all hunks first
	hunks, err := h.ParseAllHunks(diffLines)
	if err != nil {
		return "", err
	}
	
	// Apply hunks in forward order using standard patch algorithm
	result := make([]string, len(originalLines))
	copy(result, originalLines)
	offset := 0 // Track cumulative line offset from previous hunks
	
	for _, hunk := range hunks {
		// Calculate the actual position accounting for previous changes
		actualStart := hunk.Header.OldStart - 1 + offset
		if actualStart < 0 {
			actualStart = 0
		}
		if actualStart > len(result) {
			return "", fmt.Errorf("hunk refers to line %d but file has %d lines", hunk.Header.OldStart, len(result))
		}
		
		// Validate context and apply hunk using standard algorithm
		newResult, hunkOffset, err := h.applyHunkStandard(result, hunk, actualStart)
		if err != nil {
			return "", fmt.Errorf("failed to apply hunk at line %d: %v", hunk.Header.OldStart, err)
		}
		
		// Update offset for next hunk
		offset += hunkOffset
		result = newResult
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

// HunkOperation represents a single operation within a hunk
type HunkOperation struct {
	Type    byte   // '+', '-', or ' '
	Content string
}

// ParsedHunk represents a complete hunk with its operations
type ParsedHunk struct {
	Header     *HunkHeader
	Operations []HunkOperation
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

// ParseAllHunks parses all hunks from the diff lines
func (h *ContentHandler) ParseAllHunks(diffLines []string) ([]*ParsedHunk, error) {
	var hunks []*ParsedHunk
	
	for i := 0; i < len(diffLines); i++ {
		line := diffLines[i]
		
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header
			header, err := h.parseHunkHeader(line)
			if err != nil {
				return nil, fmt.Errorf("invalid hunk header: %v", err)
			}
			
			// Parse hunk operations
			var operations []HunkOperation
			i++ // Skip hunk header
			for i < len(diffLines) && !strings.HasPrefix(diffLines[i], "@@") {
				hunkLine := diffLines[i]
				if len(hunkLine) == 0 {
					i++
					continue
				}
				
				if hunkLine[0] == '+' || hunkLine[0] == '-' || hunkLine[0] == ' ' {
					operations = append(operations, HunkOperation{
						Type:    hunkLine[0],
						Content: hunkLine[1:],
					})
				}
				i++
			}
			i-- // Adjust for outer loop increment
			
			hunks = append(hunks, &ParsedHunk{
				Header:     header,
				Operations: operations,
			})
		}
	}
	
	return hunks, nil
}

// applyHunkStandard applies a single hunk using the standard patch algorithm
// Returns the new result, offset change, and any error
func (h *ContentHandler) applyHunkStandard(result []string, hunk *ParsedHunk, actualStart int) ([]string, int, error) {
	// First, validate that context lines match the current file content
	if err := h.validateHunkContext(result, hunk, actualStart); err != nil {
		return nil, 0, err
	}
	
	// For pure insertions (OldCount=0), handle specially
	if hunk.Header.OldCount == 0 {
		return h.applyPureInsertion(result, hunk, actualStart)
	}
	
	// For replacements, extract only the replacement content (no context)
	replacementLines := h.extractReplacementContent(hunk)
	
	// Replace exactly OldCount lines with the replacement content
	endPos := actualStart + hunk.Header.OldCount
	if endPos > len(result) {
		endPos = len(result)
	}
	
	// Build new result
	newResult := make([]string, 0, len(result)+len(replacementLines)-(endPos-actualStart))
	newResult = append(newResult, result[:actualStart]...)
	newResult = append(newResult, replacementLines...)
	newResult = append(newResult, result[endPos:]...)
	
	// Calculate offset change
	offset := len(replacementLines) - hunk.Header.OldCount
	
	return newResult, offset, nil
}

// validateHunkContext validates that context lines match the current file content
// Normalizes line endings to handle CRLF vs LF differences
func (h *ContentHandler) validateHunkContext(result []string, hunk *ParsedHunk, actualStart int) error {
	filePos := actualStart
	
	for _, op := range hunk.Operations {
		switch op.Type {
		case ' ':
			// Context line - must match file content (ignoring line ending differences)
			if filePos >= len(result) {
				return fmt.Errorf("context line extends beyond file end")
			}
			if !h.LinesEqual(result[filePos], op.Content) {
				return fmt.Errorf("context mismatch at line %d: expected %q, got %q", filePos+1, op.Content, result[filePos])
			}
			filePos++
		case '-':
			// Line to be removed - must match file content (ignoring line ending differences)
			if filePos >= len(result) {
				return fmt.Errorf("line to remove extends beyond file end")
			}
			if !h.LinesEqual(result[filePos], op.Content) {
				return fmt.Errorf("line to remove mismatch at line %d: expected %q, got %q", filePos+1, op.Content, result[filePos])
			}
			filePos++
		case '+':
			// Line to be added - doesn't affect validation position
			continue
		}
	}
	
	return nil
}

// LinesEqual compares two lines ignoring line ending differences
func (h *ContentHandler) LinesEqual(line1, line2 string) bool {
	// Normalize line endings by removing all CR characters
	normalized1 := strings.ReplaceAll(line1, "\r", "")
	normalized2 := strings.ReplaceAll(line2, "\r", "")
	return normalized1 == normalized2
}

// extractReplacementContent extracts the final content that should replace the old lines
// This includes context lines and additions, but only for the lines being replaced (OldCount)
func (h *ContentHandler) extractReplacementContent(hunk *ParsedHunk) []string {
	replacementLines := make([]string, 0)
	oldLinesSeen := 0 // Track how many original lines we've processed
	
	for i, op := range hunk.Operations {
		switch op.Type {
		case ' ':
			// Context line - include in replacement only if within OldCount
			if oldLinesSeen < hunk.Header.OldCount {
				replacementLines = append(replacementLines, op.Content)
			}
			oldLinesSeen++
		case '-':
			// Removed line - do NOT include in final result, but count towards OldCount
			oldLinesSeen++
		case '+':
			// Added line - include if we haven't exceeded OldCount yet
			// OR if there are still more old lines to process after this addition
			if oldLinesSeen < hunk.Header.OldCount || h.hasMoreOldLines(hunk.Operations[i+1:], hunk.Header.OldCount-oldLinesSeen) {
				replacementLines = append(replacementLines, op.Content)
			}
		}
		
		// Stop processing once we've seen all the old lines that are being replaced
		if oldLinesSeen >= hunk.Header.OldCount {
			break
		}
	}
	
	return replacementLines
}

// hasMoreOldLines checks if there are more old lines (context or removals) in the remaining operations
func (h *ContentHandler) hasMoreOldLines(remainingOps []HunkOperation, neededCount int) bool {
	count := 0
	for _, op := range remainingOps {
		if op.Type == ' ' || op.Type == '-' {
			count++
			if count >= neededCount {
				return true
			}
		}
	}
	return false
}

// applyPureInsertion handles hunks with OldCount=0 (pure insertions)
func (h *ContentHandler) applyPureInsertion(result []string, hunk *ParsedHunk, actualStart int) ([]string, int, error) {
	// Extract only the lines to be inserted (ignore context)
	insertionLines := make([]string, 0)
	for _, op := range hunk.Operations {
		if op.Type == '+' {
			insertionLines = append(insertionLines, op.Content)
		}
	}
	
	// Insert at the specified position
	newResult := make([]string, 0, len(result)+len(insertionLines))
	newResult = append(newResult, result[:actualStart]...)
	newResult = append(newResult, insertionLines...)
	newResult = append(newResult, result[actualStart:]...)
	
	return newResult, len(insertionLines), nil
}