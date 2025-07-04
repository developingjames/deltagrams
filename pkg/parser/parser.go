package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// DefaultParser implements the Parser interface
type DefaultParser struct{}

// NewParser creates a new default parser
func NewParser() Parser {
	return &DefaultParser{}
}

// Parse parses a deltagram string into a Deltagram struct
func (p *DefaultParser) Parse(content string) (*Deltagram, error) {
	// Normalize line endings to LF
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	// Extract UUID from first boundary
	boundaryRegex := regexp.MustCompile(`--====(?:DELTAGRAM|MIMEOGRAM)_([a-f0-9]{32})====`)
	matches := boundaryRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid deltagram format: missing or malformed boundary")
	}
	
	uuid := matches[1]
	
	// Validate UUID format (32 lowercase hex characters)
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(uuid) {
		return nil, fmt.Errorf("invalid UUID format: %s", uuid)
	}

	// Split by boundary markers (support both DELTAGRAM and MIMEOGRAM formats)
	var boundaryPattern string
	if strings.Contains(content, "DELTAGRAM_") {
		boundaryPattern = fmt.Sprintf(`--====DELTAGRAM_%s====`, uuid)
	} else {
		boundaryPattern = fmt.Sprintf(`--====MIMEOGRAM_%s====`, uuid)
	}
	parts := strings.Split(content, boundaryPattern)
	
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid deltagram format: no parts found")
	}

	// Remove empty first part (before first boundary)
	if strings.TrimSpace(parts[0]) == "" {
		parts = parts[1:]
	}

	deltagram := &Deltagram{
		UUID:  uuid,
		Parts: make([]DeltagramPart, 0),
	}

	for i, part := range parts {
		// Check if this is the final boundary (ends with --)
		if strings.HasSuffix(strings.TrimSpace(part), "--") {
			// Remove the trailing -- and process if there's content
			part = strings.TrimSuffix(strings.TrimSpace(part), "--")
			if strings.TrimSpace(part) == "" {
				break // Final boundary with no content
			}
		}

		parsedPart, err := p.parsePart(part)
		if err != nil {
			return nil, fmt.Errorf("error parsing part %d: %v", i+1, err)
		}
		
		deltagram.Parts = append(deltagram.Parts, *parsedPart)
	}

	return deltagram, nil
}

func (p *DefaultParser) parsePart(partContent string) (*DeltagramPart, error) {
	// Trim leading/trailing whitespace
	partContent = strings.TrimSpace(partContent)
	lines := strings.Split(partContent, "\n")
	
	var contentLocation, contentType, deltaOperation string
	var contentStartIndex int
	
	// Parse headers
	for i, line := range lines {
		line = strings.TrimSpace(line)
		
		if line == "" {
			contentStartIndex = i + 1
			break
		}
		
		if strings.HasPrefix(line, "Content-Location:") {
			contentLocation = strings.TrimSpace(strings.TrimPrefix(line, "Content-Location:"))
		} else if strings.HasPrefix(line, "Content-Type:") {
			contentType = strings.TrimSpace(strings.TrimPrefix(line, "Content-Type:"))
		} else if strings.HasPrefix(line, "Delta-Operation:") {
			deltaOperation = strings.TrimSpace(strings.TrimPrefix(line, "Delta-Operation:"))
		}
	}
	
	if contentLocation == "" {
		return nil, fmt.Errorf("missing Content-Location header")
	}
	
	if contentType == "" {
		return nil, fmt.Errorf("missing Content-Type header")
	}
	
	// For message parts, Delta-Operation is optional
	isMessage := contentLocation == "mimeogram://message" || contentLocation == "deltagram://message"
	if !isMessage && deltaOperation == "" {
		// Default to CREATE for backward compatibility with mimeogram format
		deltaOperation = "create"
	}
	
	// Extract content
	var content string
	if contentStartIndex < len(lines) {
		content = strings.Join(lines[contentStartIndex:], "\n")
	}
	
	return &DeltagramPart{
		ContentLocation: contentLocation,
		ContentType:     contentType,
		DeltaOperation:  deltaOperation,
		Content:         content,
	}, nil
}