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


func TestParser_Parse_FlexibleIdentifiers(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		shouldPass bool
	}{
		{
			name:       "traditional UUID",
			identifier: "0123456789abcdef0123456789abcdef",
			shouldPass: true,
		},
		{
			name:       "mixed case alphanumeric",
			identifier: "voice456sample789012345678901234ef",
			shouldPass: true,
		},
		{
			name:       "uppercase letters",
			identifier: "0123456789ABCDEF0123456789abcdef",
			shouldPass: true,
		},
		{
			name:       "shorter valid identifier",
			identifier: "test1234",
			shouldPass: true,
		},
		{
			name:       "too short identifier",
			identifier: "test123",
			shouldPass: false,
		},
		{
			name:       "with underscores",
			identifier: "test_123_456",
			shouldPass: true,
		},
		{
			name:       "with dashes",
			identifier: "test-123-456",
			shouldPass: true,
		},
		{
			name:       "mixed with underscores and dashes",
			identifier: "test_123-456_789",
			shouldPass: true,
		},
		{
			name:       "invalid characters (space)",
			identifier: "test 123 456",
			shouldPass: false,
		},
		{
			name:       "invalid characters (special chars)",
			identifier: "test@123#456",
			shouldPass: false,
		},
	}

	parser := NewParser()
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := `--====DELTAGRAM_` + test.identifier + `====
Content-Location: test/file.txt
Content-Type: text/plain; charset=utf-8; linesep=LF

Hello, World!
--====DELTAGRAM_` + test.identifier + `====--`

			deltagram, err := parser.Parse(content)
			if test.shouldPass {
				if err != nil {
					t.Errorf("Expected no error for valid identifier '%s', got: %v", test.identifier, err)
				}
				if deltagram.UUID != test.identifier {
					t.Errorf("Expected identifier '%s', got: %s", test.identifier, deltagram.UUID)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid identifier '%s', got none", test.identifier)
				}
				if !strings.Contains(err.Error(), "invalid boundary identifier format") && !strings.Contains(err.Error(), "missing or malformed boundary") {
					t.Errorf("Expected identifier format or boundary error, got: %v", err)
				}
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