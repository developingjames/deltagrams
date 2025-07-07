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
	
	// Apply hunks sequentially with automatic offset calculation
	// Each hunk references original file line numbers, but we apply to evolving result
	result := make([]string, len(originalLines))
	copy(result, originalLines)
	
	// Track mapping from original line numbers to current result line numbers
	// lineMapping[originalLineIndex] = currentResultLineIndex
	lineMapping := make([]int, len(originalLines))
	for i := range lineMapping {
		lineMapping[i] = i
	}
	
	for _, hunk := range hunks {
		// Hunk references original file line numbers
		originalStart := hunk.Header.OldStart - 1 // Convert to 0-based indexing
		if originalStart < 0 || originalStart >= len(originalLines) {
			return "", fmt.Errorf("hunk refers to line %d but original file has %d lines", hunk.Header.OldStart, len(originalLines))
		}
		
		// Find the best position for this hunk in the original file (with fuzzy matching)
		bestPosition, err := h.findBestHunkPosition(originalLines, hunk, originalStart)
		if err != nil {
			return "", fmt.Errorf("failed to find position for hunk at line %d: %v", hunk.Header.OldStart, err)
		}
		
		// Update originalStart to the best position found
		originalStart = bestPosition
		
		// Find where this original line is now located in the current result
		currentStart := lineMapping[originalStart]
		
		// Apply the hunk at the current position
		newResult, netLineChange, err := h.applyHunkAtPosition(result, hunk, currentStart)
		if err != nil {
			return "", fmt.Errorf("failed to apply hunk at line %d: %v", hunk.Header.OldStart, err)
		}
		
		// Update line mapping for all lines after the affected region
		h.updateLineMapping(lineMapping, originalStart, hunk.Header.OldCount, netLineChange)
		
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

// validateHunkAgainstOriginal validates that hunk context matches the original file
func (h *ContentHandler) validateHunkAgainstOriginal(originalLines []string, hunk *ParsedHunk, originalStart int) error {
	originalPos := originalStart
	
	for _, op := range hunk.Operations {
		switch op.Type {
		case ' ':
			// Context line - must match original file content
			if originalPos >= len(originalLines) {
				return fmt.Errorf("context line extends beyond original file")
			}
			if !h.LinesEqual(originalLines[originalPos], op.Content) {
				return fmt.Errorf("context mismatch at original line %d: expected %q, got %q", 
					originalPos+1, op.Content, originalLines[originalPos])
			}
			originalPos++
		case '-':
			// Line to be removed - must match original file content
			if originalPos >= len(originalLines) {
				return fmt.Errorf("line to remove extends beyond original file")
			}
			if !h.LinesEqual(originalLines[originalPos], op.Content) {
				return fmt.Errorf("removal mismatch at original line %d: expected %q, got %q", 
					originalPos+1, op.Content, originalLines[originalPos])
			}
			originalPos++
		case '+':
			// Line to be added - doesn't advance original position
			continue
		}
	}
	
	return nil
}

// applyHunkAtPosition applies a hunk at the specified current position
func (h *ContentHandler) applyHunkAtPosition(result []string, hunk *ParsedHunk, currentStart int) ([]string, int, error) {
	// Handle pure insertions (OldCount=0) specially
	if hunk.Header.OldCount == 0 {
		insertLines := make([]string, 0)
		for _, op := range hunk.Operations {
			if op.Type == '+' {
				insertLines = append(insertLines, op.Content)
			}
		}
		
		newResult := make([]string, 0, len(result)+len(insertLines))
		newResult = append(newResult, result[:currentStart]...)
		newResult = append(newResult, insertLines...)
		newResult = append(newResult, result[currentStart:]...)
		
		return newResult, len(insertLines), nil
	}
	
	// For replacements, build the new content 
	// Only include lines that should be in the replacement (within OldCount range)
	replacementLines := make([]string, 0)
	oldLinesProcessed := 0
	
	for _, op := range hunk.Operations {
		switch op.Type {
		case ' ':
			// Context line - only include if within OldCount range
			if oldLinesProcessed < hunk.Header.OldCount {
				replacementLines = append(replacementLines, op.Content)
			}
			oldLinesProcessed++
		case '+':
			// Added line - include if we haven't exceeded OldCount yet
			if oldLinesProcessed < hunk.Header.OldCount {
				replacementLines = append(replacementLines, op.Content)
			} else {
				// This is an addition after the replacement range - include it
				replacementLines = append(replacementLines, op.Content)
			}
		case '-':
			// Removed line - do NOT include in result but count toward OldCount
			oldLinesProcessed++
		}
	}
	
	// Replace exactly OldCount lines with the replacement content
	endPos := currentStart + hunk.Header.OldCount
	if endPos > len(result) {
		endPos = len(result)
	}
	
	newResult := make([]string, 0, len(result)+len(replacementLines)-(endPos-currentStart))
	newResult = append(newResult, result[:currentStart]...)
	newResult = append(newResult, replacementLines...)
	newResult = append(newResult, result[endPos:]...)
	
	// Calculate net line change
	netChange := len(replacementLines) - hunk.Header.OldCount
	
	return newResult, netChange, nil
}

// findBestHunkPosition finds the best position for a hunk with fuzzy matching
func (h *ContentHandler) findBestHunkPosition(originalLines []string, hunk *ParsedHunk, suggestedStart int) (int, error) {
	// Try the suggested position first (exact match)
	if h.validateHunkAgainstOriginal(originalLines, hunk, suggestedStart) == nil {
		return suggestedStart, nil
	}
	
	// If exact match fails, try positions within a reasonable range
	searchRange := 5 // Search +/- 5 lines around the suggested position
	
	// Try positions before the suggested start
	for offset := 1; offset <= searchRange; offset++ {
		// Try position before
		if suggestedStart-offset >= 0 {
			if h.validateHunkAgainstOriginal(originalLines, hunk, suggestedStart-offset) == nil {
				return suggestedStart - offset, nil
			}
		}
		
		// Try position after
		if suggestedStart+offset < len(originalLines) {
			if h.validateHunkAgainstOriginal(originalLines, hunk, suggestedStart+offset) == nil {
				return suggestedStart + offset, nil
			}
		}
	}
	
	// If no fuzzy match found, return the original error
	return suggestedStart, h.validateHunkAgainstOriginal(originalLines, hunk, suggestedStart)
}

// updateLineMapping updates the mapping after a hunk is applied
func (h *ContentHandler) updateLineMapping(lineMapping []int, originalStart, oldCount, netChange int) {
	// Update mapping for all original lines after the affected region
	for i := originalStart + oldCount; i < len(lineMapping); i++ {
		lineMapping[i] += netChange
	}
}