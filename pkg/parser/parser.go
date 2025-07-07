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

	// Extract boundary identifier from first boundary (more flexible than strict UUID)
	boundaryRegex := regexp.MustCompile(`--====DELTAGRAM_([a-zA-Z0-9_-]+)====`)
	matches := boundaryRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid deltagram format: missing or malformed boundary")
	}

	identifier := matches[1]

	// Validate identifier format (alphanumeric, underscore, dash, at least 8 characters for reasonable uniqueness)
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]{8,}$`).MatchString(identifier) {
		return nil, fmt.Errorf("invalid boundary identifier format: %s (must be at least 8 characters using alphanumeric, underscore, or dash)", identifier)
	}

	// Split by boundary markers
	boundaryPattern := fmt.Sprintf(`--====DELTAGRAM_%s====`, identifier)
	parts := strings.Split(content, boundaryPattern)

	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid deltagram format: no parts found")
	}

	// Remove empty first part (before first boundary)
	if strings.TrimSpace(parts[0]) == "" {
		parts = parts[1:]
	}

	deltagram := &Deltagram{
		UUID:  identifier,
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
	isMessage := contentLocation == "deltagram://message"
	if !isMessage && deltaOperation == "" {
		// Default to CREATE for backward compatibility
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
