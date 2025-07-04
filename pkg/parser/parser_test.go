package parser

import (
	"strings"
	"testing"
)

func TestParser_Parse_ValidDeltagram(t *testing.T) {
	parser := NewParser()
	
	content := `--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: deltagram://message
Content-Type: text/plain; charset=utf-8; linesep=LF

Test message
--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: test/file.txt
Content-Type: application/x-deltagram-fileop; charset=utf-8
Delta-Operation: create

+++ test/file.txt
Hello, World!
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--`

	deltagram, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if deltagram.UUID != "0123456789abcdef0123456789abcdef" {
		t.Errorf("Expected UUID '0123456789abcdef0123456789abcdef', got: %s", deltagram.UUID)
	}

	if len(deltagram.Parts) != 2 {
		t.Errorf("Expected 2 parts, got: %d", len(deltagram.Parts))
	}

	// Check message part
	messagePart := deltagram.Parts[0]
	if messagePart.ContentLocation != "deltagram://message" {
		t.Errorf("Expected message content location, got: %s", messagePart.ContentLocation)
	}
	if !strings.Contains(messagePart.Content, "Test message") {
		t.Errorf("Expected message content, got: %s", messagePart.Content)
	}

	// Check file part
	filePart := deltagram.Parts[1]
	if filePart.ContentLocation != "test/file.txt" {
		t.Errorf("Expected file content location, got: %s", filePart.ContentLocation)
	}
	if filePart.DeltaOperation != "create" {
		t.Errorf("Expected create operation, got: %s", filePart.DeltaOperation)
	}
}

func TestParser_Parse_ValidMimeogram(t *testing.T) {
	parser := NewParser()
	
	content := `--====MIMEOGRAM_0123456789abcdef0123456789abcdef====
Content-Location: test/file.txt
Content-Type: text/plain; charset=utf-8; linesep=LF

Hello, World!
--====MIMEOGRAM_0123456789abcdef0123456789abcdef====--`

	deltagram, err := parser.Parse(content)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(deltagram.Parts) != 1 {
		t.Errorf("Expected 1 part, got: %d", len(deltagram.Parts))
	}

	part := deltagram.Parts[0]
	if part.DeltaOperation != "create" {
		t.Errorf("Expected default create operation, got: %s", part.DeltaOperation)
	}
}

func TestParser_Parse_InvalidUUID(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name: "uppercase letters in UUID",
			content: `--====DELTAGRAM_0123456789ABCDEF0123456789abcdef====
Content-Location: test/file.txt
Content-Type: text/plain; charset=utf-8; linesep=LF

Hello, World!
--====DELTAGRAM_0123456789ABCDEF0123456789abcdef====--`,
		},
		{
			name: "wrong length UUID",
			content: `--====DELTAGRAM_0123456789abcdef01234567====
Content-Location: test/file.txt
Content-Type: text/plain; charset=utf-8; linesep=LF

Hello, World!
--====DELTAGRAM_0123456789abcdef01234567====--`,
		},
	}

	parser := NewParser()
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parser.Parse(test.content)
			if err == nil {
				t.Error("Expected error for invalid UUID, got none")
			}
			if !strings.Contains(err.Error(), "invalid UUID format") && !strings.Contains(err.Error(), "missing or malformed boundary") {
				t.Errorf("Expected UUID format or boundary error, got: %v", err)
			}
		})
	}
}

func TestParser_Parse_MissingHeaders(t *testing.T) {
	parser := NewParser()
	
	content := `--====DELTAGRAM_0123456789abcdef0123456789abcdef====
Content-Location: test/file.txt

Hello, World!
--====DELTAGRAM_0123456789abcdef0123456789abcdef====--`

	_, err := parser.Parse(content)
	if err == nil {
		t.Error("Expected error for missing Content-Type header, got none")
	}
	if !strings.Contains(err.Error(), "missing Content-Type header") {
		t.Errorf("Expected Content-Type header error, got: %v", err)
	}
}

func TestParser_Parse_NoBoundary(t *testing.T) {
	parser := NewParser()
	
	content := `Content-Location: test/file.txt
Content-Type: text/plain; charset=utf-8; linesep=LF

Hello, World!`

	_, err := parser.Parse(content)
	if err == nil {
		t.Error("Expected error for missing boundary, got none")
	}
	if !strings.Contains(err.Error(), "missing or malformed boundary") {
		t.Errorf("Expected boundary error, got: %v", err)
	}
}